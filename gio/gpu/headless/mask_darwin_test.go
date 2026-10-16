// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"image"
	"image/color"
	"math"
	"testing"
)

// Compare the R8 sampled brush with its former sRGB RGBA representation,
// including mipmaps, antialiased clips, tint and opacity-layer fallbacks.
func TestSingleChannelMaskPixels(t *testing.T) {
	w, err := NewWindow(180, 120)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	backing := image.NewAlpha(image.Rect(0, 0, 75, 53))
	for y := 0; y < 53; y++ {
		for x := 0; x < 75; x++ {
			backing.SetAlpha(x, y, color.Alpha{uint8((x*13 + y*7) % 256)})
		}
	}
	mask := backing.SubImage(image.Rect(3, 5, 70, 49)).(*image.Alpha)
	rgba := image.NewRGBA(mask.Rect)
	for y := mask.Rect.Min.Y; y < mask.Rect.Max.Y; y++ {
		for x := mask.Rect.Min.X; x < mask.Rect.Max.X; x++ {
			a := mask.AlphaAt(x, y).A
			v := float64(a) / 255
			if v <= .0031308 {
				v *= 12.92
			} else {
				v = 1.055*math.Pow(v, 1/2.4) - .055
			}
			q := uint8(math.Round(v * 255))
			rgba.SetRGBA(x, y, color.RGBA{q, q, q, a})
		}
	}
	for _, ink := range []color.NRGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 211, G: 73, B: 157, A: 189}} {
		for _, scale := range []float32{1, .53, 1.7} {
			for _, opacity := range []float32{1, .43} {
				render := func(img paint.ImageOp) *image.RGBA {
					var frame op.Ops
					paint.Fill(&frame, color.NRGBA{R: 31, G: 41, B: 61, A: 255})
					tr := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(f32.Pt(7.25, 9.5))).Push(&frame)
					c := clip.UniformRRect(image.Rect(0, 0, 67, 44), 9).Push(&frame)
					c2 := clip.Ellipse(image.Rect(-12, -9, 80, 55)).Push(&frame)
					layer := paint.PushOpacity(&frame, opacity)
					img.AddTinted(&frame, ink)
					paint.PaintOp{}.Add(&frame)
					layer.Pop()
					c2.Pop()
					c.Pop()
					tr.Pop()
					if err := w.Frame(&frame); err != nil {
						t.Fatal(err)
					}
					dst := image.NewRGBA(image.Rect(0, 0, 180, 120))
					if err := w.Screenshot(dst); err != nil {
						t.Fatal(err)
					}
					return dst
				}
				got, want := render(paint.NewMaskOp(mask)), render(paint.NewImageOp(rgba))
				maxDiff := 0
				for i, v := range got.Pix {
					d := int(v) - int(want.Pix[i])
					if d < 0 {
						d = -d
					}
					maxDiff = max(maxDiff, d)
				}
				if maxDiff > 2 {
					t.Errorf("scale=%v opacity=%v max byte difference=%d", scale, opacity, maxDiff)
				}
			}
		}
	}
}
