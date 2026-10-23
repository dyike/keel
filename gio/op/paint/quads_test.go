// SPDX-License-Identifier: Unlicense OR MIT

package paint_test

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

// PaintQuads must draw exactly what the offset, clip and paint operations it
// replaces draw, under every transform, clip and opacity a glyph run meets.
func TestPaintQuadsMatchesSeparateOperations(t *testing.T) {
	w, err := headless.NewWindow(96, 96)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	page := image.NewRGBA(image.Rect(0, 0, 40, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 40; x++ {
			a := uint8((x*37 + y*91) % 256)
			page.SetRGBA(x, y, color.RGBA{R: a, G: uint8(x * 6), B: uint8(y*10) & a, A: a})
		}
	}
	quads := []paint.ImageQuad{
		{Offset: image.Pt(4, 6), Rect: image.Rect(0, 0, 9, 12)},
		{Offset: image.Pt(10, 6), Rect: image.Rect(9, 0, 20, 12)},
		{Offset: image.Pt(20, 30), Rect: image.Rect(20, 12, 40, 24)},
		{Offset: image.Pt(-30, 50), Rect: image.Rect(30, 0, 40, 24)}, // overlaps a neighbor
		{Offset: image.Pt(200, 200), Rect: image.Rect(0, 0, 8, 8)},   // off screen
		{Offset: image.Pt(5, 70), Rect: image.Rect(0, 0, 0, 0)},      // empty
	}
	modes := []string{"offset", "fractional offset", "scale", "rotate", "rect clip", "rounded clip", "opacity", "nested opacity", "brush after"}
	for _, mode := range modes {
		var shots [2]*image.RGBA
		for i := range shots {
			var ops op.Ops
			paint.Fill(&ops, color.NRGBA{R: 30, G: 40, B: 50, A: 255})
			var outer paint.OpacityStack
			if mode == "nested opacity" {
				outer = paint.PushOpacity(&ops, .8)
			}
			at := f32.AffineId().Offset(f32.Pt(7, 3))
			switch mode {
			case "fractional offset":
				at = f32.AffineId().Offset(f32.Pt(7.4, 3.6))
			case "scale":
				at = at.Scale(f32.Point{}, f32.Pt(1.5, 1.25))
			case "rotate":
				at = at.Rotate(f32.Pt(40, 40), math.Pi/9)
			}
			tr := op.Affine(at).Push(&ops)
			var cl clip.Stack
			switch mode {
			case "rect clip":
				cl = clip.Rect(image.Rect(8, 9, 50, 60)).Push(&ops)
			case "rounded clip":
				cl = clip.UniformRRect(image.Rect(2, 2, 60, 70), 14).Push(&ops)
			}
			var layer paint.OpacityStack
			if mode == "opacity" || mode == "nested opacity" {
				layer = paint.PushOpacity(&ops, .5)
			}
			img := paint.NewImageOp(page)
			img.Filter = paint.FilterNearest
			if i == 0 {
				for _, q := range quads {
					o := op.Offset(q.Offset).Push(&ops)
					c := clip.Rect(q.Rect).Push(&ops)
					img.Add(&ops)
					paint.PaintOp{}.Add(&ops)
					c.Pop()
					o.Pop()
				}
			} else {
				img.PaintQuads(&ops, quads)
			}
			if mode == "brush after" {
				// The image stays the brush, as after the separate operations.
				c := clip.Rect(image.Rect(60, 0, 90, 30)).Push(&ops)
				paint.PaintOp{}.Add(&ops)
				c.Pop()
			}
			if mode == "opacity" || mode == "nested opacity" {
				layer.Pop()
			}
			if mode == "rect clip" || mode == "rounded clip" {
				cl.Pop()
			}
			tr.Pop()
			if mode == "nested opacity" {
				outer.Pop()
			}
			if err := w.Frame(&ops); err != nil {
				t.Fatal(err)
			}
			shots[i] = image.NewRGBA(image.Rect(0, 0, 96, 96))
			if err := w.Screenshot(shots[i]); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(shots[0].Pix, shots[1].Pix) {
			t.Errorf("%s: PaintQuads pixels differ from separate operations", mode)
		}
	}
}

// A uniform or empty image falls back to the separate operations.
func TestPaintQuadsUniformImage(t *testing.T) {
	w, err := headless.NewWindow(32, 32)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	uniform := paint.NewImageOp(image.NewUniform(color.NRGBA{R: 200, A: 255}))
	var shots [2]*image.RGBA
	for i := range shots {
		var ops op.Ops
		q := paint.ImageQuad{Offset: image.Pt(3, 4), Rect: image.Rect(0, 0, 10, 10)}
		if i == 0 {
			o := op.Offset(q.Offset).Push(&ops)
			c := clip.Rect(q.Rect).Push(&ops)
			uniform.Add(&ops)
			paint.PaintOp{}.Add(&ops)
			c.Pop()
			o.Pop()
		} else {
			uniform.PaintQuads(&ops, []paint.ImageQuad{q})
			paint.NewImageOp(image.NewRGBA(image.Rectangle{})).PaintQuads(&ops, []paint.ImageQuad{q})
		}
		if err := w.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		shots[i] = image.NewRGBA(image.Rect(0, 0, 32, 32))
		if err := w.Screenshot(shots[i]); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(shots[0].Pix, shots[1].Pix) {
		t.Error("uniform PaintQuads pixels differ from separate operations")
	}
}

// BenchmarkPaintQuads encodes and draws a screen of glyph-sized quads.
func BenchmarkPaintQuads(b *testing.B) {
	w, err := headless.NewWindow(800, 600)
	if err != nil {
		b.Skip(err)
	}
	defer w.Release()
	page := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for i := range page.Pix {
		page.Pix[i] = uint8(i)
	}
	img := paint.NewImageOp(page)
	var quads []paint.ImageQuad
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			r := image.Rect(x%16*16, y%16*16, x%16*16+9, y%16*16+14)
			quads = append(quads, paint.ImageQuad{Offset: image.Pt(x*10, y*15).Sub(r.Min), Rect: r})
		}
	}
	for _, mode := range []string{"Separate", "Quads"} {
		b.Run(mode, func(b *testing.B) {
			var ops op.Ops
			b.ReportAllocs()
			for b.Loop() {
				ops.Reset()
				if mode == "Quads" {
					img.PaintQuads(&ops, quads)
				} else {
					for _, q := range quads {
						o := op.Offset(q.Offset).Push(&ops)
						c := clip.Rect(q.Rect).Push(&ops)
						img.Add(&ops)
						paint.PaintOp{}.Add(&ops)
						c.Pop()
						o.Pop()
					}
				}
				if err := w.Frame(&ops); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
