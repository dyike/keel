// SPDX-License-Identifier: Unlicense OR MIT

package headless

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

// TestRenderPixels renders filled and clipped colors and reads them back.
// Keel patch: it checks whichever context the platform selects, including
// the Linux EGL and OpenGL ES bindings without cgo. Set
// GIO_HEADLESS_REQUIRED=1 to fail rather than skip when no context exists.
func TestRenderPixels(t *testing.T) {
	checkRenderPixels(t)
}

func checkRenderPixels(t *testing.T) {
	t.Helper()
	w, err := NewWindow(64, 48)
	if err != nil {
		if os.Getenv("GIO_HEADLESS_REQUIRED") == "1" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
	defer w.Release()
	var ops op.Ops
	paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	c := clip.Rect(image.Rect(8, 8, 40, 24)).Push(&ops)
	paint.Fill(&ops, color.NRGBA{R: 200, G: 30, B: 60, A: 255})
	c.Pop()
	c = clip.Ellipse(image.Rect(40, 24, 64, 48)).Push(&ops)
	paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 255, A: 128})
	c.Pop()
	if err := w.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	if err := w.Screenshot(img); err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct {
		x, y int
		want color.RGBA
	}{
		{2, 2, color.RGBA{255, 255, 255, 255}},
		{20, 16, color.RGBA{200, 30, 60, 255}},
		{52, 36, color.RGBA{187, 187, 255, 255}}, // half blue over white, blended in linear light
	} {
		got := img.RGBAAt(p.x, p.y)
		if diff(got.R, p.want.R) > 2 || diff(got.G, p.want.G) > 2 || diff(got.B, p.want.B) > 2 || got.A != p.want.A {
			t.Errorf("pixel (%d,%d) = %v, want %v", p.x, p.y, got, p.want)
		}
	}
}

func diff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
