package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

func TestListStableEntriesAndMultipleSelection(t *testing.T) {
	l := List().MultiSelect()
	l.SetEntries(ListItem{ID: "a", Label: "A"}, ListItem{ID: "b", Label: "B", Disabled: true}, ListItem{ID: "c", Label: "C"}, ListItem{ID: "d", Label: "D"})
	calls := 0
	l.OnSelectionChange(func(values []int) {
		calls++
		if len(values) > 0 {
			values[0] = 999
		}
	})
	h := sized(300, l)
	click(t, h, "A")
	h.Key(key.NameDownArrow, key.ModShift)
	if !reflect.DeepEqual(l.SelectedValues(), []int{0, 2}) {
		t.Fatalf("disabled range %v", l.SelectedValues())
	}
	click(t, h, "B")
	if l.Value() != 2 {
		t.Fatal("disabled item selected")
	}
	tableClickModifiers(t, h, "D", key.ModShortcut)
	if !reflect.DeepEqual(l.SelectedValues(), []int{0, 2, 3}) {
		t.Fatal("additive selection")
	}
	before := calls
	l.SetEntries(ListItem{ID: "d", Label: "D new"}, ListItem{ID: "a", Label: "A new"}, ListItem{ID: "c", Label: "C new"})
	if l.Value() != 0 || !reflect.DeepEqual(l.SelectedValues(), []int{0, 1, 2}) || calls != before {
		t.Fatalf("identity %d %v", l.Value(), l.SelectedValues())
	}
	entries := l.Entries()
	entries[0].Label = "external"
	if l.Items()[0] != "D new" {
		t.Fatal("entries alias")
	}
	l.SetItemDisabled(2, true)
	h.Frame()
	h.Key(key.NameEnd, 0)
	if l.Value() != 1 {
		t.Fatalf("End disabled skipping %d", l.Value())
	}
	h.Key("A", key.ModShortcut)
	if !reflect.DeepEqual(l.SelectedValues(), []int{0, 1}) {
		t.Fatal("select all includes disabled")
	}
}

func TestListDragReorderAndCancel(t *testing.T) {
	from, to := -1, -1
	l := List("A", "B", "C", "D").Reorderable(func(a, b int) { from, to = a, b })
	l.SetValue(1)
	h := sized(300, l)
	b := bounds(h, "B")
	x, y := float32(b.Min.X+20), float32(b.Min.Y+b.Dy()/2)
	h.Drag(x, y, x, y+64)
	if from != 1 || to != 3 || !reflect.DeepEqual(l.Items(), []string{"A", "C", "D", "B"}) || l.Value() != 3 {
		t.Fatalf("drag %d %d %v selected %d", from, to, l.Items(), l.Value())
	}
	from = -1
	l.dragItem(0, el.DragEvent{Kind: el.DragStart, Y: 10})
	l.dragItem(0, el.DragEvent{Kind: el.DragEnd, Y: 100, Canceled: true})
	if from != -1 || l.Items()[0] != "A" {
		t.Fatal("cancel committed reorder")
	}
	l.SetItemDisabled(0, true)
	h.Frame()
	b = bounds(h, "A")
	h.Drag(float32(b.Min.X+20), float32(b.Min.Y+10), float32(b.Min.X+20), float32(b.Min.Y+74))
	if from != -1 || l.Items()[0] != "A" {
		t.Fatal("disabled item moved")
	}
	l.Move(3, 0)
	if l.Value() != 0 || from != -1 {
		t.Fatal("program move did not preserve identity or fired callback")
	}
}

func TestListInvalidEntriesAreAtomic(t *testing.T) {
	l := List("A")
	l.SetValue(0)
	for _, entries := range [][]ListItem{{{ID: "", Label: "bad"}}, {{ID: "x"}, {ID: "x"}}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid IDs accepted")
				}
			}()
			l.SetEntries(entries...)
		}()
		if l.Items()[0] != "A" || l.Value() != 0 {
			t.Fatal("invalid entries partly applied")
		}
	}
}
