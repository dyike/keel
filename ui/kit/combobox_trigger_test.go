package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestComboboxCustomTriggerSearchAndRestore(t *testing.T) {
	c := Combobox("Country").Placeholder("Pick")
	c.SetItems(ComboboxItem{Value: "cn", Label: "China"}, ComboboxItem{Value: "us", Label: "USA"})
	var snapshot ComboboxTriggerContext
	c.RenderTrigger(func(state ComboboxTriggerContext) el.View {
		snapshot = state
		return el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(12).Child(el.Text("Custom country")) })
	})
	h := page(c)
	click(t, h, "Custom country")
	h.Frame()
	h.Frame()
	search := locale.Current().Name(locale.Current().Search, "Country")
	if !c.open || !shown(h, search) || !snapshot.Open || snapshot.Size != 36 {
		t.Fatal("custom open/search", c.open, snapshot)
	}
	h.Type("Chi")
	h.Frame()
	if c.text != "Chi" || len(c.matches()) != 1 {
		t.Fatal("popup search focus/filter", c.text, c.matches())
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() != "cn" || c.open || len(snapshot.Selection) != 1 || snapshot.Selection[0].Label != "China" {
		t.Fatal(c.Value(), snapshot)
	}
	snapshot.Selection[0].Label = "external"
	if c.optionLabel("cn") != "China" {
		t.Fatal("selection snapshot mutated model")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	h.Frame()
	if !c.open {
		t.Fatal("custom focus not restored")
	}
	c.RenderTrigger(nil)
	h.Frame()
	clickClass(t, h, "Editor", "Country")
	if c.open || shown(h, "Custom country") || c.Value() != "cn" {
		t.Fatal("restore failed")
	}
}

func TestComboboxCustomTriggerActionsAndDisabled(t *testing.T) {
	c := Combobox("Choice", "a", "b").Searchable(false)
	c.SetValue("a")
	var state ComboboxTriggerContext
	open := Button("Open custom", func() { state.Toggle() })
	clear := Button("Clear custom", func() { state.Clear() })
	c.RenderTrigger(func(s ComboboxTriggerContext) el.View {
		state = s
		return el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Row().Child(open.Render(cx), clear.Render(cx)) })
	})
	disabled := false
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(c.Render(cx)) }), 400, 1)
	click(t, h, "Open custom")
	h.Frame()
	if !c.open {
		t.Fatal("nested toggle double activated")
	}
	click(t, h, "b")
	h.Frame()
	if c.Value() != "b" || c.open {
		t.Fatal("nonsearchable custom selection")
	}
	click(t, h, "Clear custom")
	h.Frame()
	if c.Value() != "" || c.open {
		t.Fatal("custom clear also toggled")
	}
	disabled = true
	h.Frame()
	click(t, h, "Open custom")
	if c.open {
		t.Fatal("ancestor disabled toggle")
	}
}
