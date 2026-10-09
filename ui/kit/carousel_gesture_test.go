package kit

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestCarouselFollowsPlatformScrollGestures(t *testing.T) {
	defer core.ReportScrollGesture(core.ScrollDeviceUnknown, false, false, false)
	for _, vertical := range []bool{false, true} {
		car := Carousel(text("one"), text("two"), text("three")).Basis(.5).Height(240).Vertical(vertical).Loop(false)
		now := time.Unix(100, 0)
		root := el.Embed(el.ViewFunc(func(c *el.Context) el.Element { return el.Div().W(el.Dp(240)).Child(car.Content().Render(c)) }))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Now = now
			gtx.Constraints.Max = image.Pt(300, 300)
			root.Layout(gtx)
		})
		for i := 0; i < 4; i++ {
			h.Frame()
		}
		scroll := func(d float32) {
			delta := f32.Pt(d, 0)
			if vertical {
				delta = f32.Pt(0, d)
			}
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(40, 40), Scroll: delta})
			h.Frame()
		}

		// Fingers down: no snap however long the pause.
		core.ReportScrollGesture(core.ScrollDeviceTrackpad, true, false, false)
		scroll(90)
		now = now.Add(900 * time.Millisecond)
		h.Frame()
		h.Frame()
		if !car.scroll.active || car.Value() != 0 {
			t.Fatal("snapped while fingers down", vertical, car.Value())
		}
		// Lift: snap right away.
		core.ReportScrollGesture(core.ScrollDeviceTrackpad, false, false, true)
		h.Frame()
		h.Frame()
		if car.scroll.active || car.Value() != 1 {
			t.Fatal("lift did not snap", vertical, car.Value())
		}
		// Momentum after the lift is ignored.
		core.ReportScrollGesture(core.ScrollDeviceTrackpad, false, true, false)
		scroll(200)
		if car.scroll.active || car.Value() != 1 {
			t.Fatal("momentum moved the carousel", vertical)
		}
		now = now.Add(carouselTransition)
		h.Frame()

		// A wheel steps one item per notch along the axis.
		core.ReportScrollGesture(core.ScrollDeviceWheel, false, false, false)
		scroll(10)
		if car.Value() != 2 || car.scroll.active {
			t.Fatal("wheel notch", vertical, car.Value())
		}
		scroll(-10)
		if car.Value() != 1 {
			t.Fatal("wheel back", vertical, car.Value())
		}
		core.ReportScrollGesture(core.ScrollDeviceUnknown, false, false, false)
	}
}
