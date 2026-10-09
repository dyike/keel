package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
	"time"
)

func TestCarouselVerticalNavigationAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		changes := 0
		car := Carousel(text("one"), text("two"), text("three")).Height(160).Vertical(true).OnChange(func(int) { changes++ })
		var cx *el.Context
		h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return car.Render(c) }), 180, scale)
		prev, next := bounds(h, locale.Current().PrevSlide), bounds(h, locale.Current().NextSlide)
		if prev.Min.X != next.Min.X || prev.Max.Y >= next.Min.Y || next.Max.X > 180*scale {
			t.Fatal("vertical controls", prev, next)
		}
		h.Router.MoveFocus(key.FocusForward)
		h.Frame()
		h.Key(key.NameDownArrow, 0)
		if car.Value() != 1 || changes != 1 {
			t.Fatal("down navigation", car.Value(), changes)
		}
		h.Key(key.NameRightArrow, 0)
		if car.Value() != 1 {
			t.Fatal("inactive axis handled")
		}
		h.Key(key.NameEnd, 0)
		if car.Value() != 2 {
			t.Fatal("end ignored")
		}
		h.Key(key.NameHome, 0)
		h.Key(key.NameUpArrow, 0)
		if car.Value() != 2 {
			t.Fatal("home/up wrapping")
		}
		click(t, h, locale.Current().NextSlide)
		if car.Value() != 0 {
			t.Fatal("next did not wrap")
		}
		cx.Focus(autoID("carousel", car) + "/next")
		h.Frame()
		h.Frame()
		car.Vertical(false)
		h.Frame()
		h.Key(key.NameSpace, 0)
		if car.Value() != 1 {
			t.Fatal("orientation lost button focus", car.Value())
		}
		prev, next = bounds(h, locale.Current().PrevSlide), bounds(h, locale.Current().NextSlide)
		if prev.Min.Y != next.Min.Y || prev.Max.X >= next.Min.X {
			t.Fatal("horizontal controls not restored", prev, next)
		}
		car.Vertical(true).Height(float32(math.Inf(1))).Height(float32(math.NaN()))
		h.Frame()
		if car.height != 160 {
			t.Fatal("nonfinite height accepted")
		}
		car.SetDisabled(true)
		h.Frame()
		click(t, h, locale.Current().NextSlide)
		if car.Value() != 1 {
			t.Fatal("disabled vertical navigation")
		}
	}
}

func TestCarouselVerticalAutoplayPause(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	car := Carousel(text("one"), text("two")).Height(120).Vertical(true).Autoplay(time.Second)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(250)).Child(car.Render(cx)) })
	h.Move(390, 290)
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 1 {
		t.Fatal("vertical autoplay")
	}
	h.Move(20, 20)
	h.Frame()
	c.advance(h, 2*time.Second)
	h.Frame()
	if car.Value() != 1 {
		t.Fatal("hover did not pause")
	}
	car.Vertical(false)
	h.Frame()
	h.Move(390, 290)
	h.Frame()
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 0 {
		t.Fatal("orientation interrupted resumed autoplay")
	}
}
