package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestCarouselBasisPreservesInputFocus(t *testing.T) {
	input := kit.Input("Edit slide")
	car := kit.Carousel(input, labelView("Second"), labelView("Third")).Basis(.6).ItemBasis(1, .3).Height(220)
	w := openTest(t, kitPage(car.Content()))
	w.click(element(t, w, "Edit slide").center())
	w.typeText("hello")
	if input.Value() != "hello" {
		t.Fatal("initial input", input.Value())
	}
	car.ItemBasis(0, .75).Basis(.5)
	w.snapshot()
	w.press("Right")
	w.press("Backspace")
	if input.Value() != "ello" {
		t.Fatal("mixed resize lost input or focus", input.Value())
	}
	car.Vertical(true)
	w.snapshot()
	w.typeText("x")
	if len(input.Value()) != 5 {
		t.Fatal("axis change lost focus", input.Value())
	}
}
