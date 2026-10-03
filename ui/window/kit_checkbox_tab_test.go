package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestCheckboxTabConfiguration(t *testing.T) {
	a := kit.Checkbox("A", false).TabIndex(2)
	b := kit.Checkbox("B", false).TabStop(false)
	c := kit.Checkbox("C", false).TabIndex(1)
	var cx *el.Context
	w := openTest(t, Options{Width: 300, Height: 200, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(a.Render(ctx), b.Render(ctx), c.Render(ctx))
	}))})
	for _, want := range []*kit.CheckboxView{c, a, c} {
		if err := w.press("tab"); err != nil {
			t.Fatal(err)
		}
		w.render()
		if !cx.Focused(want.FocusID()) {
			t.Fatal("wrong tab target")
		}
	}
	if err := w.press("space"); err != nil {
		t.Fatal(err)
	}
	if !c.Value() {
		t.Fatal("keyboard toggle")
	}
	e := element(t, w, "B")
	w.click(e.center())
	if !b.Value() {
		t.Fatal("skipped checkbox not clickable")
	}
	b.TabStop(true).TabIndex(0)
	w.render()
	if err := w.press("shift+tab"); err != nil {
		t.Fatal(err)
	}
	w.render()
	if !cx.Focused(a.FocusID()) {
		t.Fatal("reverse traversal")
	}
}
