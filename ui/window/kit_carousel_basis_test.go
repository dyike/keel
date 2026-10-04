package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestCarouselBasisPreservesInputFocus(t *testing.T) {
	for _, fixed := range []bool{false, true} {
		input := kit.Input("Edit slide")
		car := kit.Carousel(input, labelView("Second"), labelView("Third")).Basis(.6).ItemBasis(1, .3).Height(220)
		w := openTest(t, kitPage(car.Content()))
		w.click(element(t, w, "Edit slide").center())
		w.typeText("hello")
		if input.Value() != "hello" {
			t.Fatal("initial input", input.Value())
		}
		car.ItemBasis(0, .75).Basis(.5)
		if fixed {
			car.ItemSize(0, 350)
		}
		w.snapshot()
		w.press("Right")
		w.press("Backspace")
		if input.Value() != "ello" {
			t.Fatal("resize lost input or focus", input.Value())
		}
		car.Vertical(true)
		w.snapshot()
		w.typeText("x")
		if len(input.Value()) != 5 {
			t.Fatal("axis change lost focus", input.Value())
		}
	}

}

func TestCarouselLoopRunwayPreservesInputFocus(t *testing.T) {
	input := kit.Input("Loop input")
	car := kit.Carousel(input, labelView("Second"), labelView("Third")).Height(220)
	w := openTest(t, kitPage(car.Content()))
	w.click(element(t, w, "Loop input").center())
	w.typeText("hello")
	car.Loop(false)
	w.snapshot()
	w.press("Right")
	w.press("Backspace")
	if input.Value() != "ello" {
		t.Fatal("removing runway lost focus", input.Value())
	}
	car.Loop(true)
	w.snapshot()
	w.typeText("x")
	if input.Value() != "xello" {
		t.Fatal("restoring runway lost focus", input.Value())
	}
	car.Previous()
	w.snapshot()
	car.Next()
	w.snapshot()
	if input.Value() != "xello" {
		t.Fatal("wrapping reset input state", input.Value())
	}
}
