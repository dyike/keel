// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package paint_test

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/internal/f32color"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func TestTintedImagePixels(t *testing.T) {
	w, err := headless.NewWindow(80, 80)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	for _, tint := range []color.NRGBA{{R: 220, G: 120, B: 35, A: 255}, {R: 35, G: 150, B: 240, A: 128}, {R: 255, G: 255, B: 255, A: 255}, {}} {
		mask, colored := image.NewRGBA(image.Rect(0, 0, 32, 32)), image.NewRGBA(image.Rect(0, 0, 32, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				a := uint8(x * 255 / 31)
				mask.SetRGBA(x, y, f32color.NRGBAToRGBA(color.NRGBA{R: 255, G: 255, B: 255, A: a}))
				c := tint
				c.A = uint8((uint32(a)*uint32(c.A) + 127) / 255)
				colored.SetRGBA(x, y, f32color.NRGBAToRGBA(c))
			}
		}
		for _, mode := range []string{"quad", "path", "opacity", "transform"} {
			var shots [2]*image.RGBA
			for i := range shots {
				var ops op.Ops
				paint.Fill(&ops, color.NRGBA{R: 40, G: 50, B: 60, A: 255})
				tr := op.Offset(image.Pt(20, 20)).Push(&ops)
				var opacity paint.OpacityStack
				if mode == "opacity" {
					opacity = paint.PushOpacity(&ops, .5)
				}
				var transform op.TransformStack
				if mode == "transform" {
					transform = op.Affine(f32.AffineId().Rotate(f32.Pt(16, 16), .2)).Push(&ops)
				}
				var c clip.Stack
				if mode == "path" {
					c = clip.UniformRRect(image.Rect(0, 0, 32, 32), 8).Push(&ops)
				} else {
					c = clip.Rect(image.Rect(0, 0, 32, 32)).Push(&ops)
				}
				if i == 0 {
					img := paint.NewImageOp(colored)
					img.Filter = paint.FilterNearest
					img.Add(&ops)
				} else {
					img := paint.NewImageOp(mask)
					img.Filter = paint.FilterNearest
					img.AddTinted(&ops, tint)
				}
				paint.PaintOp{}.Add(&ops)
				c.Pop()
				if mode == "transform" {
					transform.Pop()
				}
				if mode == "opacity" {
					opacity.Pop()
				}
				tr.Pop()
				if err := w.Frame(&ops); err != nil {
					t.Fatal(err)
				}
				shots[i] = image.NewRGBA(image.Rect(0, 0, 80, 80))
				if err := w.Screenshot(shots[i]); err != nil {
					t.Fatal(err)
				}
			}
			for i, a := range shots[0].Pix {
				d := int(a) - int(shots[1].Pix[i])
				if d < 0 {
					d = -d
				}
				// Encoding the white coverage and multiplying in the shader introduces
				// one additional 8-bit quantization compared with precolored pixels.
				if d > 2 {
					t.Fatalf("mode=%s tint=%v byte=%d delta=%d", mode, tint, i, d)
				}
			}
		}
	}
}
