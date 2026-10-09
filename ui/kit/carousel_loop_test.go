package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
	"time"
)

func TestCarouselLoopBoundariesAndPublicNavigation(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		calls := 0
		car := Carousel(text("one"), text("two"), text("three")).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
		var cx *el.Context
		h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return car.Render(c) }), 240, 1)
		if car.CanPrevious() || !car.CanNext() {
			t.Fatal("initial boundary")
		}
		click(t, h, locale.Current().PrevSlide)
		car.Previous()
		if calls != 0 || car.Value() != 0 {
			t.Fatal("previous crossed start")
		}
		car.Next()
		h.Frame()
		if car.Value() != 1 || calls != 1 || !car.CanPrevious() || !car.CanNext() {
			t.Fatal("public next")
		}
		click(t, h, locale.Current().NextSlide)
		if car.Value() != 2 || calls != 2 || car.CanNext() {
			t.Fatal("end boundary")
		}
		car.Next()
		click(t, h, locale.Current().NextSlide)
		if calls != 2 {
			t.Fatal("boundary emits duplicate change")
		}
		cx.Focus(autoID("carousel", car))
		h.Frame()
		h.Frame()
		forward, back := key.NameRightArrow, key.NameLeftArrow
		if vertical {
			forward, back = key.NameDownArrow, key.NameUpArrow
		}
		h.Key(forward, 0)
		if car.Value() != 2 || calls != 2 {
			t.Fatal("key crossed end")
		}
		h.Key(back, 0)
		if car.Value() != 1 || calls != 3 {
			t.Fatal("key back failed")
		}
		car.Loop(true)
		car.SetValue(2)
		car.Next()
		if car.Value() != 0 || calls != 4 {
			t.Fatal("loop restoration")
		}
		car.Previous()
		if car.Value() != 2 || calls != 5 {
			t.Fatal("loop previous")
		}
		car.SetDisabled(true)
		if car.CanNext() || car.CanPrevious() {
			t.Fatal("disabled availability")
		}
		car.Next()
		car.Previous()
		car.SetValue(1)
		if car.Value() != 1 || calls != 5 {
			t.Fatal("disabled API or programmatic selection")
		}
	}
	for _, car := range []*CarouselView{Carousel(), Carousel(text("only"))} {
		for _, loop := range []bool{false, true} {
			car.Loop(loop)
			car.Next()
			car.Previous()
			if car.Value() != 0 || car.CanNext() || car.CanPrevious() {
				t.Fatal("empty/single boundary")
			}
		}
	}
}

func TestCarouselNonLoopAutoplayStopsAndRestarts(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	calls := 0
	car := Carousel(text("one"), text("two")).Loop(false).Autoplay(time.Second).OnChange(func(int) { calls++ })
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(250)).Child(car.Render(cx)) })
	h.Move(390, 290)
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 1 || calls != 1 {
		t.Fatal("autoplay did not reach end")
	}
	c.advance(h, 10*time.Second)
	h.Frame()
	if car.Value() != 1 || calls != 1 {
		t.Fatal("autoplay crossed end")
	}
	car.SetValue(0)
	h.Frame()
	c.advance(h, 500*time.Millisecond)
	h.Frame()
	if car.Value() != 0 {
		t.Fatal("old timer resumed immediately")
	}
	c.advance(h, 500*time.Millisecond)
	h.Frame()
	if car.Value() != 1 || calls != 2 {
		t.Fatal("autoplay did not resume")
	}
	car.Loop(true)
	h.Frame()
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 0 || calls != 3 {
		t.Fatal("loop autoplay not restored")
	}
}
