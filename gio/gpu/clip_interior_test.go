// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
)

func TestRoundedClipInterior(t *testing.T) {
	for _, tc := range []struct {
		name           string
		rect           image.Rectangle
		transform      f32.Affine2D
		stroke, nested bool
		skip           bool
	}{
		{name: "inside", rect: image.Rect(30, 30, 40, 40), skip: true},
		{name: "edge", rect: image.Rect(0, 0, 30, 30)},
		{name: "antialias margin", rect: image.Rect(20, 20, 30, 30)},
		{name: "fractional translation fallback", rect: image.Rect(30, 30, 40, 40), transform: f32.Affine2D{}.Offset(f32.Pt(3.5, 4.5))},
		{name: "scaled rectangle fallback", rect: image.Rect(30, 30, 40, 40), transform: f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(2, 2))},
		{name: "rotated", rect: image.Rect(30, 30, 40, 40), transform: f32.Affine2D{}.Rotate(f32.Point{}, .1)},
		{name: "stroke", rect: image.Rect(30, 30, 40, 40), stroke: true},
		{name: "unknown parent", rect: image.Rect(30, 30, 40, 40), nested: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var frame op.Ops
			tr := op.Affine(tc.transform).Push(&frame)
			var parent clip.Stack
			if tc.nested {
				parent = clip.Ellipse(image.Rect(0, 0, 120, 120)).Push(&frame)
			}
			rr := clip.UniformRRect(image.Rect(0, 0, 100, 100), 20)
			co := rr.Op(&frame)
			if tc.stroke {
				co = clip.Stroke{Path: rr.Path(&frame), Width: 4}.Op()
			}
			c := co.Push(&frame)
			paint.FillShape(&frame, color.NRGBA{A: 128}, clip.Rect(tc.rect).Op())
			c.Pop()
			if tc.nested {
				parent.Pop()
			}
			tr.Pop()
			d := drawOps{pathCache: newOpCache()}
			d.reset(image.Pt(300, 300))
			d.collect(&frame, image.Pt(300, 300))
			if len(d.imageOps) != 1 {
				t.Fatalf("paint count %d", len(d.imageOps))
			}
			if got := d.imageOps[0].path == nil; got != tc.skip {
				t.Fatalf("coverage skipped=%v, want %v", got, tc.skip)
			}
		})
	}
}
