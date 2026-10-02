package kit

import (
	"encoding/json"
	"github.com/dyike/keel/ui/el"
	"math"
	"reflect"
	"testing"
)

func TestTableColumnLayoutRoundTrip(t *testing.T) {
	table := Table(Col("a").Width(100), Col("b").Flex(2), Col("c").Width(90)).FrozenColumns(1, 1)
	table.SetRows([][]string{{"b", "2", "y"}, {"a", "1", "x"}})
	calls := 0
	table.OnChange(func(int) { calls++ })
	table.SetValue(1)
	table.SortBy(0, false)
	table.MoveColumn(2, 0)
	table.SetColumnVisible(1, false)
	table.SetColumnWidth(2, 150)
	state := table.LayoutState()
	if state.Columns[0].Column != 2 || !state.Columns[2].Hidden {
		t.Fatalf("layout %+v", state)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	restored := Table(Col("a"), Col("b"), Col("c"))
	var decoded TableLayout
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if err = restored.SetLayoutState(decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, restored.LayoutState()) {
		t.Fatal("roundtrip changed layout")
	}
	decoded.Columns[0].Width = 900
	if restored.LayoutState().Columns[0].Width != 150 {
		t.Fatal("restored layout aliases caller data")
	}
	state.Columns[0].Width = 800
	if table.LayoutState().Columns[0].Width != 150 {
		t.Fatal("snapshot aliases table")
	}
	if table.Value() != 1 || table.order[0] != 1 || calls != 0 || table.Row(0)[0] != "b" {
		t.Fatal("column management changed row state")
	}
	before := table.LayoutState()
	for _, mutate := range []func(*TableLayout){
		func(s *TableLayout) { s.Columns = s.Columns[:1] },
		func(s *TableLayout) { s.Columns[1].Column = s.Columns[0].Column },
		func(s *TableLayout) { s.Columns[1].Width = float32(math.NaN()) },
		func(s *TableLayout) { s.Columns[1].Width = 1 },
		func(s *TableLayout) { s.Columns[1].Flex = 0 },
		func(s *TableLayout) { s.FrozenRight = 100 },
	} {
		bad := table.LayoutState()
		mutate(&bad)
		if table.SetLayoutState(bad) == nil {
			t.Fatal("invalid layout accepted")
		}
		if !reflect.DeepEqual(before, table.LayoutState()) {
			t.Fatal("invalid layout partially applied")
		}
	}
}

func TestTableMovingColumnsPreservesEditorAndFrozenEdges(t *testing.T) {
	table := Table(Col("a").Width(100), Col("b").Width(100).Cell(func(cx *el.Context, row int) el.Element { return el.Input().Name("edit") }), Col("c").Width(100)).Height(80)
	table.SetRows([][]string{{"a", "b", "c"}})
	h := sized(360, table)
	clickClass(t, h, "Editor", "edit")
	h.Type("draft")
	table.MoveColumn(1, 0)
	h.Frame()
	if desc(h, "edit") != "draft" {
		t.Fatal("column move lost editor state")
	}
	if bounds(h, "b").Min.X >= bounds(h, "a").Min.X {
		t.Fatal("header order unchanged")
	}
	table.SetColumnVisible(0, false)
	h.Frame()
	if shown(h, "a") {
		t.Fatal("hidden header still rendered")
	}
	if desc(h, "edit") != "draft" {
		t.Fatal("hiding sibling reset editor")
	}
	table.FrozenColumns(1, 1)
	h.Frame()
	if bounds(h, "c").Max.X != 359 {
		t.Fatalf("right edge %v", bounds(h, "c"))
	}
	table.SetColumnVisible(0, true)
	h.Frame()
	if !shown(h, "a") {
		t.Fatal("show column failed")
	}
	for c := range 3 {
		table.SetColumnVisible(c, false)
	}
	h.Frame()
	if shown(h, "a") || shown(h, "b") || shown(h, "c") {
		t.Fatal("all-hidden state")
	}
}
