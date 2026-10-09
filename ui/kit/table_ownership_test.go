package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestTableOwnsRowsAndColumnConfiguration(t *testing.T) {
	spec := Col("number").Width(100)
	a, b := Table(spec), Table(spec)
	source := [][]string{{"2"}, {"1"}}
	a.SetRows(source)
	a.SortBy(0, false)
	source[0][0] = "0"
	source[1] = []string{"9"}
	if a.Row(0)[0] != "2" || a.Row(1)[0] != "1" || a.order[0] != 1 {
		t.Fatal("source mutations escaped into sorted data")
	}
	rows := a.Rows()
	rows[0][0] = "7"
	rows[1] = nil
	row := a.Row(1)
	row[0] = "8"
	if a.Row(0)[0] != "2" || a.Row(1)[0] != "1" {
		t.Fatal("returned rows expose internal data")
	}
	spec.Width(250)
	if a.cols[0].width != 100 || b.cols[0].width != 100 {
		t.Fatal("external column remains shared")
	}
	a.SetColumnWidth(0, 180)
	if b.cols[0].width != 100 || a.cols[0].width != 180 {
		t.Fatal("column width leaked across tables")
	}
	calls := 0
	a.OnChange(func(int) { calls++ })
	a.SetValue(1)
	a.SetRows([][]string{{"0"}})
	if a.Value() != -1 || calls != 0 || a.order[0] != 0 {
		t.Fatal("programmatic replacement selection/sort")
	}
	if a.Row(-1) != nil || a.Row(1) != nil {
		t.Fatal("invalid row must be nil")
	}
}

func TestTableSortPreservesCustomCellIdentity(t *testing.T) {
	var table *TableView
	table = Table(Col("name"), Col("editor").Cell(func(cx *el.Context, row int) el.Element { return el.Input().Name("edit-" + table.Row(row)[0]) })).Height(120)
	table.SetRows([][]string{{"b"}, {"a"}})
	h := sized(360, table)
	clickClass(t, h, "Editor", "edit-b")
	h.Type("draft")
	if desc(h, "edit-b") != "draft" {
		t.Fatal("row click stole focus from the cell editor")
	}
	table.SortBy(0, false)
	h.Frame()
	if desc(h, "edit-b") != "draft" || desc(h, "edit-a") == "draft" {
		t.Fatalf("sorted editor state: a=%q b=%q", desc(h, "edit-a"), desc(h, "edit-b"))
	}
}

func TestTableRowFocusNavigatesBeforeViewportScroll(t *testing.T) {
	table := Table(Col("name")).Height(120)
	table.SetRows([][]string{{"first"}, {"second"}, {"third"}, {"fourth"}, {"fifth"}})
	h := sized(360, table)
	click(t, h, "first")
	h.Key(key.NameDownArrow, 0)
	if table.Value() != 1 {
		t.Fatalf("Down after row click selected %d, want 1", table.Value())
	}
}
