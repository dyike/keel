package kit

import (
	"gioui.org/io/key"
	"testing"
)

func TestComboboxDisabledOptionsKeyboardAndClick(t *testing.T) {
	c := Combobox("Choice", "a", "b", "c", "d").DisableOption("a", true).DisableOption("c", true)
	calls := 0
	c.OnChange(func(string) { calls++ })
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if c.active != 1 {
		t.Fatal("initial highlight", c.active)
	}
	click(t, h, "a")
	if c.Value() != "" || calls != 0 {
		t.Fatal("disabled click selected")
	}
	h.Key(key.NameDownArrow, 0)
	if c.active != 3 {
		t.Fatal("down did not skip", c.active)
	}
	h.Key(key.NameDownArrow, 0)
	if c.active != 1 {
		t.Fatal("wrap", c.active)
	}
	h.Key(key.NameUpArrow, 0)
	if c.active != 3 {
		t.Fatal("up wrap", c.active)
	}
	c.DisableOption("d", true)
	if c.active != 1 {
		t.Fatal("dynamic disable", c.active)
	}
	h.Frame()
	h.Key(key.NameReturn, 0)
	if c.Value() != "b" || calls != 1 {
		t.Fatal(c.Value(), calls)
	}
}

func TestComboboxDisabledDraftAndAllDisabled(t *testing.T) {
	c := Combobox("Choice", "a", "b").AllowCustom().DisableOption("a", true).DisableOption("b", true)
	c.SetValue("old")
	c.text = "a"
	c.settle()
	if c.Value() != "old" || c.text != "old" {
		t.Fatal("custom bypassed disabled")
	}
	c.choose("b")
	if c.Value() != "old" {
		t.Fatal("direct choose bypassed")
	}
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameUpArrow, 0)
	if c.active != -1 {
		t.Fatal("all disabled highlight", c.active)
	}
	h.Key(key.NameReturn, 0)
	if c.Value() != "old" {
		t.Fatal("all disabled selected")
	}
	c.DisableOption("a", false)
	c.text = "a"
	c.settle()
	if c.Value() != "a" {
		t.Fatal("reenable failed")
	}
}

func TestComboboxDisabledAsyncAndExistingSelection(t *testing.T) {
	var token uint64
	c := Combobox("Remote").Multiple().DisableOption("a", true).OnSearch(func(_ string, t uint64) { token = t })
	c.SetValues([]string{"a"})
	c.open = true
	c.searchChanged()
	if !c.SetResults(token, "a", "b") || c.active != 1 {
		t.Fatal("remote disabled highlight", c.active)
	}
	c.choose("a")
	if len(c.Values()) != 1 {
		t.Fatal("disabled existing choice changed")
	}
	c.removeValue("a")
	if len(c.Values()) != 0 {
		t.Fatal("cannot remove existing disabled value")
	}
	c.SetOptions("b", "a")
	if !c.disabledOptions["a"] {
		t.Fatal("replacement forgot disabled config")
	}
	c.SetValue("a")
	if c.Value() != "a" {
		t.Fatal("program assignment restricted")
	}
}

func TestComboboxDisabledPageNavigationAndFallback(t *testing.T) {
	c := Combobox("Choice", "a0", "a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10").DisableOption("a8", true).DisableOption("a0", true)
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NamePageDown, 0)
	if c.active < 0 || c.disabledOptions[c.matches()[c.active]] {
		t.Fatal("page selected disabled", c.active)
	}
	h.Key(key.NamePageUp, 0)
	if c.active < 0 || c.disabledOptions[c.matches()[c.active]] {
		t.Fatal("reverse page selected disabled", c.active)
	}
	c.text = "a"
	c.active = -1
	h.Frame()
	h.Key(key.NameReturn, 0)
	if c.Value() != "a1" {
		t.Fatal("fallback did not skip disabled first match", c.Value())
	}
}
