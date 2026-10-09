package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestDropdownInnerButtonAndLoading(t *testing.T) {
	calls := 0
	inner := Button("Save", func() { calls++ }).Variant(ButtonSecondary).Size(44).Loading(true)
	menu := Menu().Item("Export", "", nil)
	d := DropdownButton("Fallback", menu).Button(inner)
	h := page(d)
	click(t, h, "Save")
	if calls != 0 {
		t.Fatal("loading action fired")
	}
	click(t, h, locale.Current().Name("Save", locale.Current().MoreOptions))
	if !menu.Value() {
		t.Fatal("loading main blocked arrow")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	d.Loading(false).Size(52).Variant(ButtonDanger)
	h.Frame()
	click(t, h, "Save")
	if calls != 1 {
		t.Fatal("inner handler not forwarded")
	}
	if inner.height != 44 || inner.variant != ButtonSecondary || !inner.loading || inner.id != "" {
		t.Fatal("render mutated source")
	}
	if bounds(h, "Save").Dy() < 52 {
		t.Fatal("size not forwarded")
	}
	d.Loading(true)
	h.Frame()
	h.Key(key.NameSpace, 0)
	h.Frame()
	if calls != 1 {
		t.Fatal("keyboard activated loading button")
	}
	d.Loading(false)
	h.Frame()
	h.Key(key.NameSpace, 0)
	h.Frame()
	if calls != 2 {
		t.Fatal("loading lost focus")
	}
	d.SetDisabled(true)
	h.Frame()
	click(t, h, locale.Current().Name("Save", locale.Current().MoreOptions))
	if menu.Value() {
		t.Fatal("disabled arrow opened")
	}
}

func TestDropdownPlainLoadingAndButtonReset(t *testing.T) {
	m := Menu().Item("Entry", "", nil)
	d := DropdownButton("Open", m).Loading(true)
	h := page(d)
	click(t, h, "Open")
	if m.Value() {
		t.Fatal("plain loading opened menu")
	}
	d.Loading(false)
	h.Frame()
	click(t, h, "Open")
	if !m.Value() {
		t.Fatal("plain loading reset")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	d.Button(Button("Custom", nil))
	h.Frame()
	if !shown(h, "Custom") {
		t.Fatal("inner button missing")
	}
	d.Button(nil)
	h.Frame()
	click(t, h, "Open")
	if !m.Value() {
		t.Fatal("button reset lost menu action")
	}
}
