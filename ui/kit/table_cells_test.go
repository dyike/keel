package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

func TestTableCellRangesColumnsAndCopy(t *testing.T) {
	table := Table(Col("a").Width(80), Col("b").Width(80), Col("c").Width(80)).CellSelect().Height(160)
	table.SetRows([][]string{{"a1", "b1", "c1"}, {"a2", "b2", "c2"}, {"a3", "b3", "c3"}})
	changes := 0
	table.OnCellSelectionChange(func(c []TableCell) {
		changes++
		if len(c) > 0 {
			c[0] = TableCell{999, 999}
		}
	})
	h := sized(300, table)
	click(t, h, "cell 0,0")
	h.Key(key.NameRightArrow, key.ModShift)
	h.Key(key.NameDownArrow, key.ModShift)
	want := []TableCell{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	if !reflect.DeepEqual(table.SelectedCells(), want) {
		t.Fatalf("rectangle %v", table.SelectedCells())
	}
	h.Key("C", key.ModShortcut)
	if _, text, ok := h.Router.WriteClipboard(); !ok || string(text) != "a1\tb1\na2\tb2" {
		t.Fatalf("copy %q", text)
	}
	table.SortBy(0, true)
	table.MoveColumn(1, 0)
	h.Frame()
	if table.SelectionText() != "b2\ta2\nb1\ta1" {
		t.Fatalf("sorted copy %q", table.SelectionText())
	}
	click(t, h, "c")
	if table.SelectionText() != "c3\nc2\nc1" {
		t.Fatalf("column copy %q", table.SelectionText())
	}
	tableClickModifiers(t, h, "a", key.ModShift)
	if len(table.SelectedCells()) != 6 {
		t.Fatalf("column range %v", table.SelectedCells())
	}
	before := changes
	table.SetSelectedCells([]TableCell{{0, 0}, {2, 2}, {999, 0}})
	if changes != before || len(table.SelectedCells()) != 2 {
		t.Fatal("program selection")
	}
	if table.SelectionText() != "\tc3\na1\t" {
		t.Fatalf("sparse copy %q", table.SelectionText())
	}
	table.SetColumnVisible(2, false)
	h.Frame()
	if table.SelectionText() != "\na1" {
		t.Fatalf("hidden copy %q", table.SelectionText())
	}
	table.SetSelectedCells(nil)
	if table.Value() != -1 || len(table.SelectedCells()) != 0 {
		t.Fatal("clear selection")
	}
}

func TestTableCellSelectionKeyboardRevealAndEditorFocus(t *testing.T) {
	var cx *el.Context
	table := Table(Col("first").Width(80), Col("second").Width(180), Col("third").Width(180), Col("last").Width(80)).FrozenColumns(1, 1).CellSelect().Height(120)
	table.SetRows([][]string{{"a", "b", "c", "d"}})
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(320)).Child(table.Render(c)) })
	click(t, h, "cell 0,0")
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if x, _, _ := cx.ScrollStateX(autoID("table", table)); x <= 0 {
		t.Fatal("cell keyboard did not reveal horizontally")
	}
	if table.activeCell != (TableCell{0, 2}) {
		t.Fatalf("active cell %v", table.activeCell)
	}
	table.SetDisabled(true)
	h.Frame()
	h.Key(key.NameLeftArrow, 0)
	if table.activeCell != (TableCell{0, 2}) {
		t.Fatal("disabled cell navigation")
	}

	editor := Table(Col("editor").Cell(func(cx *el.Context, row int) el.Element { return el.Input().Name("editable") })).CellSelect()
	editor.SetRows([][]string{{""}})
	eh := sized(320, editor)
	clickClass(t, eh, "Editor", "editable")
	eh.Type("draft")
	if desc(eh, "editable") != "draft" {
		t.Fatal("cell click stole input focus")
	}
}

func TestTableSelectionVirtualKeyboardAndInitialReveal(t *testing.T) {
	for _, cells := range []bool{false, true} {
		table := Table(Col("first").Width(180), Col("last").Width(180)).Height(80)
		if cells {
			table.CellSelect()
		} else {
			table.MultiSelect()
		}
		rows := make([][]string, 1000)
		for i := range rows {
			rows[i] = []string{"first", "last"}
		}
		table.SetRows(rows)
		var cx *el.Context
		if cells {
			table.SetSelectedCells([]TableCell{{999, 1}})
		} else {
			table.SetValue(999)
		}
		h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(300)).Child(table.Render(c)) })
		h.Frame()
		h.Frame()
		if cells {
			if x, _, _ := cx.ScrollStateX(autoID("table", table)); x == 0 {
				t.Fatal("initial cell reveal lost")
			}
			click(t, h, "cell 999,1")
		} else {
			click(t, h, "first | last")
		}
		h.Key(key.NameHome, key.ModShortcut)
		h.Key(key.NameDownArrow, 0)
		h.Key(key.NameEnd, key.ModShortcut)
		h.Key(key.NameUpArrow, 0)
		if table.Value() != 998 {
			t.Fatalf("virtual focus lost (cells=%v): %d", cells, table.Value())
		}
	}
}
