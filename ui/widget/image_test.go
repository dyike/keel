package widget

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
	"time"

	"gioui.org/layout"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
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
func TestDecodeImageSourcesAndLimits(t *testing.T) {
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
func TestImageAspectRatioAndClick(t *testing.T) {
	calls := 0
	view := &ImageView{Asset: ImageData(image.NewNRGBA(image.Rect(0, 0, 600, 200))), Alt: "test image", OnClick: func() { calls++ }}
	var size image.Point
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Constraints{Max: image.Pt(300, 300)}
		size = view.Layout(gtx).Size
	})
	if size != image.Pt(300, 100) {
		t.Fatalf("aspect ratio: %v", size)
	}
	h.Click(100, 50)
	if calls != 1 {
		t.Fatal("image click not dispatched")
	}
}
func TestImageAsyncLoad(t *testing.T) {
	done := make(chan struct{})
	asset := LoadImage("test", func(context.Context, string) (image.Image, error) {
		<-done
		return image.NewNRGBA(image.Rect(0, 0, 30, 10)), nil
	})
	view := &ImageView{Asset: asset, Alt: "loading fixture"}
	h := uitest.New(view)
	if asset.Ready() {
		t.Fatal("loader should still be pending")
	}
	close(done)
	for until := time.Now().Add(2 * time.Second); !asset.Ready() && time.Now().Before(until); {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if !asset.Ready() || asset.Error() != nil || asset.Size() != image.Pt(30, 10) || asset.Revision() != 2 {
		t.Fatal("async completion not reflected in UI")
	}
}
