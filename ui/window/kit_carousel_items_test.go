package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestCarouselMultipleItemsClippingAndDisabledResize(t *testing.T) {
	calls := 0
	car := kit.Carousel(kit.Button("First action", func() { calls++ }), kit.Button("Second action", func() { calls++ }), kit.Button("Third action", func() { calls++ })).ItemsPerView(2).Height(160).Loop(false)
	width := float32(300)
	content := car.Content()
	w := openTest(t, kitPage(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(width)).Child(content.Render(cx)) })))
	visible := func(name string) bool {
		for _, e := range w.snapshot() {
			if e.Name == name {
				return true
			}
		}
		return false
	}
	if !visible("First action") || !visible("Second action") || visible("Third action") {
		t.Fatal("initial visible items", w.snapshot())
	}
	w.click(element(t, w, "Second action").center())
	if calls != 1 {
		t.Fatal("visible item not interactive")
	}
	car.SetValue(2)
	if visible("First action") || !visible("Third action") {
		t.Fatal("selected end visibility", w.snapshot())
	}
	car.SetDisabled(true)
	car.SetValue(0)
	width = 180
	car.Vertical(true)
	if !visible("First action") || !visible("Second action") || visible("Third action") {
		t.Fatal("disabled resized axis visibility", w.snapshot())
	}
	if !element(t, w, "First action").Disabled {
		t.Fatal("disabled item")
	}
	w.click(element(t, w, "First action").center())
	if calls != 1 {
		t.Fatal("disabled action ran")
	}
	car.SetDisabled(false)
	car.ItemsPerView(1)
	if !visible("First action") || visible("Second action") {
		t.Fatal("restore single item", w.snapshot())
	}
}
