package kit

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestCarouselSmoothScrollDebouncesAndSnaps(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			car := Carousel(text("one"), text("two"), text("three")).Basis(.5).Height(240).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
			now := time.Unix(100, 0)
			var cx *el.Context
			root := el.Embed(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(240)).Child(car.Content().Render(c)) }))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
				gtx.Constraints.Max = image.Pt(int(300*scale), int(300*scale))
				root.Layout(gtx)
			})
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			scroll := func(d float32) {
				delta := f32.Pt(d*scale, 0)
				if vertical {
					delta = f32.Pt(0, d*scale)
				}
				h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(40*scale, 40*scale), Scroll: delta})
				h.Frame()
			}
			scroll(90)
			if !car.scroll.active || car.scroll.offset != 90 || car.Value() != 0 || calls != 0 {
				t.Fatal("continuous scroll", vertical, scale, car.scroll, car.Value(), calls)
			}
			now = now.Add(100 * time.Millisecond)
			h.Frame()
			scroll(5)
			now = now.Add(100 * time.Millisecond)
			h.Frame()
			if car.Value() != 0 || calls != 0 {
				t.Fatal("old idle timer fired")
			}
			now = now.Add(50 * time.Millisecond)
			h.Frame()
			h.Frame()
			if car.Value() != 1 || calls != 1 || car.scroll.active {
				t.Fatal("idle snap", car.Value(), calls, car.scroll)
			}
			now = now.Add(carouselTransition)
			h.Frame()
			id := autoID("carousel", car) + "/stage"
			offset, _, _ := cx.ScrollStateX(id)
			if vertical {
				offset, _, _ = cx.ScrollState(id)
			}
			if offset != 124 {
				t.Fatal("snap offset", offset)
			}
			scroll(-80)
			car.SetValue(2)
			h.Frame()
			now = now.Add(time.Second)
			h.Frame()
			if car.Value() != 2 || calls != 1 {
				t.Fatal("stale timer changed programmatic selection")
			}
			car.SetValue(0)
			car.WheelStep(true)
			h.Frame()
			scroll(1)
			if car.Value() != 1 || calls != 2 {
				t.Fatal("step event", car.Value(), calls)
			}
			car.Scrollable(false)
			h.Frame()
			scroll(80)
			if car.Value() != 1 || calls != 2 {
				t.Fatal("scroll disabled")
			}
		}
	}
}

func TestCarouselScrollBoundaryRoutesToParent(t *testing.T) {
	car := Carousel(text("one"), text("two"), text("three")).Vertical(true).Height(120).Loop(false)
	car.SetValue(2)
	c := &clock{now: time.Unix(100, 0)}
	var cx *el.Context
	h := c.harness(func(current *el.Context) el.Element {
		cx = current
		return el.Div().ID("outer").W(el.Dp(240)).H(el.Dp(200)).ScrollY().Child(car.Content().Render(current), el.Div().H(el.Dp(800)))
	})
	for i := 0; i < 4; i++ {
		h.Frame()
	}
	h.Scroll(40, 40, 60)
	offset, _, _ := cx.ScrollState("outer")
	if offset != 60 || car.scroll.active || car.Value() != 2 {
		t.Fatal("edge swallowed outer scroll", offset, car.scroll)
	}
	cx.ScrollTo("outer", 0)
	car.SetValue(0)
	h.Frame()
	h.Frame()
	h.Scroll(40, 40, 100)
	h.Scroll(40, 40, 300)
	offset, _, _ = cx.ScrollState("outer")
	if offset != 0 || !car.scroll.active {
		t.Fatal("active gesture leaked to parent", offset, car.scroll)
	}
	c.advance(h, 150*time.Millisecond)
	h.Frame()
	if car.Value() != 2 {
		t.Fatal("end snap", car.Value())
	}
	h.Scroll(40, 40, 20)
	offset, _, _ = cx.ScrollState("outer")
	if offset != 20 {
		t.Fatal("next boundary gesture not transferred", offset)
	}
}
