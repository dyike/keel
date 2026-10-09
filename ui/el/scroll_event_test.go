package el

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestOnScrollRangesAndDisabledAncestors(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		child, parent := ScrollEvent{}, ScrollEvent{}
		disabled := false
		root := Embed(ViewFunc(func(*Context) Element {
			return Div().W(Dp(200)).H(Dp(100)).OnScroll(ScrollRange{-1000, 1000}, ScrollRange{-1000, 1000}, func(e ScrollEvent) { parent.X += e.X; parent.Y += e.Y }).Child(
				Div().Disabled(disabled).Child(Div().W(Dp(100)).H(Dp(80)).OnScroll(ScrollRange{-10, 10}, ScrollRange{}, func(e ScrollEvent) { child.X += e.X; child.Y += e.Y })))
		}))
		h := uitest.NewFunc(func(gtx core.C) { gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}; root.Layout(gtx) })
		scroll := func() {
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(30*scale, 30*scale), Scroll: f32.Pt(25*scale, 15*scale)})
			h.Frame()
		}
		scroll()
		if child.X != 10 || child.Y != 0 || parent.X != 15 || parent.Y != 15 {
			t.Fatal("ranges/axis routing", scale, child, parent)
		}
		child, parent = ScrollEvent{}, ScrollEvent{}
		disabled = true
		h.Frame()
		scroll()
		if child.X != 0 || child.Y != 0 || parent.X != 25 || parent.Y != 15 {
			t.Fatal("disabled child stole scroll", child, parent)
		}
	}
}
