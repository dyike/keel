package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/internal/uitest"
	"reflect"
	"testing"
)

func tableClickModifiers(t *testing.T, h *uitest.Harness, name string, mods key.Modifiers) {
	t.Helper()
	b := bounds(h, name)
	if b.Empty() {
		t.Fatalf("missing %s", name)
	}
	p := f32.Pt(float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2))
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: p, Buttons: pointer.ButtonPrimary, Modifiers: mods}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p, Modifiers: mods})
	h.Frame()
}
func TestTableMultipleRowSelectionAndCopy(t *testing.T) {
	changes := 0
	table := Table(Col("name").Width(100), Col("value").Width(100)).MultiSelect().Height(180).OnSelectionChange(func(rows []int) {
		changes++
		if len(rows) > 0 {
			rows[0] = 999
		}
	})
	table.SetRows([][]string{{"a", "1"}, {"b", "2"}, {"c", "3"}, {"d", "4"}})
	h := sized(320, table)
	click(t, h, "a | 1")
	tableClickModifiers(t, h, "c | 3", key.ModShortcut)
	if !reflect.DeepEqual(table.SelectedRows(), []int{0, 2}) {
		t.Fatalf("add %v", table.SelectedRows())
	}
	tableClickModifiers(t, h, "a | 1", key.ModShortcut)
	if !reflect.DeepEqual(table.SelectedRows(), []int{2}) {
		t.Fatalf("toggle %v", table.SelectedRows())
	}
	click(t, h, "b | 2")
	h.Key(key.NameDownArrow, key.ModShift)
	if !reflect.DeepEqual(table.SelectedRows(), []int{1, 2}) {
		t.Fatalf("range %v", table.SelectedRows())
	}
	h.Key(key.NameUpArrow, key.ModShift)
	if !reflect.DeepEqual(table.SelectedRows(), []int{1}) {
		t.Fatalf("shrink %v", table.SelectedRows())
	}
	table.SortBy(0, true)
	h.Frame()
	tableClickModifiers(t, h, "d | 4", key.ModShift)
	if !reflect.DeepEqual(table.SelectedRows(), []int{3, 2, 1}) {
		t.Fatalf("sorted range %v", table.SelectedRows())
	}
	table.MoveColumn(1, 0)
	h.Frame()
	h.Key("C", key.ModShortcut)
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != "4\td\n3\tc\n2\tb" {
		t.Fatalf("clipboard %q %v", data, ok)
	}
	h.Key("A", key.ModShortcut)
	if len(table.SelectedRows()) != 4 {
		t.Fatal("select all")
	}
	before := changes
	table.SetSelectedRows([]int{-1, 1, 1, 500, 2})
	if changes != before || !reflect.DeepEqual(table.SelectedRows(), []int{2, 1}) {
		t.Fatal("program selection")
	}
	table.SetRows([][]string{{"a", "1"}})
	if len(table.SelectedRows()) != 0 {
		t.Fatal("removed selection survived")
	}
}
func TestTableSelectionModesAndTSVEscaping(t *testing.T) {
	table := Table(Col("a"), Col("b"))
	table.SetRows([][]string{{"a\tb", "x\ny"}, {"z", "v"}})
	table.SetValue(0)
	table.MultiSelect()
	if !reflect.DeepEqual(table.SelectedRows(), []int{0}) {
		t.Fatal("enabling multi lost selection")
	}
	if got := table.SelectionText(); got != "\"a\tb\"\t\"x\ny\"" {
		t.Fatalf("TSV %q", got)
	}
	table.SetColumnVisible(1, false)
	if got := table.SelectionText(); got != "\"a\tb\"" {
		t.Fatalf("hidden copy %q", got)
	}
	table.SetValue(1)
	if !reflect.DeepEqual(table.SelectedRows(), []int{1}) {
		t.Fatal("SetValue not singleton")
	}
	table.SetDisabled(true)
	h := sized(320, table)
	click(t, h, "a\tb | x\ny")
	h.Key("A", key.ModShortcut)
	if !reflect.DeepEqual(table.SelectedRows(), []int{1}) {
		t.Fatal("disabled selection changed")
	}
}
