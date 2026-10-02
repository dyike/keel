package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestDropdownDisabledCannotReopenExternalMenu(t *testing.T) {
	for _, split := range []bool{false, true} {
		calls := 0
		m := Menu().Item("Choice", "", func() { calls++ })
		d := DropdownButton("Open", m)
		if split {
			d.Split(func() { calls++ })
		}
		parentDisabled := false
		h := page(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(parentDisabled).Child(d.Render(cx)) }))
		m.SetValue(true)
		h.Frame()
		h.Frame()
		if !m.Value() {
			t.Fatal("initial external open")
		}
		parentDisabled = true
		h.Frame()
		h.Frame()
		if m.Value() {
			t.Fatal("ancestor disabled menu remained open")
		}
		parentDisabled = false
		d.SetDisabled(true)
		m.SetValue(true)
		h.Frame()
		h.Frame()
		if m.Value() || shown(h, "Choice") {
			t.Fatal("external open bypassed disabled dropdown")
		}
		click(t, h, "Open")
		if calls != 0 {
			t.Fatal("disabled main action")
		}
	}
}
func TestDropdownNilMenuAllowsSplitAction(t *testing.T) {
	calls := 0
	d := DropdownButton("Run", nil).Split(func() { calls++ })
	h := page(d)
	click(t, h, "Run")
	if calls != 1 {
		t.Fatal("nil-menu split action")
	}
	d.SetDisabled(true)
	h.Frame()
}
