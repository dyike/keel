package sys

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"unicode/utf16"

	"github.com/dyike/keel/native"
)

const clipboardByteLimit = 16 << 20
const clipboardItemLimit = 128

type clipboardImage struct {
	MIME string `json:"mime"`
	Data []byte `json:"data"`
}
type clipboardSnapshot struct {
	Text   string           `json:"text"`
	Images []clipboardImage `json:"images"`
	Files  []string         `json:"files"`
}

// collectWindowsClipboard operates while the caller holds OpenClipboard. The
// supplied readers return owned copies; nil bytes mean an unavailable format.
func collectWindowsClipboard(read func(string) ([]byte, error), files func() ([]string, error)) ([]byte, error) {
	var out clipboardSnapshot
	data, err := read("text")
	if err != nil {
		return nil, err
	}
	if data != nil {
		out.Text, err = clipboardUTF16(data)
		if err != nil {
			return nil, err
		}
	}
	out.Files, err = files()
	if err != nil {
		return nil, err
	}
	if len(out.Files) > clipboardItemLimit {
		return nil, clipboardFormatError("too many files")
	}
	size := len(out.Text)
	for _, path := range out.Files {
		size += len(path)
	}
	if size > clipboardByteLimit {
		return nil, clipboardFormatError("snapshot exceeds 16MiB")
	}
	// File references take priority over Explorer's image preview.
	if len(out.Files) == 0 {
		data, err = read("png")
		if err != nil {
			return nil, err
		}
		mime := "image/png"
		if data == nil {
			data, err = read("dib")
			if err != nil {
				return nil, err
			}
			if data != nil {
				data, err = clipboardDIB(data)
				if err != nil {
					return nil, err
				}
			}
			mime = "image/bmp"
		}
		if data != nil {
			out.Images = []clipboardImage{{MIME: mime, Data: data}}
			size += len(data)
		}
	}
	if size > clipboardByteLimit {
		return nil, clipboardFormatError("snapshot exceeds 16MiB")
	}
	return json.Marshal(out)
}

func clipboardFormatError(reason string) error {
	return fmt.Errorf("%w: clipboard: %s", native.ErrFailed, reason)
}

func clipboardUTF16(data []byte) (string, error) {
	if len(data) > clipboardByteLimit || len(data)%2 != 0 {
		return "", clipboardFormatError("invalid Unicode text size")
	}
	units := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u := binary.LittleEndian.Uint16(data[i:])
		if u == 0 {
			return string(utf16.Decode(units)), nil
		}
		units = append(units, u)
	}
	return "", clipboardFormatError("unterminated Unicode text")
}

// A packed CF_DIB has no BMP file header. Add that header without decoding
// pixels. Support standard RGB/palette/bitfield layouts; retain encoded bytes.
func clipboardDIB(data []byte) ([]byte, error) {
	invalid := func() ([]byte, error) { return nil, clipboardFormatError("invalid or unsupported DIB") }
	if len(data) < 40 || len(data) > clipboardByteLimit-14 {
		return invalid()
	}
	u32 := func(at int) uint32 { return binary.LittleEndian.Uint32(data[at:]) }
	header := u32(0)
	if header != 40 && header != 108 && header != 124 || uint64(header) > uint64(len(data)) {
		return invalid()
	}
	width, height := int32(u32(4)), int32(u32(8))
	planes, bits := binary.LittleEndian.Uint16(data[12:]), binary.LittleEndian.Uint16(data[14:])
	compression, colors := u32(16), u32(32)
	if width <= 0 || height == 0 || height == -1<<31 || planes != 1 {
		return invalid()
	}
	if bits != 1 && bits != 4 && bits != 8 && bits != 16 && bits != 24 && bits != 32 {
		return invalid()
	}
	if compression != 0 && compression != 3 && compression != 6 {
		return invalid()
	}
	if compression != 0 && bits != 16 && bits != 32 {
		return invalid()
	}
	if bits <= 8 {
		if colors == 0 {
			colors = 1 << bits
		}
		if colors > 1<<bits {
			return invalid()
		}
	}
	offset := uint64(header) + uint64(colors)*4
	if header == 40 {
		if compression == 3 {
			offset += 12
		}
		if compression == 6 {
			offset += 16
		}
	}
	// Embedded/linked V5 color profiles have an additional packed layout; do
	// not guess their position. Windows normally supplies converted CF_DIB.
	if header == 124 && (u32(112) != 0 || u32(116) != 0) {
		return invalid()
	}
	h := int64(height)
	if h < 0 {
		h = -h
	}
	stride := (uint64(width)*uint64(bits) + 31) / 32 * 4
	if offset > uint64(len(data)) || stride > (uint64(len(data))-offset)/uint64(h) {
		return invalid()
	}
	out := make([]byte, 14+len(data))
	copy(out, "BM")
	binary.LittleEndian.PutUint32(out[2:], uint32(len(out)))
	binary.LittleEndian.PutUint32(out[10:], uint32(14+offset))
	copy(out[14:], data)
	return out, nil
}
