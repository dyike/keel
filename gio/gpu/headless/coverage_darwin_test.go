// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

func TestCoveredTintedQuadsPixels(t *testing.T) {
	w, err := NewWindow(240, 40)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	var frame op.Ops
	paint.Fill(&frame, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	for i := 0; i < 10; i++ {
		// Distinct images exhaust batch texture slots; nested paths use intersection coverage.
		src := image.NewRGBA(image.Rect(0, 0, 20, 20))
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				src.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
		tr := op.Offset(image.Pt(i*24, 10)).Push(&frame)
		c := clip.UniformRRect(image.Rect(0, 0, 20, 20), 8).Push(&frame)
		var nested clip.Stack
		if i%2 != 0 {
			nested = clip.Ellipse(image.Rect(0, 0, 20, 20)).Push(&frame)
		}
		paint.NewImageOp(src).AddTinted(&frame, color.NRGBA{R: 220, G: 30, B: 50, A: 255})
		paint.PaintOp{}.Add(&frame)
		if i%2 != 0 {
			nested.Pop()
		}
		c.Pop()
		tr.Pop()
	}
	if err = w.Frame(&frame); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 240, 40))
	if err = w.Screenshot(img); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		if c := img.RGBAAt(i*24, 10); c != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
			t.Fatalf("corner %d not clipped: %v", i, c)
		}
		c := img.RGBAAt(i*24+10, 20)
		if absByte(c.R, 220) > 1 || absByte(c.G, 30) > 1 || absByte(c.B, 50) > 1 || c.A != 255 {
			t.Fatalf("center %d tint: %v", i, c)
		}
	}
}
func absByte(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
