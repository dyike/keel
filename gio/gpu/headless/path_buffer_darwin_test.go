// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

// rotatingPathFrame deliberately produces more than 4 KiB of path vertices.
// Its non-translation transform changes every frame, which makes the GPU path
// cache miss without changing the path in local coordinates.
func rotatingPathFrame(frame *op.Ops, angle float32) {
	frame.Reset()
	paint.Fill(frame, color.NRGBA{R: 245, G: 247, B: 250, A: 255})
	ancestor := clip.UniformRRect(image.Rect(12, 12, 244, 244), 29).Push(frame)
	layer := paint.PushOpacity(frame, .73)
	center := f32.Pt(128, 128)
	transform := op.Affine(f32.AffineId().Rotate(center, angle)).Push(frame)
	var path clip.Path
	path.Begin(frame)
	for i := range 96 {
		a := float64(i) * 2 * math.Pi / 96
		radius := float32(104)
		if i%2 != 0 {
			radius = 57
		}
		point := f32.Pt(
			center.X+radius*float32(math.Cos(a)),
			center.Y+radius*float32(math.Sin(a)),
		)
		if i == 0 {
			path.MoveTo(point)
		} else {
			path.LineTo(point)
		}
	}
	path.Close()
	shape := clip.Outline{Path: path.End()}.Op().Push(frame)
	paint.Fill(frame, color.NRGBA{R: 34, G: 113, B: 210, A: 255})
	shape.Pop()
	transform.Pop()
	layer.Pop()
	ancestor.Pop()
}

func TestRotatingPathPooledPixelsMatchDedicated(t *testing.T) {
	pooled, err := NewWindow(256, 256)
	if err != nil {
		t.Skip(err)
	}
	defer pooled.Release()
	dedicated, err := NewWindow(256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer dedicated.Release()

	var frame op.Ops
	got := image.NewRGBA(image.Rect(0, 0, 256, 256))
	want := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for i := range 24 {
		rotatingPathFrame(&frame, float32(i)*.037)
		if err := pooled.Frame(&frame); err != nil {
			t.Fatal(err)
		}
		// Trim drops the buffer pool and path cache, preserving the old
		// dedicated-allocation behavior as a pixel oracle.
		if err := contextDo(dedicated.ctx, func() error {
			dedicated.gpu.(interface{ Trim() }).Trim()
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := dedicated.Frame(&frame); err != nil {
			t.Fatal(err)
		}
		if err := pooled.Screenshot(got); err != nil {
			t.Fatal(err)
		}
		if err := dedicated.Screenshot(want); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Pix, want.Pix) {
			t.Fatalf("frame %d: pooled rotating path changed antialiasing, clipping or opacity", i)
		}
	}
}

func BenchmarkRotatingPathFrames(b *testing.B) {
	w, err := NewWindow(256, 256)
	if err != nil {
		b.Skip(err)
	}
	defer w.Release()
	var frame op.Ops
	for i := range 8 {
		rotatingPathFrame(&frame, float32(i)*.037)
		if err := w.Frame(&frame); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rotatingPathFrame(&frame, float32(i+8)*.037)
		if err := w.Frame(&frame); err != nil {
			b.Fatal(err)
		}
	}
}
