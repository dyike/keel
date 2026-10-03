package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestSwitchTabOrderAndPointerFocus(t *testing.T) {
	a := kit.Switch("A", false).TabIndex(2)
	b := kit.Switch("B", false).TabStop(false)
	c := kit.Switch("C", false).TabIndex(1)
	var cx *el.Context
	w := openTest(t, Options{Width: 300, Height: 220, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Items(el.Start).Child(a.Render(ctx), b.Render(ctx), c.Render(ctx))
	}))})
	press := func(k string) {
		t.Helper()
		if err := w.press(k); err != nil {
			t.Fatal(err)
		}
	}
	assertFocus := func(s *kit.SwitchView) {
		t.Helper()
		w.render()
		if !cx.Focused(s.FocusID()) {
			t.Fatal("wrong focus", s.FocusID())
		}
	}
	press("tab")
	assertFocus(c)
	press("tab")
	assertFocus(a)
	press("tab")
	assertFocus(c)
	press("shift+tab")
	assertFocus(a)
	e := element(t, w, "B")
	w.click(e.center())
	assertFocus(b)
	before := b.Value()
	press("space")
	if b.Value() == before {
		t.Fatal("skipped control cannot activate")
	}
	press("tab")
	assertFocus(c)
	cx.Focus(b.FocusID())
	w.render()
	assertFocus(b)
	b.TabStop(true).TabIndex(0)
	w.render()
	press("tab")
	assertFocus(c)
	a.SetDisabled(true)
	w.render()
	press("tab")
	assertFocus(b)
	c.TabIndex(-1)
	w.render()
	press("tab")
	assertFocus(b)
}

func TestSwitchTabInsideModal(t *testing.T) {
	a := kit.Switch("A", false).TabIndex(2)
	b := kit.Switch("B", false).TabStop(false)
	c := kit.Switch("C", false).TabIndex(1)
	outside := kit.Switch("Outside", false).TabIndex(0)
	open := true
	var cx *el.Context
	w := openTest(t, Options{Width: 400, Height: 260, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		if open {
			ctx.Overlay("modal", el.Modal(el.Div().Child(a.Render(ctx), b.Render(ctx), c.Render(ctx))).OnDismiss(func() { open = false }))
		}
		return outside.Render(ctx)
	}))})
	w.render()
	w.render()
	if !cx.Focused(c.FocusID()) {
		t.Fatal("modal did not start at first ordered stop")
	}
	for _, want := range []*kit.SwitchView{a, c, a, c} {
		if err := w.press("tab"); err != nil {
			t.Fatal(err)
		}
		w.render()
		if !cx.Focused(want.FocusID()) {
			t.Fatal("modal tab escaped or wrong order")
		}
	}
	if err := w.press("esc"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Fatal("modal did not close")
	}
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	w.render()
	if !cx.Focused(outside.FocusID()) {
		t.Fatal("closed modal stops retained")
	}
}

func TestCustomTabOrderIncludesInputsAndSkipsHidden(t *testing.T) {
	first := kit.Input("Input")
	second := kit.Switch("Second", false).TabIndex(0)
	last := kit.Switch("Last", false).TabIndex(0)
	hidden := false
	var cx *el.Context
	w := openTest(t, Options{Width: 300, Height: 220, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(first.Render(ctx), el.Div().Hidden(hidden).Child(second.Render(ctx)), last.Render(ctx))
	}))})
	focus := func(id string) {
		t.Helper()
		if err := w.press("tab"); err != nil {
			t.Fatal(err)
		}
		w.render()
		if !cx.Focused(id) {
			t.Fatal("wrong equal-index order", id)
		}
	}
	focus(first.FocusID())
	focus(second.FocusID())
	focus(last.FocusID())
	hidden = true
	w.render()
	focus(first.FocusID())
	focus(last.FocusID())
}
