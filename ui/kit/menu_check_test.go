package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMenuCheckStateKeyboardAndDisabled(t *testing.T) {
	calls := 0
	m := Menu().CheckItem("Show", "", false, func(on bool) {
		calls++
		if !on {
			t.Fatal("wrong callback state")
		}
	}).IconItem("Copy", "", IconCopy, nil)
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	h.Frame()
	n, ok := node(h, "Show")
	if !ok || n.Desc.Description != "menuitemcheckbox" || n.Desc.Selected {
		t.Fatal("unchecked semantics", n.Desc)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	on, ok := m.ItemChecked("Show")
	if !ok || !on || calls != 1 || m.Value() {
		t.Fatal("keyboard toggle/close")
	}
	click(t, h, "Open")
	h.Frame()
	n, _ = node(h, "Show")
	if !n.Desc.Selected {
		t.Fatal("reopen lost state")
	}
	m.SetItemDisabled("Show", true)
	h.Frame()
	click(t, h, "Show")
	if calls != 1 || !m.Value() {
		t.Fatal("disabled check activated")
	}
	m.SetItemChecked("Show", false)
	h.Frame()
	if on, _ = m.ItemChecked("Show"); on || calls != 1 {
		t.Fatal("programmatic update fired callback")
	}
	m.SetItemIcon("Show", IconSettings)
	m.CheckSide(el.Right)
	h.Frame()
	if !m.Value() {
		t.Fatal("style update closed menu")
	}
}

func TestMenuSubCheckClosesChain(t *testing.T) {
	calls := 0
	sub := Menu().CheckItem("Enabled", "", true, func(on bool) {
		calls++
		if on {
			t.Fatal("did not uncheck")
		}
	})
	m := Menu().Sub("Options", sub)
	m.SetItemIcon("Options", IconSettings)
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	click(t, h, "Options")
	h.Frame()
	click(t, h, "Enabled")
	h.Frame()
	if m.Value() || calls != 1 {
		t.Fatal("nested check did not close root")
	}
}
