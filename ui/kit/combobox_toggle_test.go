package kit

import (
	"gioui.org/io/input"
	"gioui.org/io/key"
	"image"
	"reflect"
	"testing"
)

func TestComboboxMultipleCandidateToggle(t *testing.T) {
	c := Combobox("Tags", "a", "b").Multiple()
	var events [][]string
	var primary []string
	c.OnValuesChange(func(values []string) {
		events = append(events, values)
		if len(values) > 0 {
			values[0] = "external"
		}
	})
	c.OnChange(func(value string) { primary = append(primary, value) })
	h := page(c)
	clickClass(t, h, "Editor", "Tags")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"a"}) || !c.open {
		t.Fatal(c.Values(), c.open)
	}
	click(t, h, "b")
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"a", "b"}) {
		t.Fatal(c.Values())
	}
	// Candidate and tag both expose the value; target the option explicitly.
	var option image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Label == "a" {
			option = n.Desc.Bounds
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	if option.Empty() {
		t.Fatal("missing option")
	}
	h.Click(center(option))
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"b"}) || !c.open {
		t.Fatal("remove non-primary", c.Values(), c.open)
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if len(c.Values()) != 0 || !c.open || len(events) != 4 || !reflect.DeepEqual(primary, []string{"a", "b", ""}) {
		t.Fatal(c.Values(), len(events), primary)
	}
}

func TestComboboxMultipleToggleRestrictionsAndSearch(t *testing.T) {
	var token uint64
	c := Combobox("Tags").Multiple().OnSearch(func(_ string, t uint64) { token = t })
	c.SetValues([]string{"a"})
	c.open = true
	c.searchChanged()
	old := token
	c.SetResults(token, "a", "b")
	c.DisableOption("a", true)
	c.choose("a")
	if len(c.Values()) != 1 || token != old {
		t.Fatal("disabled candidate toggled")
	}
	c.DisableOption("a", false)
	c.choose("a")
	if len(c.Values()) != 0 || !c.open || token == old || c.text != "" {
		t.Fatal("toggle did not refresh empty query")
	}
	if c.SetResults(old, "stale") {
		t.Fatal("old results accepted")
	}
	c.SetResults(token, "a", "b")
	c.choose("b")
	if !reflect.DeepEqual(c.Values(), []string{"b"}) {
		t.Fatal(c.Values())
	}
}
