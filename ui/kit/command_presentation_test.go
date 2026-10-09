package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestCommandInlineWithoutSearch(t *testing.T) {
	ran, searches, outside := 0, 0, 0
	c := Command(CommandItem{Title: "Blocked", Disabled: true}, CommandItem{Title: "Run", Action: func() { ran++ }}).Searchable(false).Inline(true).OnSearch(func(string, uint64) { searches++ })
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(Button("Outside", func() { outside++ }).Render(ctx), c.Render(ctx))
	})
	click(t, h, "Outside")
	h.Key(key.NameSpace, 0)
	h.Frame()
	if outside != 2 || searches != 0 {
		t.Fatal("inline stole focus or queried", outside, searches)
	}
	c.Focus(cx)
	h.Frame()
	h.Frame()
	h.Type("ignored")
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran != 1 || !c.Value() || c.query != "" {
		t.Fatal("inline confirmation/search", ran, c.Value(), c.query)
	}
	c.Searchable(true)
	h.Frame()
	c.Focus(cx)
	h.Frame()
	h.Frame()
	if searches != 1 {
		t.Fatal("restore search", searches)
	}
	token := c.request
	c.Searchable(false)
	h.Frame()
	if c.SetResults(token, CommandItem{Title: "stale"}) || c.loading {
		t.Fatal("mode transition retained async state")
	}
	c.SetDisabled(true)
	h.Frame()
	c.SetValue(true)
	if c.Value() {
		t.Fatal("disabled reopened")
	}
}

func TestCommandCustomContentAndActions(t *testing.T) {
	ran, nested := 0, 0
	action := Button("Inspect", func() { nested++ })
	header := Button("Header action", func() { nested++ })
	footer := Button("Footer action", func() { nested++ })
	c := Command(CommandItem{Title: "Run", Action: func() { ran++ }}).Header(header).Footer(footer).RowHeight(60).
		RenderItem(func(it CommandItem, active bool) el.View {
			return el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Row().Child(el.Text("Custom row"), action.Render(cx)) })
		})
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	click(t, h, "Header action")
	click(t, h, "Footer action")
	click(t, h, "Inspect")
	if nested != 3 || ran != 0 || !c.Value() {
		t.Fatal("nested click ran command", nested, ran, c.Value())
	}
	click(t, h, "Custom row")
	h.Frame()
	if ran != 1 || c.Value() {
		t.Fatal("row background did not execute", ran, c.Value())
	}
	c.RenderItem(nil).SetValue(true)
	h.Frame()
	h.Frame()
	if !shown(h, "Run") {
		t.Fatal("default renderer not restored")
	}
}

func TestCommandSlotsAcrossResultStates(t *testing.T) {
	footerCalls := 0
	c := Command().Header(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Command header") })).
		Footer(Button("Persistent footer", func() { footerCalls++ })).
		Empty(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Custom empty") })).OnSearch(func(string, uint64) {})
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	click(t, h, "Persistent footer")
	if !shown(h, "Command header") {
		t.Fatal("loading header")
	}
	c.SetSearchError(c.request, "failed")
	h.Frame()
	click(t, h, "Persistent footer")
	c.SetResults(c.request)
	h.Frame()
	click(t, h, "Persistent footer")
	if !shown(h, "Custom empty") || footerCalls != 3 {
		t.Fatal("slots lost", footerCalls)
	}
}

func TestCommandCustomRowsStayVirtual(t *testing.T) {
	items := make([]CommandItem, 10000)
	for i := range items {
		items[i] = CommandItem{Title: fmt.Sprintf("item-%05d", i)}
	}
	built := 0
	c := Command(items...).Searchable(false).RowHeight(48).RenderItem(func(it CommandItem, _ bool) el.View {
		built++
		return el.ViewFunc(func(*el.Context) el.Element { return el.Text(it.Title) })
	})
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	if built > 100 {
		t.Fatal("constructed all rows", built)
	}
	c.active = 9990
	h.Key(key.NamePageDown, 0)
	h.Frame()
	h.Frame()
	if c.active != 9999 || !shown(h, "item-09999") {
		t.Fatal("custom row scroll", c.active)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() {
		t.Fatal("non-searchable confirmation failed")
	}
}
