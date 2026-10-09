package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestComboboxFooterAcrossSearchStates(t *testing.T) {
	var token uint64
	actions, changes := 0, 0
	c := Combobox("Search").OnSearch(func(_ string, t uint64) { token = t }).OnChange(func(string) { changes++ })
	c.Footer(Button("Create", func() { actions++ }))
	h := page(c)
	clickClass(t, h, "Editor", "Search")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	click(t, h, "Create")
	if actions != 1 || changes != 0 || !c.open {
		t.Fatal("loading footer", actions, changes, c.open)
	}
	c.SetSearchError(token, "failed")
	h.Frame()
	click(t, h, "Create")
	if actions != 2 || !c.open {
		t.Fatal("error footer")
	}
	c.searchChanged()
	c.SetResults(token)
	h.Frame()
	click(t, h, "Create")
	if actions != 3 || !c.open {
		t.Fatal("empty footer")
	}
	c.searchChanged()
	c.SetResults(token, "one", "two")
	h.Frame()
	h.Key(key.NameReturn, 0)
	if actions != 4 || changes != 0 {
		t.Fatal("footer focus lost after results", actions, changes)
	}
	c.Footer(nil)
	h.Frame()
	if shown(h, "Create") {
		t.Fatal("footer not removed")
	}
}

func TestComboboxFooterCanSetSelection(t *testing.T) {
	c := Combobox("Search", "one", "two")
	calls := 0
	c.OnChange(func(string) { calls++ })
	c.Footer(Button("Create", func() { c.SetValue("created") }))
	h := page(c)
	clickClass(t, h, "Editor", "Search")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	click(t, h, "Create")
	h.Frame()
	if c.Value() != "created" || c.open || calls != 0 {
		t.Fatal(c.Value(), c.open, calls)
	}
}

func TestComboboxFooterFitsAndDisabledAnchor(t *testing.T) {
	for _, scale := range []int{1, 2} {
		options := make([]string, 100)
		for i := range options {
			options[i] = fmt.Sprint(i)
		}
		calls := 0
		c := Combobox("Search", options...).Footer(Button("Create", func() { calls++ }))
		disabled := false
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(c.Render(cx)) }), 300, scale)
		clickClass(t, h, "Editor", "Search")
		h.Key(key.NameDownArrow, 0)
		h.Frame()
		h.Frame()
		footer := bounds(h, "Create")
		if footer.Empty() || footer.Min.X < 0 || footer.Max.X > 300*scale {
			t.Fatal("footer outside width", footer)
		}
		disabled = true
		h.Frame()
		h.Frame()
		h.Click(center(footer))
		if calls != 0 || c.open {
			t.Fatal("disabled anchor retained footer interaction")
		}
	}
}
