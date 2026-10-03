package kit

import (
	"fmt"
	"gioui.org/io/key"
	"reflect"
	"testing"
)

func TestComboboxGroupsFilterAndKeyboard(t *testing.T) {
	c := Combobox("Food")
	c.SetGroups(ComboboxGroup{ID: "fruit", Label: "Fruit", Items: []ComboboxItem{{Value: "apple", Label: "Apple"}, {Value: "pear", Label: "Pear", Disabled: true}}}, ComboboxGroup{ID: "veg", Label: "Vegetables", Items: []ComboboxItem{{Value: "carrot", Label: "Carrot"}}}, ComboboxGroup{ID: "empty", Label: "Empty"})
	h := page(c)
	clickClass(t, h, "Editor", "Food")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if !shown(h, "Fruit") || !shown(h, "Vegetables") || shown(h, "Empty") {
		t.Fatal("headers")
	}
	click(t, h, "Fruit")
	if c.Value() != "" || !c.open {
		t.Fatal("header selected or dismissed")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if c.Value() != "carrot" {
		t.Fatal("navigation did not skip heading/disabled row", c.Value())
	}
	c.text = "app"
	c.open = true
	c.searchChanged()
	h.Frame()
	h.Frame()
	if !shown(h, "Fruit") || shown(h, "Vegetables") || !reflect.DeepEqual(c.matches(), []string{"apple"}) {
		t.Fatal("empty filtered group visible")
	}
	if c.displayIndex(0) != 1 {
		t.Fatal("header offset missing")
	}
	c.SetItems(ComboboxItem{Value: "plain"})
	h.Frame()
	if shown(h, "Fruit") || len(c.groupFor) != 0 {
		t.Fatal("flat mode retained groups")
	}
}

func TestComboboxGroupsCopyDedupAndAsync(t *testing.T) {
	items := []ComboboxItem{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}}
	groups := []ComboboxGroup{{ID: "one", Items: items}, {ID: "two", Items: []ComboboxItem{{Value: "a", Label: "duplicate"}, {Value: "c", Label: "C"}}}, {ID: "one", Items: []ComboboxItem{{Value: "d"}}}, {Items: []ComboboxItem{{Value: "invalid"}}}}
	var token uint64
	c := Combobox("Choice").OnSearch(func(_ string, t uint64) { token = t })
	c.open = true
	c.searchChanged()
	if !c.SetGroupResults(token, groups...) {
		t.Fatal("current groups rejected")
	}
	groups[0].ID = "external"
	items[0].Label = "external"
	if !reflect.DeepEqual(c.options, []string{"a", "b", "c"}) || c.optionLabel("a") != "A" || c.groupFor["a"] != "one" {
		t.Fatal("copy/dedup", c.options)
	}
	if len(c.displayRows) != 5 || c.displayIndex(2) != 4 {
		t.Fatal("visual row mapping", c.displayRows)
	}
	c.SetValue("a")
	if c.SetGroupResults(token, ComboboxGroup{ID: "late"}) || c.groupFor["a"] != "one" {
		t.Fatal("stale replaced groups")
	}
}

func TestComboboxGroupedVirtualReveal(t *testing.T) {
	groups := make([]ComboboxGroup, 40)
	for i := range groups {
		groups[i] = ComboboxGroup{ID: fmt.Sprint(i), Label: fmt.Sprintf("Group %d", i), Items: []ComboboxItem{{Value: fmt.Sprint(i), Label: fmt.Sprintf("Choice %d", i)}}}
	}
	c := Combobox("Grouped").Size(48)
	c.SetGroups(groups...)
	h := renderView(c, 300, 2)
	clickClass(t, h, "Editor", "Grouped")
	h.Key(key.NameDownArrow, 0)
	for i := 0; i < 25; i++ {
		h.Key(key.NameDownArrow, 0)
		h.Frame()
	}
	h.Frame()
	if c.active != 25 || !shown(h, "Choice 25") {
		t.Fatal("grouped virtual reveal", c.active)
	}
	c.Size(28).RowHeight(34)
	h.Frame()
	h.Frame()
	if !shown(h, "Choice 25") {
		t.Fatal("resize reveal lost group offsets")
	}
	h.Key(key.NameReturn, 0)
	if c.Value() != "25" {
		t.Fatal("wrong grouped candidate selected", c.Value())
	}
}
