package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestCarouselBoundaryControlSemantics(t *testing.T) {
	text := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
	car := kit.Carousel(text("First"), text("Last")).Loop(false).Vertical(true)
	w := openTest(t, kitPage(car))
	prev, next := locale.Current().PrevSlide, locale.Current().NextSlide
	if !element(t, w, prev).Disabled || element(t, w, next).Disabled {
		t.Fatal("start disabled semantics")
	}
	w.click(element(t, w, prev).center())
	if car.Value() != 0 {
		t.Fatal("disabled previous activated")
	}
	w.click(element(t, w, next).center())
	if !element(t, w, next).Disabled || element(t, w, prev).Disabled {
		t.Fatal("end disabled semantics")
	}
	car.Loop(true)
	if element(t, w, next).Disabled {
		t.Fatal("loop did not restore button")
	}
	w.click(element(t, w, next).center())
	if car.Value() != 0 {
		t.Fatal("restored button did not wrap")
	}
	empty := openTest(t, kitPage(kit.Carousel()))
	if e := element(t, empty, "0/0"); e.Role != "group" {
		t.Fatal("empty count", e)
	}
	if !element(t, empty, prev).Disabled || !element(t, empty, next).Disabled {
		t.Fatal("empty enabled controls")
	}
}
