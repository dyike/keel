package sys

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"image/color"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/dyike/keel/native"
	"golang.org/x/image/bmp"
)

func unicodeClipboard(s string) []byte {
	units := append(utf16.Encode([]rune(s)), 0)
	out := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(out[i*2:], u)
	}
	return out
}
func testDIB(bits uint16, height int32) []byte {
	stride := (2*int(bits) + 31) / 32 * 4
	h := int(height)
	if h < 0 {
		h = -h
	}
	colors := 0
	if bits <= 8 {
		colors = 1 << bits
	}
	data := make([]byte, 40+colors*4+stride*h)
	binary.LittleEndian.PutUint32(data, 40)
	binary.LittleEndian.PutUint32(data[4:], 2)
	binary.LittleEndian.PutUint32(data[8:], uint32(height))
	binary.LittleEndian.PutUint16(data[12:], 1)
	binary.LittleEndian.PutUint16(data[14:], bits)
	return data
}

func TestClipboardUnicodeAndSnapshot(t *testing.T) {
	for _, text := range []string{"", "中文 😀\r\nnext", "a\tb"} {
		got, err := clipboardUTF16(unicodeClipboard(text))
		if err != nil || got != text {
			t.Fatal(got, err)
		}
	}
	for _, raw := range [][]byte{{1}, {1, 0}, make([]byte, clipboardByteLimit+2)} {
		if _, err := clipboardUTF16(raw); !errors.Is(err, native.ErrFailed) {
			t.Fatal("invalid text accepted", len(raw), err)
		}
	}
	formats := map[string][]byte{"text": unicodeClipboard("hello"), "png": {137, 80, 78, 71}, "dib": testDIB(24, 1)}
	var requested []string
	read := func(name string) ([]byte, error) { requested = append(requested, name); return formats[name], nil }
	collect := func(paths []string) clipboardSnapshot {
		t.Helper()
		raw, err := collectWindowsClipboard(read, func() ([]string, error) { return paths, nil })
		if err != nil {
			t.Fatal(err)
		}
		var out clipboardSnapshot
		if err = json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	out := collect(nil)
	if out.Text != "hello" || len(out.Images) != 1 || out.Images[0].MIME != "image/png" || strings.Join(requested, ",") != "text,png" {
		t.Fatal("PNG priority", out, requested)
	}
	requested = nil
	out = collect([]string{`C:\文件\hello.png`})
	if len(out.Images) != 0 || len(out.Files) != 1 || strings.Join(requested, ",") != "text" {
		t.Fatal("file preview duplicated", out, requested)
	}
	delete(formats, "png")
	out = collect(nil)
	if len(out.Images) != 1 || out.Images[0].MIME != "image/bmp" || string(out.Images[0].Data[:2]) != "BM" {
		t.Fatal("DIB fallback", out)
	}
	formats = nil
	out = collect(nil)
	if out.Text != "" || len(out.Files) != 0 || len(out.Images) != 0 {
		t.Fatal("empty clipboard", out)
	}
}

func TestClipboardSnapshotFailureIsAtomic(t *testing.T) {
	for _, tc := range []struct {
		name  string
		read  func(string) ([]byte, error)
		files func() ([]string, error)
	}{
		{"text error", func(string) ([]byte, error) { return nil, native.ErrFailed }, func() ([]string, error) { return nil, nil }},
		{"files error", func(string) ([]byte, error) { return unicodeClipboard("text"), nil }, func() ([]string, error) { return nil, native.ErrFailed }},
		{"too many files", func(string) ([]byte, error) { return nil, nil }, func() ([]string, error) { return make([]string, 129), nil }},
		{"aggregate", func(name string) ([]byte, error) {
			if name == "text" {
				return unicodeClipboard("a"), nil
			}
			return make([]byte, clipboardByteLimit), nil
		}, func() ([]string, error) { return nil, nil }},
		{"bad DIB", func(name string) ([]byte, error) {
			if name == "dib" {
				return []byte{1}, nil
			}
			return nil, nil
		}, func() ([]string, error) { return nil, nil }},
	} {
		raw, err := collectWindowsClipboard(tc.read, tc.files)
		if raw != nil || !errors.Is(err, native.ErrFailed) {
			t.Fatal(tc.name, "returned partial success", len(raw), err)
		}
	}
}

func TestClipboardDIBPixelsAndBounds(t *testing.T) {
	for _, height := range []int32{2, -2} {
		dib := testDIB(24, height)
		copy(dib[40:], []byte{0, 0, 255, 255, 0, 0, 0, 0, 0, 255, 0, 255, 255, 255})
		encoded, err := clipboardDIB(dib)
		if err != nil {
			t.Fatal(err)
		}
		img, err := bmp.Decode(bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		y := 0
		if height > 0 {
			y = 1
		}
		if img.At(0, y) != (color.RGBA{R: 255, A: 255}) || img.At(1, y) != (color.RGBA{B: 255, A: 255}) || img.At(0, 1-y) != (color.RGBA{G: 255, A: 255}) {
			t.Fatal("BGR/orientation", img.At(0, 0), img.At(1, 0))
		}
	}
	pal := testDIB(8, 1)
	pal[42] = 255
	encoded, err := clipboardDIB(pal)
	if err != nil {
		t.Fatal(err)
	}
	if binary.LittleEndian.Uint32(encoded[10:]) != 14+40+256*4 {
		t.Fatal("palette offset")
	}
	if _, err = bmp.Decode(bytes.NewReader(encoded)); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]byte) []byte{
		func(b []byte) []byte { return b[:20] },
		func(b []byte) []byte { return b[:len(b)-1] },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b, 200); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[4:], 0x7fffffff); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[8:], 0x80000000); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[32:], 0xffffffff); return b },
		func(b []byte) []byte { binary.LittleEndian.PutUint32(b[16:], 1); return b },
	} {
		if _, err := clipboardDIB(mutate(testDIB(24, 1))); !errors.Is(err, native.ErrFailed) {
			t.Fatal("invalid bitmap accepted", err)
		}
	}
}

