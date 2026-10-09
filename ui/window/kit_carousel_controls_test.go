package window

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestCarouselComposedControls(t *testing.T) {
	calls, originalCalls := 0, 0
	car := kit.Carousel(labelView("First slide"), labelView("Second slide"), labelView("Third slide")).Loop(false).OnChange(func(int) { calls++ })
	source := kit.Button("Forward", func() { originalCalls++ }).Variant(kit.ButtonSecondary).Size(40)
	next, prev := car.NextControl(source), car.PreviousControl(kit.Button("Back", nil))
	page := car.PaginationItem(2, kit.Button("Third page", nil))
	invalid := car.PaginationItem(5, kit.Button("Invalid page", nil))
	content := car.Content()
	parentDisabled := false
	w := openTest(t, kitPage(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Disabled(parentDisabled).Items(el.Stretch).Child(content.Render(cx), el.Div().Row().Child(prev.Render(cx), next.Render(cx), page.Render(cx), invalid.Render(cx)))
	})))
	if !element(t, w, "Back").Disabled || element(t, w, "Forward").Disabled || !element(t, w, "Invalid page").Disabled {
		t.Fatal("initial boundary semantics")
	}
	w.click(element(t, w, "Forward").center())
	if car.Value() != 1 || calls != 1 || originalCalls != 0 {
		t.Fatal("composed next callback")
	}
	w.click(element(t, w, "Third page").center())
	if car.Value() != 2 || calls != 2 || !element(t, w, "Forward").Disabled {
		t.Fatal("page selection")
	}
	if e := element(t, w, "Third page"); e.Selected == nil || !*e.Selected {
		t.Fatal("selected page semantics", e)
	}
	w.press("Space")
	if calls != 2 {
		t.Fatal("same page emitted change")
	}
	car.SetValue(0)
	car.Vertical(true)
	w.snapshot()
	w.press("Space")
	if car.Value() != 2 || calls != 3 {
		t.Fatal("pagination lost focus across programmatic update/orientation")
	}
	car.SetValue(0)
	source.SetLoading(true)
	w.click(element(t, w, "Forward").center())
	if car.Value() != 0 {
		t.Fatal("loading control activated")
	}
	source.SetLoading(false)
	parentDisabled = true
	if !element(t, w, "Forward").Disabled || !element(t, w, "Third page").Disabled {
		t.Fatal("parent disabled not inherited")
	}
	w.click(element(t, w, "Third page").center())
	if car.Value() != 0 {
		t.Fatal("disabled parent activated page")
	}
	parentDisabled = false
	car.SetDisabled(true)
	if !element(t, w, "Forward").Disabled || !element(t, w, "Third page").Disabled {
		t.Fatal("shared disabled state")
	}
	car.SetDisabled(false)
	w.click(element(t, w, "Forward").center())
	if car.Value() != 1 || calls != 4 {
		t.Fatal("restore")
	}
}

func labelView(s string) el.View {
	return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) })
}

func TestCarouselContentKeyboardWithoutControls(t *testing.T) {
	car := kit.Carousel(labelView("First"), labelView("Last")).Loop(false)
	w := openTest(t, kitPage(car.Content()))
	w.snapshot()
	w.press("Tab")
	if err := w.press(string(key.NameEnd)); err != nil {
		t.Fatal(err)
	}
	if car.Value() != 1 {
		t.Fatal("content has no keyboard navigation", w.snapshot())
	}
	car.Vertical(true)
	w.snapshot()
	w.press("Up")
	if car.Value() != 0 {
		t.Fatal("vertical content keyboard navigation")
	}
	for _, e := range w.snapshot() {
		if e.Role == "button" {
			t.Fatal("built-in controls visible", e)
		}
	}
}
