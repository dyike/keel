package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestComboboxClearSelection(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		c := Combobox("Choice", "a", "b").Clearable(true)
		if multiple {
			c.Multiple()
		}
		c.SetValues([]string{"a", "b"})
		values, changes := 0, 0
		c.OnValuesChange(func(v []string) {
			values++
			if len(v) != 0 {
				t.Fatal(v)
			}
		})
		c.OnChange(func(v string) {
			changes++
			if v != "" {
				t.Fatal(v)
			}
		})
		h := page(c)
		c.SetError("old error")
		h.Frame()
		name := locale.Current().Name(locale.Current().Clear, "Choice")
		click(t, h, name)
		h.Frame()
		if len(c.Values()) != 0 || c.text != "" || c.open || c.Error() != "" || changes != 1 || shown(h, name) {
			t.Fatal("clear state", c.Values(), c.text, c.open, changes)
		}
		if (multiple && values != 1) || (!multiple && values != 0) {
			t.Fatal("values callback", values)
		}
		h.Key(key.NameDownArrow, 0)
		h.Frame()
		if !c.open {
			t.Fatal("focus not returned to input")
		}
		c.clearSelection()
		if changes != 1 {
			t.Fatal("empty clear emitted")
		}
	}
}

func TestComboboxClearAsyncAndReentrant(t *testing.T) {
	var token uint64
	c := Combobox("Choice").Multiple().Clearable(true).OnSearch(func(_ string, t uint64) { token = t })
	c.SetValue("a")
	c.open = true
	c.searchChanged()
	old := token
	c.text = "draft"
	c.OnValuesChange(func([]string) { c.SetValue("replacement") })
	c.clearSelection()
	if c.Value() != "replacement" || c.SetResults(old, "late") {
		t.Fatal("clear overwrote callback or accepted late result")
	}
}

func TestComboboxClearDisabledAndHidden(t *testing.T) {
	disabled := false
	c := Combobox("Choice", "a").Clearable(true)
	c.SetValue("a")
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(c.Render(cx)) })
	name := locale.Current().Name(locale.Current().Clear, "Choice")
	r := bounds(h, name)
	disabled = true
	h.Frame()
	h.Click(center(r))
	if c.Value() != "a" {
		t.Fatal("ancestor disabled cleared")
	}
	disabled = false
	c.SetDisabled(true)
	h.Frame()
	h.Click(center(r))
	if c.Value() != "a" {
		t.Fatal("disabled cleared")
	}
	c.SetDisabled(false)
	c.Clearable(false)
	h.Frame()
	if shown(h, name) {
		t.Fatal("clear button still shown")
	}
}

func TestComboboxClearFromKeyboard(t *testing.T) {
	c := Combobox("Choice", "a").Clearable(true)
	c.SetValue("a")
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() != "" || c.open {
		t.Fatal("keyboard clear failed", c.Value(), c.open)
	}
}
