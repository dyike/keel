package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestComboboxCustomRowsAndFallback(t *testing.T) {
	c := Combobox("Choice").RowHeight(60)
	c.SetItems(ComboboxItem{Value: "a", Label: "Alpha"}, ComboboxItem{Value: "b", Label: "Beta", Disabled: true})
	var seen ComboboxItem
	c.RenderItem(func(item ComboboxItem, selected bool) el.View {
		if item.Value == "b" {
			seen = item
			return nil
		}
		return el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(el.Text("Rich Alpha"), el.Text("Description")) })
	})
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if !shown(h, "Description") || !shown(h, "Beta") || !seen.Disabled {
		t.Fatal("custom row/fallback/metadata")
	}
	if bounds(h, "Beta").Dy() != 58 {
		t.Fatal("row height", bounds(h, "Beta"))
	}
	click(t, h, "Rich Alpha")
	if c.Value() != "a" {
		t.Fatal("custom content not selectable")
	}
	c.RenderItem(nil).RowHeight(0)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if shown(h, "Description") || !shown(h, "Alpha") {
		t.Fatal("default label not restored")
	}
	c.RowHeight(float32(math.NaN())).RowHeight(-1)
	if c.rowHeight != 0 {
		t.Fatal("invalid height accepted")
	}
}

func TestComboboxRendererVirtualAndActionIsolation(t *testing.T) {
	options := make([]string, 10000)
	for i := range options {
		options[i] = fmt.Sprintf("Item %05d", i)
	}
	rendered, actions := 0, 0
	buttons := map[string]*ButtonView{}
	c := Combobox("Choice", options...).RowHeight(42)
	c.RenderItem(func(item ComboboxItem, _ bool) el.View {
		rendered++
		if buttons[item.Value] == nil {
			buttons[item.Value] = Button("Action "+item.Value, func() { actions++ })
		}
		return buttons[item.Value]
	})
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if rendered > 100 {
		t.Fatal("renderer eagerly built entire dataset", rendered)
	}
	click(t, h, "Action Item 00000")
	if actions != 1 || c.Value() != "" || !c.open {
		t.Fatal("child action selected parent", actions, c.Value(), c.open)
	}
	c.SetOptions("Item 00001", "Item 00000")
	h.Frame()
	h.Key(key.NameReturn, 0)
	if actions != 2 || c.Value() != "" {
		t.Fatal("reorder lost child focus", actions, c.Value())
	}
}
