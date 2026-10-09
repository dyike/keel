package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestCarouselTouchDragBoundaryHandoff(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, tc := range []struct {
			name           string
			vertical, loop bool
			selected       int
			delta          float32
			outer          bool
		}{
			{"end outward", true, false, 2, -30, true},
			{"start outward", true, false, 0, 30, true},
			{"inward", true, false, 0, -30, false},
			{"loop", true, true, 2, -30, false},
			{"accepted drag reaches end", true, false, 1, -100, false},
			{"cross axis", false, true, 0, -30, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				car := Carousel(text("one"), text("two"), text("three")).Height(160).Gap(8).Vertical(tc.vertical).Loop(tc.loop)
				car.SetValue(tc.selected)
				var cx *el.Context
				h := renderView(el.ViewFunc(func(c *el.Context) el.Element {
					cx = c
					return el.Div().Child(el.Div().ID("outer").W(el.Dp(200)).H(el.Dp(200)).ScrollY().Child(el.Div().H(el.Dp(100)).NoShrink(), car.Content().Render(c), el.Div().H(el.Dp(700)).NoShrink()))
				}), 240, scale)
				for i := 0; i < 4; i++ {
					h.Frame()
				}
				cx.ScrollTo("outer", 100)
				h.Frame()
				h.Frame()
				send := func(kind pointer.Kind, y float32) {
					h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Touch, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(50*float32(scale), y*float32(scale))})
					h.Frame()
				}
				send(pointer.Press, 80)
				send(pointer.Move, 80+tc.delta)
				send(pointer.Move, 80+2*tc.delta)
				h.Frame()
				off, _, _ := cx.ScrollState("outer")
				if tc.outer {
					if off == 100 || car.drag.active || car.Value() != tc.selected {
						t.Fatal("parent did not take over", scale, off, car.drag, car.Value())
					}
				} else if off != 100 || !car.drag.active || !car.drag.moved {
					t.Fatal("carousel lost accepted gesture", scale, off, car.drag)
				}
				send(pointer.Release, 80+2*tc.delta)
			})
		}
	}
}

func TestCarouselDragLeavesInteractiveSlideInControl(t *testing.T) {
	for _, source := range []pointer.Source{pointer.Mouse, pointer.Touch} {
		slider := Slider("Slide volume", 0, 100).Step(1)
		car := Carousel(slider, text("two")).Height(160)
		h := renderView(car.Content(), 240, 1)
		for i := 0; i < 4; i++ {
			h.Frame()
		}
		n, ok := semanticNode(h, "slider:0")
		if !ok {
			t.Fatal("missing slider")
		}
		r := n.Desc.Bounds
		point := func(x int) f32.Point { return f32.Pt(float32(x), float32(r.Min.Y+r.Dy()/2)) }
		send := func(kind pointer.Kind, x int) {
			h.Router.Queue(pointer.Event{Kind: kind, Source: source, Buttons: pointer.ButtonPrimary, Position: point(x)})
			h.Frame()
		}
		send(pointer.Press, r.Min.X+20)
		send(pointer.Move, r.Max.X-30)
		send(pointer.Move, r.Max.X-20)
		send(pointer.Release, r.Max.X-20)
		if slider.Value() < 80 || car.Value() != 0 || car.drag.active {
			t.Fatal("carousel stole slider drag", source, slider.Value(), car.Value(), car.drag)
		}
	}
}
