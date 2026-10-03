package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestResizableVisibilityPreservesInput(t *testing.T) {
	first, second := kit.Input("First"), kit.Input("Second")
	split := kit.Resizable(first, second).Min(40, 40)
	split.SetValue(180)
	w := openTest(t, Options{Width: 400, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(400)).H(el.Dp(200)).Items(el.Stretch).Child(split.Render(cx))
	}))})
	w.snapshot()
	w.snapshot()
	w.click(element(t, w, "First").center())
	w.typeText("saved")
	split.Visible(false, true)
	for _, e := range w.snapshot() {
		if e.Name == "First" || e.Role == "separator" {
			t.Fatal("hidden input/handle exposed", e)
		}
	}
	w.typeText("hidden")
	if first.Value() != "saved" {
		t.Fatal("hidden field accepted input", first.Value())
	}
	w.click(element(t, w, "Second").center())
	w.typeText("other")
	split.Visible(true, true)
	w.snapshot()
	if first.Value() != "saved" || second.Value() != "other" || split.Value() != 180 {
		t.Fatal("state lost", first.Value(), second.Value(), split.Value())
	}
	w.typeText("!")
	if second.Value() != "!other" {
		t.Fatal("visible pane lost focus", second.Value())
	}
	split.Visible(false, false)
	if len(w.snapshot()) != 0 {
		for _, e := range w.snapshot() {
			if e.Role == "textbox" || e.Role == "separator" {
				t.Fatal("both hidden", e)
			}
		}
	}
}
