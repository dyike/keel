package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"strings"
	"testing"
)

func TestCommandSeparatorsFilteringNavigation(t *testing.T) {
	c := Command(CommandItem{Separator: true}, CommandItem{Title: "Alpha", Keywords: []string{"both"}},
		CommandItem{Separator: true}, CommandItem{Separator: true}, CommandItem{Title: "Disabled", Disabled: true},
		CommandItem{Title: "Beta", Group: "Group", Keywords: []string{"both"}}, CommandItem{Separator: true}).Inline(true)
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element { cx = ctx; return c.Render(ctx) })
	settle(h)
	countSeparators := func() int {
		n := 0
		for _, r := range c.rows {
			if r.item.Separator {
				n++
			}
		}
		return n
	}
	if countSeparators() != 1 || c.MatchedCount() != 3 {
		t.Fatal("divider normalization", c.rows)
	}
	c.Focus(cx)
	h.Frame()
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if c.SelectedIndex() != 5 {
		t.Fatal("did not skip divider/header/disabled", c.SelectedIndex())
	}
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if c.SelectedIndex() != 1 {
		t.Fatal("did not wrap")
	}
	c.SetQuery("both")
	settle(h)
	if c.MatchedCount() != 2 || countSeparators() != 1 {
		t.Fatal("filtered dividers")
	}
	c.SetQuery("Beta")
	settle(h)
	if countSeparators() != 0 || c.SelectedIndex() != 5 || c.MatchedCount() != 1 {
		t.Fatal("leading divider retained", c.rows)
	}
	c.SetQuery("nothing")
	settle(h)
	if len(c.rows) != 0 || c.SelectedIndex() != -1 {
		t.Fatal("empty filter headings/dividers", c.rows)
	}
}

func TestCommandVariableRowsMeasureRevealAndResize(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			items := make([]CommandItem, 10000)
			for i := range items {
				items[i] = CommandItem{Title: fmt.Sprintf("row-%05d", i)}
			}
			built := 0
			c := Command(items...).Inline(true).Searchable(false).RowHeight(20).AutoRowHeight(true).MaxHeight(200).
				RenderItem(func(item CommandItem, _ bool) el.View {
					built++
					var i int
					fmt.Sscanf(item.Title, "row-%d", &i)
					return el.ViewFunc(func(*el.Context) el.Element {
						return el.Div().H(el.Dp(float32(20 + i%3*30))).Child(el.Text(item.Title))
					})
				})
			var cx *el.Context
			h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return c.Render(ctx) }), 320, scale)
			settle(h)
			if c.variable.measured["item:0"] != 24 || c.variable.measured["item:2"] != 84 {
				t.Fatal("intrinsic measurements", c.variable.measured)
			}
			if built > 500 {
				t.Fatal("eager row construction", built)
			}
			c.Focus(cx)
			h.Frame()
			h.Frame()
			h.Key(key.NameEnd, 0)
			settle(h)
			if c.SelectedIndex() != 9999 || !shown(h, "row-09999") || c.variable.reveal != "" {
				t.Fatal("end reveal", c.SelectedIndex(), c.variable.reveal)
			}
			if built > 1000 {
				t.Fatal("nonvirtual reveal", built)
			}
			c.AutoRowHeight(false)
			settle(h)
			if !shown(h, "row-09999") {
				t.Fatal("switch to uniform lost active")
			}
			c.AutoRowHeight(true)
			settle(h)
			c.SetItems(CommandItem{Title: "row-00000"}, CommandItem{Title: "row-00001"})
			settle(h)
			if c.variable.Count() != 2 || !shown(h, "row-00000") {
				t.Fatal("model shrink retained offset")
			}
		})
	}
}

func TestCommandVariableRowsWidthInvalidation(t *testing.T) {
	width := float32(300)
	c := Command(CommandItem{Title: "Long"}).Inline(true).AutoRowHeight(true).PanelStyle(func(d *el.DivEl) { d.W(el.Dp(width)) }).RenderItem(func(CommandItem, bool) el.View {
		return el.ViewFunc(func(*el.Context) el.Element { return el.Text(strings.Repeat("wrapping text ", 12)) })
	})
	h := page(c)
	settle(h)
	wide := c.variable.measured["item:0"]
	width = 130
	settle(h)
	narrow := c.variable.measured["item:0"]
	if narrow <= wide {
		t.Fatal("width did not remeasure", wide, narrow)
	}
	c.SetLoading(true)
	h.Frame()
	if !c.IsLoading() {
		t.Fatal("loading getter")
	}
	c.SetLoading(false)
	settle(h)
	if c.SelectedIndex() != 0 {
		t.Fatal("loading recovery")
	}
}

func TestCommandHoverSelectionDoesNotRun(t *testing.T) {
	ran := 0
	var selections []int
	c := Command(CommandItem{Title: "First"}, CommandItem{Title: "Second", Action: func() { ran++ }}, CommandItem{Title: "Disabled", Disabled: true}).Inline(true).Searchable(false).OnSelect(func(i int) { selections = append(selections, i) })
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element { cx = ctx; return c.Render(ctx) })
	settle(h)
	b := bounds(h, "Second")
	h.Move(float32(b.Min.X+5), float32(b.Min.Y+5))
	settle(h)
	if c.SelectedIndex() != 1 || ran != 0 || !reflect.DeepEqual(selections, []int{0, 1}) {
		t.Fatal("hover events", selections, ran)
	}
	c.Focus(cx)
	h.Frame()
	h.Frame()
	h.Key(key.NameUpArrow, 0)
	settle(h)
	if c.SelectedIndex() != 0 {
		t.Fatal("stationary pointer stole keyboard highlight")
	}
	b = bounds(h, "Disabled")
	h.Move(float32(b.Min.X+5), float32(b.Min.Y+5))
	settle(h)
	if c.SelectedIndex() != 0 {
		t.Fatal("disabled hover selected")
	}
}

func TestCommandBindingHintsAndExecution(t *testing.T) {
	const name = "test.command.binding"
	defer core.Bind(name)
	core.Bind(name, "ctrl+j")
	ran := 0
	c := Command(CommandItem{Title: "Action", ActionName: name, Shortcut: "ctrl+z", Icon: IconSettings, Checked: true, Action: func() { ran++ }}).Inline(true).Searchable(false)
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(Button("Outside", func() {}).Render(ctx), c.Render(ctx))
	})
	settle(h)
	if !shown(h, "ctrl+j") || shown(h, "ctrl+z") {
		t.Fatal("named hint precedence")
	}
	c.Focus(cx)
	h.Frame()
	h.Frame()
	h.Key("J", key.ModCtrl)
	h.Frame()
	if ran != 1 {
		t.Fatal("bound action not executed", ran)
	}
	core.Bind(name, "ctrl+k")
	settle(h)
	if shown(h, "ctrl+j") || !shown(h, "ctrl+k") {
		t.Fatal("hint did not rebind")
	}
	h.Key("J", key.ModCtrl)
	h.Frame()
	h.Key("K", key.ModCtrl)
	h.Frame()
	if ran != 2 {
		t.Fatal("handler did not rebind", ran)
	}
	click(t, h, "Outside")
	h.Frame()
	h.Key("K", key.ModCtrl)
	h.Frame()
	if ran != 2 {
		t.Fatal("unfocused command ran", ran)
	}
	core.Bind(name)
	settle(h)
	if shown(h, "ctrl+k") {
		t.Fatal("unbound hint retained")
	}
}
