// SPDX-License-Identifier: Unlicense OR MIT

package headless

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

func TestGradientScalePixels(t *testing.T) {
	w, err := NewWindow(140, 140)
	if err != nil {
		if os.Getenv("GIO_HEADLESS_REQUIRED") == "1" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
	defer w.Release()
	var baseline *image.RGBA
	for _, scale := range []float32{1, .5, 2} {
		var ops op.Ops
		paint.Fill(&ops, color.NRGBA{A: 255})
		transform := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(f32.Pt(20, 20))).Push(&ops)
		size := int(80 / scale)
		area := clip.Rect(image.Rect(0, 0, size, size)).Push(&ops)
		paint.LinearGradientOp{Stop1: f32.Pt(0, 0), Stop2: f32.Pt(float32(size), float32(size)), Color1: color.NRGBA{R: 255, A: 255}, Color2: color.NRGBA{B: 255, A: 255}}.Add(&ops)
		paint.PaintOp{}.Add(&ops)
		area.Pop()
		transform.Pop()
		if err := w.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, 140, 140))
		if err := w.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		if baseline == nil {
			baseline = img
		} else {
			maxDelta := 0
			for i, v := range baseline.Pix {
				maxDelta = max(maxDelta, diff(v, img.Pix[i]))
			}
			if maxDelta > 2 {
				t.Fatalf("gradient changed under equivalent scale %g: max channel delta %d", scale, maxDelta)
			}
		}
	}
}
