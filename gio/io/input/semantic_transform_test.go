// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

func TestSemanticTransformedBounds(t *testing.T) {
	for _, tc := range []struct {
		name      string
		transform f32.Affine2D
		want      image.Rectangle
	}{
		{"translation", f32.NewAffine2D(1, 0, 50, 0, 1, 20), image.Rect(50, 20, 80, 30)},
		{"quarter turn", f32.NewAffine2D(0, -1, 50, 1, 0, 20), image.Rect(40, 20, 50, 50)},
		{"reflection", f32.NewAffine2D(-2, 0, 80, 0, 2, 20), image.Rect(20, 20, 80, 40)},
		{"shear", f32.NewAffine2D(1, -2, 50, 1, 1, 20), image.Rect(30, 20, 80, 60)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var r Router
			var ops op.Ops
			tr := op.Affine(tc.transform).Push(&ops)
			area := clip.Rect(image.Rect(0, 0, 30, 10)).Push(&ops)
			semantic.Button.Add(&ops)
			semantic.LabelOp("transformed").Add(&ops)
			area.Pop()
			tr.Pop()
			r.Frame(&ops)
			for _, n := range r.AppendSemantics(nil) {
				if n.Desc.Label == "transformed" {
					if n.Desc.Bounds != tc.want {
						t.Fatalf("bounds=%v, want %v", n.Desc.Bounds, tc.want)
					}
					return
				}
			}
			t.Fatal("missing semantic node")
		})
	}
}
