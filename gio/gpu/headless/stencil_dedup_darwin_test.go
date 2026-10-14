// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"bytes"
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"image"
	"image/color"
	"testing"
)

func TestRepeatedCoverageMatchesFreshRendering(t *testing.T) {
	continuous, err := NewWindow(240, 160)
	if err != nil {
		t.Skip(err)
	}
	defer continuous.Release()
	fresh, err := NewWindow(240, 160)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Release()
	var ops op.Ops
	for frame := 0; frame < 24; frame++ {
		ops.Reset()
		paint.Fill(&ops, color.NRGBA{R: 240, G: 240, B: 240, A: 255})
		// Pairs repeat their geometry but change paint. Then move, resize, change
		// fractional offsets, nesting, and path geometry before returning to it.
		phase := frame / 2
		offset := image.Pt(phase%4*7, phase%3*9)
		tr := op.Offset(offset).Push(&ops)
		frac := op.Affine(f32.AffineId().Offset(f32.Pt(float32(phase%2)*.5, .25))).Push(&ops)
		cl := clip.UniformRRect(image.Rect(4, 4, 190-phase%3*21, 120), 12+phase%3).Push(&ops)
		var nested clip.Stack
		if phase%2 != 0 {
			nested = clip.Ellipse(image.Rect(0, 0, 170, 120)).Push(&ops)
		}
		same := clip.UniformRRect(image.Rect(4, 4, 190-phase%3*21, 120), 12+phase%3).Push(&ops)
		alpha := paint.PushOpacity(&ops, .7)
		paint.Fill(&ops, color.NRGBA{R: byte(30 + frame*8), G: 90, B: 180, A: 255})
		alpha.Pop()
		same.Pop()
		if phase%2 != 0 {
			nested.Pop()
		}
		cl.Pop()
		frac.Pop()
		tr.Pop()
		if err := contextDo(fresh.ctx, func() error { fresh.gpu.(interface{ Trim() }).Trim(); return nil }); err != nil {
			t.Fatal(err)
		}
		if frame == 8 {
			if err := contextDo(continuous.ctx, func() error { continuous.gpu.(interface{ Idle() }).Idle(); return nil }); err != nil {
				t.Fatal(err)
			}
		}
		if frame == 16 {
			if err := contextDo(continuous.ctx, func() error { continuous.gpu.(interface{ Trim() }).Trim(); return nil }); err != nil {
				t.Fatal(err)
			}
		}
		for _, w := range []*Window{continuous, fresh} {
			if err := w.Frame(&ops); err != nil {
				t.Fatal(err)
			}
		}
		a, b := image.NewRGBA(image.Rect(0, 0, 240, 160)), image.NewRGBA(image.Rect(0, 0, 240, 160))
		if err := continuous.Screenshot(a); err != nil {
			t.Fatal(err)
		}
		if err := fresh.Screenshot(b); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Pix, b.Pix) {
			t.Fatalf("frame %d continuous coverage differs from full rendering", frame)
		}
	}
}

func BenchmarkRepeatedClipFrame(b *testing.B) {
	w, err := NewWindow(512, 512)
	if err != nil {
		b.Skip(err)
	}
	defer w.Release()
	var ops op.Ops
	paint.Fill(&ops, color.NRGBA{A: 255})
	for i := 0; i < 100; i++ {
		cl := clip.UniformRRect(image.Rect(10, 10, 500, 500), 24).Push(&ops)
		paint.Fill(&ops, color.NRGBA{R: byte(i), G: 100, A: 120})
		cl.Pop()
	}
	if err := w.Frame(&ops); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := w.Frame(&ops); err != nil {
			b.Fatal(err)
		}
	}
}
