package kit

import (
	"gioui.org/io/key"
	"reflect"
	"testing"
)

func TestComboboxItemsDisplayAndValues(t *testing.T) {
	c := Combobox("Country")
	items := []ComboboxItem{{Value: "cn", Label: "中国"}, {Value: "jp", Label: "日本", Disabled: true}, {Value: "us", Label: "美国"}, {Value: "cn", Label: "duplicate"}, {Label: "empty"}}
	c.SetItems(items...)
	items[0].Label = "external"
	if len(c.options) != 3 || c.optionLabel("cn") != "中国" {
		t.Fatal("ownership/dedup")
	}
	var values []string
	c.OnChange(func(v string) { values = append(values, v) })
	h := page(c)
	clickClass(t, h, "Editor", "Country")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	click(t, h, "中国")
	if c.Value() != "cn" || c.text != "中国" || !reflect.DeepEqual(values, []string{"cn"}) {
		t.Fatal(c.Value(), c.text, values)
	}
	c.SetItems(ComboboxItem{Value: "cn", Label: "China"}, ComboboxItem{Value: "us", Label: "United States"})
	if c.text != "China" || len(values) != 1 {
		t.Fatal("rename selection", c.text, values)
	}
	c.text = "United"
	if !reflect.DeepEqual(c.matches(), []string{"us"}) {
		t.Fatal("label search", c.matches())
	}
	c.text = "United States"
	c.settle()
	if c.Value() != "us" || values[len(values)-1] != "us" {
		t.Fatal("typed label submitted instead of value")
	}
	c.SetOptions("us", "cn")
	if c.text != "us" || c.optionLabel("cn") != "cn" {
		t.Fatal("string mode restore")
	}
}

func TestComboboxItemAsyncLabelsAndDisabled(t *testing.T) {
	var token uint64
	c := Combobox("Country").Multiple().OnSearch(func(_ string, t uint64) { token = t })
	c.open = true
	c.searchChanged()
	if !c.SetItemResults(token, ComboboxItem{Value: "cn", Label: "中国"}, ComboboxItem{Value: "jp", Label: "日本", Disabled: true}) {
		t.Fatal("results rejected")
	}
	c.choose("cn")
	old := token
	c.SetItemResults(token, ComboboxItem{Value: "jp", Label: "日本", Disabled: true})
	if c.active != -1 || c.optionLabel("cn") != "中国" {
		t.Fatal("selection label or disabled highlight")
	}
	c.choose("jp")
	if !reflect.DeepEqual(c.Values(), []string{"cn"}) {
		t.Fatal("disabled chosen")
	}
	c.DisableOption("jp", false)
	c.choose("jp")
	if len(c.Values()) != 1 {
		t.Fatal("override bypassed item disabled")
	}
	c.close()
	if c.SetItemResults(old, ComboboxItem{Value: "late"}) {
		t.Fatal("stale accepted")
	}
}

func TestComboboxItemLabelCollisionDoesNotCommitOnBlur(t *testing.T) {
	c := Combobox("Choice")
	c.SetItems(ComboboxItem{Value: "one", Label: "two"}, ComboboxItem{Value: "two", Label: "second"})
	c.SetValue("one")
	c.settle()
	if c.Value() != "one" {
		t.Fatal("unchanged display resolved to another value")
	}
}
