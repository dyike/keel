package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func tableRightClick(t *testing.T, h *uitest.Harness, name string) {
	t.Helper()
	b := bounds(h, name)
	if b.Empty() {
		t.Fatalf("missing %s", name)
	}
	p := f32.Pt(float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2))
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: p, Buttons: pointer.ButtonSecondary}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	h.Frame()
	h.Frame()
}

func TestTableContextMenusSourceCoordinatesAndKeyboard(t *testing.T) {
	var selected TableCell
	rowActions := 0
	table := Table(Col("a").Width(100), Col("b").Width(100)).CellSelect().Height(120)
	table.SetRows([][]string{{"first", "one"}, {"second", "two"}})
	table.CellMenu(func(r, c int) *MenuView {
		selected = TableCell{r, c}
		if c == 0 {
			return nil
		}
		return Menu().Item("cell action", "", func() {})
	}).RowMenu(func(r int) *MenuView { return Menu().Item("row action", "", func() { rowActions = r + 1 }) })
	table.SortBy(0, true)
	table.MoveColumn(1, 0)
	h := sized(320, table)
	tableRightClick(t, h, "cell 1,1")
	if selected != (TableCell{1, 1}) || !shown(h, "cell action") {
		t.Fatalf("context %v menu=%+v", selected, table.contextMenu)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if shown(h, "cell action") {
		t.Fatal("Escape did not dismiss")
	}
	tableRightClick(t, h, "cell 0,0")
	if !shown(h, "row action") {
		t.Fatal("row fallback")
	}
	click(t, h, "row action")
	h.Frame()
	if rowActions != 1 || shown(h, "row action") {
		t.Fatal("action/dismiss")
	}
	click(t, h, "cell 1,1")
	h.Key(key.NameF10, key.ModShift)
	h.Frame()
	if !shown(h, "cell action") {
		t.Fatal("keyboard context action")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	table.SetDisabled(true)
	h.Frame()
	tableRightClick(t, h, "cell 1,1")
	if shown(h, "cell action") {
		t.Fatal("disabled context action")
	}
}

func TestTableContextMenuDoesNotConsumePrimaryCellAction(t *testing.T) {
	clicks := 0
	table := Table(Col("a").Cell(func(cx *el.Context, row int) el.Element { return Button("primary", func() { clicks++ }).Render(cx) })).RowMenu(func(int) *MenuView { return Menu().Item("context", "", func() {}) })
	table.SetRows([][]string{{"row"}})
	h := sized(320, table)
	click(t, h, "primary")
	if clicks != 1 {
		t.Fatal("context handler consumed primary button")
	}
	tableRightClick(t, h, "primary")
	if clicks != 1 || !shown(h, "context") {
		t.Fatalf("secondary click failed over button menu=%+v clicks=%d", table.contextMenu, clicks)
	}
	h.Click(390, 290)
	h.Frame()
	if shown(h, "context") {
		t.Fatal("outside click did not dismiss")
	}
	click(t, h, "primary")
	if clicks != 2 {
		t.Fatal("pointer capture stranded after menu dismissal")
	}
}

func TestTableContextMenuAncestorDisableAndFocusReturn(t *testing.T) {
	disabled := false
	table := Table(Col("a").Width(100), Col("b").Width(100)).CellSelect().RowMenu(func(int) *MenuView { return Menu().Item("context", "", func() {}) })
	table.SetRows([][]string{{"a", "b"}})
	h := render(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(320)).Disabled(disabled).Child(table.Render(cx))
	})
	click(t, h, "cell 0,0")
	h.Key(key.NameF10, key.ModShift)
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	if table.activeCell != (TableCell{0, 1}) {
		t.Fatal("context menu lost keyboard focus")
	}
	tableRightClick(t, h, "cell 0,1")
	disabled = true
	h.Frame()
	h.Frame()
	if shown(h, "context") || table.contextMenu.Value() {
		t.Fatal("ancestor disable left menu open")
	}
	disabled = false
	h.Frame()
	h.Frame()
	if shown(h, "context") {
		t.Fatal("menu reopened after enabling")
	}
}
