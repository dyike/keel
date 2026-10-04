package kit

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/ui/theme"
)

const testSVG = `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 20"><rect width="40" height="20" fill="#f00"/></svg>`

func testGIF(t *testing.T, delay int) []byte {
	t.Helper()
	g := &gif.GIF{}
	for i, c := range []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}} {
		frame := image.NewPaletted(image.Rect(0, 0, 4, 4), palette.Plan9)
		// The second frame covers only the left half; the right half keeps
		// the first frame (disposal none).
		w := 4
		if i == 1 {
			w = 2
		}
		for y := 0; y < 4; y++ {
			for x := 0; x < w; x++ {
				frame.Set(x, y, c)
			}
		}
		if i == 1 {
			frame = frame.SubImage(image.Rect(0, 0, 2, 4)).(*image.Paletted)
		}
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, delay)
		g.Disposal = append(g.Disposal, gif.DisposalNone)
	}
	var b bytes.Buffer
	if err := gif.EncodeAll(&b, g); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestImageMediaDecodesSVGAndGIF(t *testing.T) {
	m, err := decodeImageMedia([]byte("\xef\xbb\xbf\n" + testSVG))
	if err != nil || m.svg == nil || m.animated() {
		t.Fatal("svg", err)
	}
	if b := m.still.Bounds(); b.Dx() != 64 || b.Dy() != 32 {
		t.Fatal("svg still keeps the viewBox aspect", b)
	}
	m, err = decodeImageMedia(testGIF(t, 5))
	if err != nil || !m.animated() || len(m.frames) != 2 || m.delays[0] != 50*time.Millisecond {
		t.Fatal("gif", err)
	}
	second := m.frames[1].(*image.RGBA)
	if r, _, b, _ := second.At(0, 0).RGBA(); b == 0 || r != 0 {
		t.Fatal("second frame draws its own pixels")
	}
	if r, _, _, _ := second.At(3, 0).RGBA(); r == 0 {
		t.Fatal("second frame keeps the first outside its bounds")
	}
	if m, _ := decodeImageMedia(testGIF(t, 0)); m.delays[0] != 100*time.Millisecond {
		t.Fatal("zero delay defaults like browsers")
	}
}

func TestImageViewAnimatesGIFAndRendersSVG(t *testing.T) {
	src := "data:image/gif;base64," + base64.StdEncoding.EncodeToString(testGIF(t, 2))
	v := Image(nil, "spinner").Width(40).Cache(NewImageCache(1 << 20)).Source(src)
	h := renderView(v, 200, 1)
	deadline := time.Now().Add(2 * time.Second)
	for v.media == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
		h.Frame()
	}
	if v.media == nil || !v.media.animated() {
		t.Fatal("gif not loaded", v.err)
	}
	for v.frame == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		h.Frame()
	}
	if v.frame == 0 {
		t.Fatal("gif did not advance")
	}

	theme.ReducedMotion = true
	defer func() { theme.ReducedMotion = false }()
	h.Frame()
	if v.frame != 0 {
		t.Fatal("reduced motion holds the first frame")
	}

	svg := Image(nil, "logo").Width(120).Cache(NewImageCache(1 << 20)).Source("data:image/svg+xml," + strings.ReplaceAll(testSVG, "#", "%23"))
	h = renderView(svg, 200, 1)
	for svg.media == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
		h.Frame()
	}
	if svg.media == nil || svg.media.svg == nil {
		t.Fatal("svg not loaded", svg.err)
	}
	for _, fit := range []ImageFit{ImageContain, ImageCover, ImageFill} {
		svg.Size(120, 120).Fit(fit)
		h.Frame()
	}
}

func TestImageDiskCacheRevalidatesAndEvicts(t *testing.T) {
	data := sourceTestPNG(t)
	var requests, conditional atomic.Int32
	var offline atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if offline.Load() {
			http.Error(w, "down", 503)
			return
		}
		if r.Header.Get("If-None-Match") == `"v1"` {
			conditional.Add(1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		w.Write(data)
	}))
	defer server.Close()
	dir := t.TempDir()
	now := time.Now()
	newCache := func() *ImageCache {
		c := NewImageCache(1<<20).Disk(dir, 1<<20, time.Hour)
		c.disk.now = func() time.Time { return now }
		return c
	}
	if _, err := newCache().load(context.Background(), server.URL+"/a"); err != nil || requests.Load() != 1 {
		t.Fatal("first download", err, requests.Load())
	}
	// A new cache (a restart) reads the fresh copy from disk.
	if m, err := newCache().load(context.Background(), server.URL+"/a"); err != nil || m.still.Bounds().Dx() != 8 || requests.Load() != 1 {
		t.Fatal("disk hit", err, requests.Load())
	}
	now = now.Add(2 * time.Hour)
	if _, err := newCache().load(context.Background(), server.URL+"/a"); err != nil || conditional.Load() != 1 {
		t.Fatal("stale copy revalidates", err, conditional.Load())
	}
	now = now.Add(2 * time.Hour)
	offline.Store(true)
	if _, err := newCache().load(context.Background(), server.URL+"/a"); err != nil {
		t.Fatal("server error falls back to the stale copy", err)
	}
	offline.Store(false)

	// A tight budget keeps only the most recent body.
	small := NewImageCache(1<<20).Disk(dir, int64(len(data))+1, time.Hour)
	for _, p := range []string{"/b", "/c"} {
		now = now.Add(time.Minute)
		if _, err := small.load(context.Background(), server.URL+p); err != nil {
			t.Fatal(err)
		}
	}
	bodies, _ := filepath.Glob(filepath.Join(dir, "*.img"))
	if len(bodies) != 1 {
		t.Fatal("eviction kept", len(bodies))
	}
	body, _ := small.disk.paths(server.URL + "/c")
	if _, err := os.Stat(body); err != nil {
		t.Fatal("newest body evicted")
	}
}
