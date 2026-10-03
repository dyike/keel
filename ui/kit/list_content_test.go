package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"testing"
)

func TestListGroupsSearchIndexesAndIndependentActions(t *testing.T) {
	l := List().Searchable(true).MultiSelect().Height(240)
	keywords := []string{"needle"}
	l.SetEntries(ListItem{ID: "a", Label: "Alpha", Group: "one", Keywords: keywords}, ListItem{ID: "b", Label: "Beta", Group: "one"}, ListItem{ID: "c", Label: "Gamma", Group: "two", Keywords: []string{"needle"}})
	keywords[0] = "mutated"
	actions, activations := 0, 0
	l.OnActivate(func(int) { activations++ }).RenderItem(func(cx *el.Context, item ListItemContext) el.Element {
		return el.Div().Row().Grow().Items(el.Center).Child(el.Text(item.Item.Label).Grow(), Button("action "+item.Item.ID, func() { actions++ }).Size(24).Render(cx))
	}).RenderGroupFooter(func(_ *el.Context, group string) el.Element { return el.Text("end " + group) })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return l.Render(c) })
	click(t, h, "action a")
	h.Frame()
	if actions != 1 || l.Value() != -1 {
		t.Fatal("child action selected row", actions, l.Value())
	}
	l.SetQuery("needle")
	h.Frame()
	if !shown(h, "Alpha") || shown(h, "Beta") || !shown(h, "end two") {
		t.Fatal("search/group footer")
	}
	cx.Focus(autoID("list", l))
	h.Frame()
	h.Key(key.NameEnd, 0)
	if l.Value() != 2 {
		t.Fatal("keyboard source index", l.Value())
	}
	h.Key(key.NameHome, key.ModShift)
	if !slices.Equal(l.SelectedValues(), []int{0, 2}) {
		t.Fatal("filtered range", l.SelectedValues())
	}
	l.SetQuery("absent")
	h.Frame()
	h.Key(key.NameReturn, 0)
	if activations != 0 || shown(h, "one") {
		t.Fatal("hidden selection activated/header remained")
	}
}

func TestListLoadMoreRetryAndDisabled(t *testing.T) {
	l := List("one").Height(100)
	requests := 0
	l.OnLoadMore(func() { requests++ })
	l.SetHasMore(true)
	h := renderView(l, 300, 1)
	for range 8 {
		h.Frame()
	}
	if requests != 1 || !l.loading {
		t.Fatal("initial request", requests)
	}
	l.SetLoadError("failed")
	for range 4 {
		h.Frame()
	}
	if requests != 1 {
		t.Fatal("auto retried error")
	}
	click(t, h, "重试")
	h.Frame()
	if requests != 2 {
		t.Fatal("retry", requests)
	}
	l.SetDisabled(true)
	l.SetEntries(ListItem{ID: "one", Label: "one"})
	l.SetLoading(false)
	for range 4 {
		h.Frame()
	}
	if requests != 2 {
		t.Fatal("disabled requested")
	}
	l.SetDisabled(false)
	l.SetHasMore(false)
	h.Frame()
	if requests != 2 {
		t.Fatal("loaded final page requested")
	}
}
