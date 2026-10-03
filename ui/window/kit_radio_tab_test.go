package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestRadioGroupTabConfiguration(t *testing.T) {
	a := kit.RadioGroup("Plan", "Basic", "Pro").TabIndex(2)
	b := kit.Checkbox("Accept", false).TabIndex(1)
	var cx *el.Context
	w := openTest(t, Options{Width: 300, Height: 300, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(a.Render(ctx), b.Render(ctx))
	}))})
	press := func(k string, id string) {
		t.Helper()
		if err := w.press(k); err != nil {
			t.Fatal(err)
		}
		w.render()
		if !cx.Focused(id) {
			t.Fatal("wrong focus after", k, id)
		}
	}
	press("tab", b.FocusID())
	press("tab", a.FocusID())
	if err := w.press("right"); err != nil {
		t.Fatal(err)
	}
	w.render()
	if a.Value() != "Pro" {
		t.Fatal("arrow selection")
	}
	press("tab", b.FocusID())
	press("shift+tab", a.FocusID())
	a.TabStop(false)
	w.render()
	press("tab", b.FocusID())
	press("tab", b.FocusID())
	w.click(element(t, w, "Basic").center())
	if a.Value() != "Basic" {
		t.Fatal("skipped group cannot be clicked")
	}
	a.TabStop(true).TabIndex(-1)
	w.render()
	press("tab", b.FocusID())
	a.TabIndex(0)
	w.render()
	press("tab", a.FocusID())
}

func TestRadioIndividualTabOverride(t *testing.T) {
	a := kit.RadioGroup("Plan", "Basic", "Pro").TabStop(false).ItemTab("Pro", true, 1)
	b := kit.Checkbox("Accept", false).TabIndex(2)
	a.SetValue("Pro")
	proID := a.FocusID()
	a.SetValue("Basic")
	var cx *el.Context
	w := openTest(t, Options{Width: 300, Height: 300, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(a.Item("Basic").Render(ctx), a.Item("Pro").Render(ctx), b.Render(ctx))
	}))})
	tab := func(id string) {
		t.Helper()
		if err := w.press("tab"); err != nil {
			t.Fatal(err)
		}
		w.render()
		if !cx.Focused(id) {
			t.Fatal("unexpected target", id)
		}
	}
	tab(proID)
	if a.Value() != "Basic" {
		t.Fatal("Tab changed selection")
	}
	tab(b.FocusID())
	a.SetOptionDisabled("Pro", true)
	w.render()
	tab(b.FocusID())
	a.SetOptionDisabled("Pro", false)
	a.ClearItemTab("Pro")
	w.render()
	tab(b.FocusID())
	a.ItemTab("Pro", true, 1)
	a.SetOptions("Basic")
	a.SetOptions("Basic", "Pro")
	w.render()
	tab(b.FocusID())
}