func FuzzClipboardDIB(f *testing.F) {
	f.Add(testDIB(24, 1))
	f.Add(testDIB(32, -1))
	f.Add(testDIB(8, 1))
	f.Fuzz(func(t *testing.T, data []byte) {
		out, err := clipboardDIB(data)
		if err == nil {
			if len(out) > clipboardByteLimit || len(out) != len(data)+14 || !bytes.Equal(out[14:], data) {
				t.Fatal("invalid output size/content")
			}
			offset := binary.LittleEndian.Uint32(out[10:])
			if offset >= uint32(len(out)) {
				t.Fatal("pixel offset outside data")
			}
		}
	})
}

func TestClipboardDIBBitfieldsAndExtendedHeaders(t *testing.T) {
	for _, header := range []uint32{40, 108, 124} {
		dib := testDIB(32, 1)
		extra := int(header) - 40
		if header == 40 {
			extra = 12
		}
		dib = append(append(append([]byte(nil), dib[:40]...), make([]byte, extra)...), dib[40:]...)
		binary.LittleEndian.PutUint32(dib, header)
		binary.LittleEndian.PutUint32(dib[16:], 3)
		binary.LittleEndian.PutUint32(dib[40:], 0xff0000)
		binary.LittleEndian.PutUint32(dib[44:], 0xff00)
		binary.LittleEndian.PutUint32(dib[48:], 0xff)
		encoded, err := clipboardDIB(dib)
		if err != nil || binary.LittleEndian.Uint32(encoded[10:]) != uint32(14+40+extra) {
			t.Fatal("bitfield/header offset", header, err)
		}
		if header == 124 {
			binary.LittleEndian.PutUint32(dib[112:], 124)
			if _, err = clipboardDIB(dib); err == nil {
				t.Fatal("unhandled V5 profile accepted")
			}
		}
	}
}
