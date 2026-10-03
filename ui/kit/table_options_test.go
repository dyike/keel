package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"testing"
)

func TestTableWholeColumnSelectionPersistsAcrossDataAndFilter(t *testing.T) {
	v := Table(Col("A").Width(90), Col("B").Width(90).Selectable(false), Col("C").Width(90)).ColumnSelect()
	v.SetRows([][]string{{"a", "b", "c"}, {"d", "e", "f"}})
	calls := 0
	v.OnColumnSelectionChange(func([]int) { calls++ })
	h := renderView(v, 400, 1)
	click(t, h, "A")
	h.Frame()
	if !slices.Equal(v.SelectedColumns(), []int{0}) || calls != 1 || v.SelectionText() != "a\nd" {
		t.Fatal("column click", v.SelectedColumns(), calls, v.SelectionText())
	}
	h.Key(key.NameRightArrow, key.ModShift)
	if !slices.Equal(v.SelectedColumns(), []int{0, 2}) {
		t.Fatal("range did not skip locked column", v.SelectedColumns())
	}
	v.SetRows([][]string{{"a", "b", "c"}, {"d", "e", "f"}, {"g", "h", "i"}})
	if len(v.cells) != 0 || v.SelectionText() != "a\tc\nd\tf\ng\ti" {
		t.Fatal("new rows not in column selection", v.SelectionText())
	}
	v.SetFilter(func(row []string) bool { return row[0] != "d" })
	v.SetColumnVisible(0, false)
	if v.SelectionText() != "c\ni" || !slices.Equal(v.SelectedColumns(), []int{0, 2}) {
		t.Fatal("hidden/filter copy", v.SelectionText())
	}
	v.SetColumnVisible(0, true)
	h.Frame()
	click(t, h, "B")
	h.Frame()
	if !slices.Equal(v.SelectedColumns(), []int{0, 2}) {
		t.Fatal("locked header selectable")
	}
	v.SetRows(nil)
	h.Frame()
	click(t, h, "C")
	h.Frame()
	if !slices.Equal(v.SelectedColumns(), []int{2}) {
		t.Fatal("empty table column selection")
	}
	v.CellSelect()
	v.SetRows([][]string{{"a", "b", "c"}})
	h.Frame()
	v.SetSelectedCells([]TableCell{{0, 1}, {0, 2}})
	if !slices.Equal(v.SelectedCells(), []TableCell{{0, 2}}) {
		t.Fatal("cell setter ignored selectable constraint")
	}
}

func TestTableLockedMoveAndDensity(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Table(Col("A").Width(100).Movable(false).Resizable(false), Col("B").Width(100), Col("C").Width(100)).Stripe(true).RowHeight(28)
		v.MoveColumn(1, 0)
		if !slices.Equal(v.columns, []int{0, 1, 2}) {
			t.Fatal("crossed locked column")
		}
		v.MoveColumn(1, 2)
		if !slices.Equal(v.columns, []int{0, 2, 1}) {
			t.Fatal("movable columns")
		}
		v.SetRows([][]string{{"one", "b", "c"}, {"two", "e", "f"}})
		h := renderView(v, 400, scale)
		a, b := bounds(h, "one | b | c"), bounds(h, "two | e | f")
		if b.Min.Y-a.Min.Y != 28*scale {
			t.Fatal("density geometry", a, b)
		}
		v.RowHeight(48)
		h.Frame()
		a, b = bounds(h, "one | b | c"), bounds(h, "two | e | f")
		if b.Min.Y-a.Min.Y != 48*scale {
			t.Fatal("changed density geometry", a, b)
		}
	}
}

func TestTableColumnSelectionKeyboardAndDisabled(t *testing.T) {
	v := Table(Col("A"), Col("B").Selectable(false), Col("C")).ColumnSelect()
	v.SetRows([][]string{{"a", "b", "c"}})
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	cx.Focus(autoID("table", v))
	h.Frame()
	h.Key("A", key.ModShortcut)
	if !slices.Equal(v.SelectedColumns(), []int{0, 2}) {
		t.Fatal("select all")
	}
	v.SetDisabled(true)
	h.Frame()
	h.Key(key.NameHome, 0)
	if !slices.Equal(v.SelectedColumns(), []int{0, 2}) {
		t.Fatal("disabled keyboard mutated selection")
	}
	v.MultiSelect()
	if len(v.SelectedColumns()) != 0 || v.columnMode {
		t.Fatal("mode switch")
	}
}
