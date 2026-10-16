package kit

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// Assert the first painted frame, not just the final settled position: delaying
// anchor compensation until the next frame exposes a jump followed by a bounce.
func TestVariableListHeightChangesPreserveAnchorBeforePaint(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, scale := range []float32{1, 1.5, 2} {
			t.Run(fmt.Sprintf("horizontal=%v/scale=%g", horizontal, scale), func(t *testing.T) {
				keys := variableKeys(100)
				heights := map[string]float32{}
				var cx *el.Context
				list := VariableList(keys, 40, func(_ *el.Context, i int) el.Element {
					h := float32(40)
					if changed, ok := heights[keys[i]]; ok {
						h = changed
					}
					row := el.Div().Name("row-" + keys[i])
					if horizontal {
						return row.W(el.Dp(h)).H(el.Dp(40))
					}
					return row.H(el.Dp(h))
				}).Height(200).Width(200).Horizontal(horizontal)
				root := el.Root(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return list.Render(c) }))
				h := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
					gtx.Constraints.Max = image.Pt(int(400*scale), int(300*scale))
					root.Layout(gtx)
				})
				settle(h)
				list.ScrollToKeyAlign(cx, "20", ScrollStart)
				settle(h)
				anchor := list.anchor
				above := keys[list.indices[anchor]-1]
				position := func() int {
					t.Helper()
					n, ok := node(h, "row-"+anchor)
					if !ok {
						t.Fatal("lost visible anchor")
					}
					if horizontal {
						return n.Desc.Bounds.Min.X
					}
					return n.Desc.Bounds.Min.Y
				}
				before := position()
				for _, size := range []float32{121, 21, 201, 41} {
					heights[above] = size
					for frame := range 3 {
						h.Frame()
						if got := position(); got != before {
							t.Fatalf("height %g, frame %d: anchor moved %d -> %d", size, frame, before, got)
						}
					}
				}
				// A wheel event in the measurement frame must still take effect once.
				heights[above] = 160
				wheel := f32.Pt(0, 20*scale)
				if horizontal {
					wheel = f32.Pt(20*scale, 0)
				}
				h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(100*scale, 100*scale), Scroll: wheel})
				for frame := range 3 {
					h.Frame()
					if got, want := position(), before-int(20*scale); got != want {
						t.Fatalf("wheel with height change, frame %d: got %d, want %d", frame, got, want)
					}
				}
			})
		}
	}
}
