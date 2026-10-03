package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func imageFixture(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 360, 120))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
func TestDecodeSourcesAndLimits(t *testing.T) {
	data := imageFixture(t)
	path := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	}))
	defer server.Close()
	for _, source := range []string{path, "file://" + path, server.URL, "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)} {
		img, err := DecodeImage(context.Background(), source)
		if err != nil || img.Bounds().Size() != image.Pt(360, 120) {
			t.Fatalf("decode %q: %v", source, err)
		}
	}
	for _, source := range []string{path + "-missing", server.URL + "/missing", "data:image/png;base64,bad", "unknown://image"} {
		if _, err := DecodeImage(context.Background(), source); err == nil {
			t.Fatalf("expected image failure: %s", source)
		}
	}
	// A valid PNG header with enormous dimensions must fail before allocation.
	large := append([]byte(nil), data...)
	binary.BigEndian.PutUint32(large[16:20], 100000)
	binary.BigEndian.PutUint32(large[20:24], 100000)
	binary.BigEndian.PutUint32(large[29:33], crc32.ChecksumIEEE(large[12:29]))
	if _, err := DecodeImage(context.Background(), "data:image/png;base64,"+base64.StdEncoding.EncodeToString(large)); err == nil {
		t.Fatal("unbounded decoded pixel allocation")
	}
}
