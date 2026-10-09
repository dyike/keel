package kit

import (
	"gioui.org/io/key"
	"reflect"
	"testing"
)

func TestComboboxNonsearchableKeyboardAndClick(t *testing.T) {
	c := Combobox("Choice", "a", "b", "c").Searchable(false).DisableOption("b", true).AllowCustom()
	c.SetValue("a")
	h := page(c)
	if _, ok := node(h, "Choice"); !ok {
		t.Fatal("missing trigger")
	}
	clickClass(t, h, "Button", "Choice")
	h.Frame()
	h.Type("x")
	h.Frame()
	if c.text != "a" || !reflect.DeepEqual(c.matches(), []string{"a", "b", "c"}) {
		t.Fatal("typing changed static trigger", c.text, c.matches())
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() != "c" || c.open {
		t.Fatal("static navigation/confirm", c.Value(), c.open)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !c.open {
		t.Fatal("Enter did not reopen")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if c.open {
		t.Fatal("Esc did not close")
	}
}

func TestComboboxSearchModeSwitchAndAsync(t *testing.T) {
	var queries []string
	var token uint64
	c := Combobox("Choice").OnSearch(func(q string, t uint64) { queries = append(queries, q); token = t })
	c.SetValue("selected")
	c.text = "draft"
	c.open = true
	c.searchChanged()
	old := token
	c.Searchable(false)
	if c.open || c.text != "selected" || c.Value() != "selected" || c.SetResults(old, "late") {
		t.Fatal("switch did not cancel draft/request")
	}
	h := page(c)
	clickClass(t, h, "Button", "Choice")
	h.Frame()
	if queries[len(queries)-1] != "" {
		t.Fatal("static mode searched selection", queries)
	}
	c.SetResults(token, "one", "two")
	c.Searchable(true)
	h.Frame()
	clickClass(t, h, "Editor", "Choice")
	h.Type("x")
	h.Frame()
	if !c.open || queries[len(queries)-1] == "" {
		t.Fatal("search editor not restored")
	}
}

func TestComboboxNonsearchableMultiple(t *testing.T) {
	c := Combobox("Tags", "a", "b").Multiple().Searchable(false).Clearable(true)
	h := page(c)
	clickClass(t, h, "Button", "Tags")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"a"}) || !c.open {
		t.Fatal(c.Values(), c.open)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if len(c.Values()) != 0 || !c.open {
		t.Fatal("static toggle", c.Values(), c.open)
	}
	c.SetDisabled(true)
	h.Frame()
	clickClass(t, h, "Button", "Tags")
	if c.open {
		t.Fatal("disabled static trigger opened")
	}
}
