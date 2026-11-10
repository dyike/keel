package motion

import (
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestScrollTimelineTracksNativeScrollAndResize(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		tl := new(Timeline).Track("x", TweenSpec{To: 100, Duration: time.Second, Ease: Linear})
		binding := ScrollBinding{ID: "scroller", Horizontal: horizontal}
		now := time.Unix(2700, 0)
		extent := float32(500)
		var frame TimelineFrame
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			frame = tl.Sample(cx, TimelineOptions{Scroll: &binding})
			scroll := el.Div().ID("scroller").Size(el.Dp(100)).Scrollbars(el.ScrollbarAlways)
			child := el.Div().W(el.Dp(100)).H(el.Dp(extent)).NoShrink()
			if horizontal {
				scroll.ScrollX()
				child.W(el.Dp(extent)).H(el.Dp(100))
			} else {
				scroll.ScrollY()
			}
			return el.Div().Items(el.Start).Child(scroll.Child(child))
		}))
		root.RetainFrames()
		h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
		h.Frame()
		if frame.Active || frame.Progress != 0 {
			t.Fatalf("initial %+v", frame)
		}
		if horizontal {
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(50, 50), Scroll: f32.Pt(200, 0)})
			h.Frame()
		} else {
			h.Scroll(50, 50, 200)
		}
		h.Frame()
		if frame.Active || math.Abs(float64(frame.Progress-.5)) > .01 {
			t.Fatalf("scroll horizontal=%v: %+v", horizontal, frame)
		}
		now = now.Add(time.Hour)
		h.Frame()
		if frame.Active || math.Abs(float64(frame.Progress-.5)) > .01 {
			t.Fatal("clock advanced held scroll timeline")
		}
		core.Update(func() { extent = 900 })
		h.Frame()
		if _, ok := h.Router.WakeupTime(); !ok {
			t.Fatal("changed scroll extent did not schedule its dependent view")
		}
		h.Frame()
		if _, ok := h.Router.WakeupTime(); ok {
			t.Fatal("unchanged scroll extent requested another frame")
		}
		if math.Abs(float64(frame.Progress-.25)) > .01 {
			t.Fatalf("content resize not reflected: %+v", frame)
		}
		core.Update(func() { binding.Start, binding.End = 100, 300 })
		h.Frame()
		if math.Abs(float64(frame.Progress-.5)) > .01 {
			t.Fatal("explicit range not applied")
		}
		core.Update(func() { binding.ID = "missing" })
		h.Frame()
		if frame.Progress != 0 {
			t.Fatal("missing source did not reset")
		}
	}
}
