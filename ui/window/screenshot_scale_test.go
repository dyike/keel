package window

import (
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestScreenshotScale(t *testing.T) {
	for _, scale := range []float32{1, 1.5, 2} {
		path := filepath.Join(t.TempDir(), "shot.png")
		if err := ScreenshotAtScale(nil, 100, 60, scale, path); err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		config, err := png.DecodeConfig(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if config.Width != int(100*scale) || config.Height != int(60*scale) {
			t.Fatal(config)
		}
	}
	for _, scale := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1)), 0.0001} {
		if _, err := screenshotSize(100, 60, scale); err == nil {
			t.Fatalf("accepted %g", scale)
		}
	}
	if _, err := screenshotSize(0, 60, 1); err == nil {
		t.Fatal("accepted zero width")
	}
	if p, err := screenshotSize(3, 5, 1.5); err != nil || p != image.Pt(5, 8) {
		t.Fatal(p, err)
	}
}
