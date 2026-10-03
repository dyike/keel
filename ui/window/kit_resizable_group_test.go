package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestResizableGroupStableContent(t *testing.T) {
	first, second, third := kit.Input("First"), kit.Input("Second"), kit.Input("Third")
	a := kit.ResizablePanel{ID: "a", Content: first, Size: 120, Min: 60}
	b := kit.ResizablePanel{ID: "b", Content: second, Size: 140, Min: 60}
	c := kit.ResizablePanel{ID: "c", Content: third, Size: 160, Min: 60}
	g := kit.ResizableGroup(a, b, c)
	w := openTest(t, Options{Width: 500, Height: 240, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(500)).H(el.Dp(240)).Items(el.Stretch).Child(g.Render(cx))
	}))})
	w.snapshot()
	w.snapshot()
	w.click(element(t, w, "Second").center())
	w.typeText("retained")
	g.SetVisible("a", false)
	w.snapshot()
	w.typeText("!")
	if second.Value() != "!retained" {
		t.Fatal("hide lost surviving focus", second.Value())
	}
	count := 0
	for _, e := range w.snapshot() {
		if e.Name == "First" {
			t.Fatal("hidden content visible")
		}
		if e.Role == "separator" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("hidden panel handle count", count)
	}
	g.SetPanels(c, b, a)
	w.snapshot()
	w.typeText("?")
	if second.Value() != "?!retained" {
		t.Fatal("reorder lost focus", second.Value())
	}
	g.SetDisabled(true)
	w.snapshot()
	w.typeText("no")
	if second.Value() != "?!retained" {
		t.Fatal("disabled accepted input")
	}
	g.SetPanels(b)
	w.snapshot()
	w.snapshot()
	for _, e := range w.snapshot() {
		if e.Role == "separator" {
			t.Fatal("single pane has handle")
		}
	}
	if second.Value() != "?!retained" {
		t.Fatal("single pane lost state")
	}
}
