package kit

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestCarouselPointerDragSnapsOnRelease(t *testing.T) {
	for _, source := range []pointer.Source{pointer.Mouse, pointer.Touch} {
		for _, scale := range []int{1, 2} {
			for _, vertical := range []bool{false, true} {
				calls := 0
				car := Carousel(text("one"), text("two"), text("three")).Height(200).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
				var cx *el.Context
				h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return car.Content().Render(c) }), 200, scale)
				for i := 0; i < 4; i++ {
					h.Frame()
				}
				point := func(along float32) f32.Point {
					if vertical {
						return f32.Pt(60*float32(scale), along*float32(scale))
					}
					return f32.Pt(along*float32(scale), 60*float32(scale))
				}
				press := func() {
					h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: source, Buttons: pointer.ButtonPrimary, Position: point(170)})
					h.Frame()
				}
				move := func() {
					h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: source, Buttons: pointer.ButtonPrimary, Position: point(30)})
					h.Frame()
				}
				release := func() {
					h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: source, Position: point(30)})
					h.Frame()
					h.Frame()
				}
				press()
				move()
				id := autoID("carousel", car) + "/stage"
				off, _, _ := cx.ScrollStateX(id)
				if vertical {
					off, _, _ = cx.ScrollState(id)
				}
				if off != 140 || calls != 0 || car.Value() != 0 {
					t.Fatal("live drag", vertical, scale, off, calls, car.Value(), car.drag)
				}
				release()
				if car.Value() != 1 || calls != 1 || car.drag.active {
					t.Fatal("release did not snap", vertical, scale, car.Value(), calls, car.drag)
				}
				// Cancellation and programmatic navigation must not commit stale drags.
				press()
				move()
				h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: source})
				h.Frame()
				h.Frame()
				if car.Value() != 1 || calls != 1 || car.drag.active {
					t.Fatal("cancel selected a page", car.Value(), calls, car.drag)
				}
				press()
				move()
				car.SetValue(0)
				release()
				if car.Value() != 0 || calls != 1 {
					t.Fatal("stale release overwrote programmatic selection")
				}
				press()
				move()
				car.SetDisabled(true)
				release()
				if car.Value() != 0 || calls != 1 || car.drag.active {
					t.Fatal("disable committed drag")
				}
				car.SetDisabled(false)
				car.Draggable(false)
				h.Frame()
				press()
				move()
				release()
				if car.Value() != 0 || calls != 1 {
					t.Fatal("drag disabled")
				}
			}
		}
	}
}

func TestCarouselDragDoesNotStealChildButton(t *testing.T) {
	clicks := 0
	car := Carousel(Button("Action", func() { clicks++ }), text("second"))
	h := renderView(car.Content(), 240, 1)
	h.Frame()
	h.Frame()
	click(t, h, "Action")
	if clicks != 1 || car.Value() != 0 || car.drag.active {
		t.Fatal("child click intercepted", clicks, car.Value(), car.drag)
	}
}
