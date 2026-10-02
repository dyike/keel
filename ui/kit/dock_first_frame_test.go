package kit

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestDockFirstFrameAndResizeKeepCenter(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Dock(text("Center")).Panel(DockPanel{ID: "left", Title: "Left", View: text("L")}, DockLeft).Panel(DockPanel{ID: "right", Title: "Right", View: text("R")}, DockRight).Panel(DockPanel{ID: "bottom", Title: "Bottom", View: text("B")}, DockBottom)
		root := el.Root(v)
		size := image.Pt(320, 240)
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints = layout.Exact(size.Mul(scale))
			root.Layout(gtx)
		})
		check := func() {
			t.Helper()
			r := v.centerRect
			if r.w < min(float32(size.X)-8, 120)-1 || r.h < min(float32(size.Y)-4, 120)-1 || r.x < 0 || r.y < 0 || r.x+r.w > float32(size.X)+1 || r.y+r.h > float32(size.Y)+1 {
				t.Fatal("center outside first frame", size, r)
			}
		}
		check()
		for _, next := range []image.Point{image.Pt(800, 600), image.Pt(280, 200), image.Pt(100, 100)} {
			size = next
			h.Frame()
			check()
		}
	}
}
