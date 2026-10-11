// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// Full-window overlays discard the covered content. Completed opacity layers
// must be discarded with it, or their old indexes replace unrelated new ops.
func TestOpaqueClearDiscardsCoveredOpacityLayers(t *testing.T) {
	var ops op.Ops
	old := paint.PushOpacity(&ops, .5)
	paint.FillShape(&ops, color.NRGBA{A: 255}, clip.Rect(image.Rect(10, 10, 20, 20)).Op())
	old.Pop()
	paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	paint.FillShape(&ops, color.NRGBA{A: 255}, clip.Rect(image.Rect(40, 40, 50, 50)).Op())
	current := paint.PushOpacity(&ops, .8)
	paint.FillShape(&ops, color.NRGBA{A: 255}, clip.Rect(image.Rect(60, 60, 70, 70)).Op())
	current.Pop()
	var d drawOps
	size := image.Pt(100, 100)
	d.reset(size)
	d.collect(&ops, size)
	if len(d.layers) != 1 {
		t.Fatalf("retained covered layers: %+v", d.layers)
	}
	if l := d.layers[0]; l.opStart != 1 || l.opEnd != 2 || l.parent != -1 {
		t.Fatalf("wrong surviving layer: %+v", l)
	}
	if len(d.imageOps) != 2 {
		t.Fatalf("got %d paint ops", len(d.imageOps))
	}
}

// A partial cover or a fill inside an opacity group still needs earlier layers.
func TestOpaqueFillPreservesVisibleOpacityLayers(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(map[bool]string{false: "partial-cover", true: "active-opacity"}[active], func(t *testing.T) {
			var ops op.Ops
			old := paint.PushOpacity(&ops, .5)
			paint.FillShape(&ops, color.NRGBA{A: 255}, clip.Rect(image.Rect(10, 10, 20, 20)).Op())
			old.Pop()
			if active {
				group := paint.PushOpacity(&ops, .8)
				paint.Fill(&ops, color.NRGBA{R: 255, A: 255})
				group.Pop()
			} else {
				paint.FillShape(&ops, color.NRGBA{R: 255, A: 255}, clip.Rect(image.Rect(40, 40, 50, 50)).Op())
			}
			var d drawOps
			size := image.Pt(100, 100)
			d.reset(size)
			d.collect(&ops, size)
			want := 1
			if active {
				want = 2
			}
			if len(d.layers) != want || len(d.imageOps) != 2 {
				t.Fatalf("visible content discarded: layers=%+v, ops=%d", d.layers, len(d.imageOps))
			}
			if l := d.layers[0]; l.opStart != 0 || l.opEnd != 1 {
				t.Fatalf("covered layer changed: %+v", l)
			}
		})
	}
}
