package kit

import (
	"fmt"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestTableFilterKeepsSourceIndexesAndSelection(t *testing.T) {
	table := Table(Col("a"), Col("b")).MultiSelect()
	table.SetRows([][]string{{"c", "yes"}, {"b", "no"}, {"a", "yes"}})
	table.SetSelectedRows([]int{0, 1})
	table.SortBy(0, false)
	table.SetFilter(func(row []string) bool { match := row[1] == "yes"; row[0] = "mutated"; return match })
	if table.Len() != 3 || table.VisibleLen() != 2 || table.list.Count() != 2 || table.order[0] != 2 || table.Row(0)[0] != "c" {
		t.Fatal("filter corrupted source/order/count")
	}
	if table.SelectionText() != "c\tyes" {
		t.Fatalf("filtered copy %q", table.SelectionText())
	}
	table.SetFilter(nil)
	if len(table.SelectedRows()) != 2 {
		t.Fatal("hidden selection lost")
	}
	table.SetFilter(func([]string) bool { return false })
	h := sized(320, table)
	if !shown(h, "暂无数据") {
		t.Fatal("empty filtered table has no empty state")
	}
}

func TestTableLoadMoreRetryAndDisabled(t *testing.T) {
	disabled := false
	table := Table(Col("a").Width(600)).Height(100)
	rows := make([][]string, 100)
	for i := range rows {
		rows[i] = []string{fmt.Sprint(i)}
	}
	table.SetRows(rows)
	requests := 0
	table.OnLoadMore(func() { requests++ })
	table.SetHasMore(true)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element {
		cx = c
		return el.Div().W(el.Dp(300)).Disabled(disabled).Child(table.Render(c))
	})
	for range 4 {
		h.Frame()
	}
	if requests != 0 {
		t.Fatal("loaded before approaching bottom")
	}
	disabled = true
	h.Frame()
	table.SetValue(99)
	for range 4 {
		h.Frame()
	}
	if requests != 0 {
		t.Fatal("disabled table loaded")
	}
	disabled = false
	for range 4 {
		h.Frame()
	}
	if requests != 1 || !table.loading {
		t.Fatalf("near end request %d", requests)
	}
	for range 4 {
		h.Frame()
	}
	if requests != 1 {
		t.Fatal("duplicate inflight request")
	}
	table.SetLoadError("network failed")
	h.Frame()
	h.Frame()
	if !shown(h, "重试") {
		t.Fatal("missing retry")
	}
	b := bounds(h, "重试")
	if b.Min.X < 0 || b.Max.X > 300 {
		t.Fatalf("retry outside viewport %v", b)
	}
	cx.ScrollIntoViewX(autoID("table", table), 550, 600)
	h.Frame()
	h.Frame()
	if bounds(h, "重试") != b {
		t.Fatal("retry scrolled with content")
	}
	for range 4 {
		h.Frame()
	}
	if requests != 1 {
		t.Fatal("error automatically retried")
	}
	click(t, h, "重试")
	if requests != 2 || !table.loading || table.loadError != "" {
		t.Fatal("retry failed")
	}
	table.SetLoading(false)
	for range 4 {
		h.Frame()
	}
	if requests != 2 {
		t.Fatal("same page requested again")
	}
	table.SetHasMore(false)
	table.SetRows(rows)
	for range 4 {
		h.Frame()
	}
	if requests != 2 {
		t.Fatal("loaded past final page")
	}
}
