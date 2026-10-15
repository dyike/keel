package headless

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

func strokeCacheFrame(frame *op.Ops, angle float32) {
	frame.Reset()
	paint.Fill(frame, color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	parent := clip.UniformRRect(image.Rect(3, 3, 237, 157), 17).Push(frame)
	for i := range 4 {
		layer := paint.PushOpacity(frame, .65)
		tr := op.Affine(f32.NewAffine2D(1, .2, float32(i*53+30), -.1, 1.3, 60).
			Rotate(f32.Point{}, angle+float32(i)*.2)).Push(frame)
		var p clip.Path
		p.Begin(frame)
		p.MoveTo(f32.Pt(-18, 0))
		p.CubeTo(f32.Pt(-18, -25), f32.Pt(18, -25), f32.Pt(18, 0))
		s := clip.Stroke{Path: p.End(), Width: float32(i+1) * 1.25}.Op().Push(frame)
		paint.Fill(frame, color.NRGBA{R: 200, G: 50, B: 30, A: 255})
		s.Pop()
		tr.Pop()
		layer.Pop()
	}
	parent.Pop()
}

func TestStrokeCacheWarmPixelsMatchFreshRenderer(t *testing.T) {
	w, err := NewWindow(240, 160)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	fresh, err := NewWindow(240, 160)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Release()
	var frame op.Ops
	for i := range 12 {
		strokeCacheFrame(&frame, float32(i)*.07)
		if err := w.Frame(&frame); err != nil {
			t.Fatal(err)
		}
	}
	a, b := image.NewRGBA(image.Rect(0, 0, 240, 160)), image.NewRGBA(image.Rect(0, 0, 240, 160))
	if err := w.Screenshot(a); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Frame(&frame); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Screenshot(b); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("cached transformed strokes changed clipping, opacity or antialiasing")
	}
	if trim, ok := w.gpu.(interface{ Trim() }); ok {
		if err := contextDo(w.ctx, func() error { trim.Trim(); return nil }); err != nil {
			t.Fatal(err)
		}
		if err := w.Frame(&frame); err != nil {
			t.Fatal(err)
		}
		if err := w.Screenshot(a); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Pix, b.Pix) {
			t.Fatal("stroke resume after renderer trim changed pixels")
		}
	}
}
