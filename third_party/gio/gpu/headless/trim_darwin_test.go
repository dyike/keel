// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func TestTrimResumesRendering(t *testing.T) {
	w, err := NewWindow(80, 80)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	trim := func() {
		t.Helper()
		if err := contextDo(w.ctx, func() error { w.gpu.(interface{ Trim() }).Trim(); return nil }); err != nil {
			t.Fatal(err)
		}
	}
	shot := func() []byte {
		t.Helper()
		img := image.NewRGBA(image.Rect(0, 0, 80, 80))
		if err := w.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		return img.Pix
	}
	for pass := 0; pass < 3; pass++ {
		// New image handles exercise fresh uploads after trimming, while a curved
		// clip also exercises the general renderer alongside the quad batch path.
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				img.SetRGBA(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 7), B: uint8(pass * 70), A: 255})
			}
		}
		var ops op.Ops
		paint.Fill(&ops, color.NRGBA{R: 20, A: 255})
		for i := 0; i < 2; i++ {
			tr := op.Offset(image.Pt(i*36, 10)).Push(&ops)
			c := clip.UniformRRect(image.Rect(0, 0, 32, 32), i*8).Push(&ops)
			paint.NewImageOp(img).Add(&ops)
			paint.PaintOp{}.Add(&ops)
			c.Pop()
			tr.Pop()
		}
		if err := w.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		trim() // Also exercise the submitted frame before a synchronous readback.
		before := shot()
		trim()
		trim() // Idempotent, including readback buffer cleanup.
		if got := shot(); !bytes.Equal(before, got) {
			t.Fatal("trim changed framebuffer")
		}
		if err := w.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		if got := shot(); !bytes.Equal(before, got) {
			t.Fatal("resumed frame differs")
		}
	}
}
