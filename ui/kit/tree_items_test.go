package kit

import (
	"reflect"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
)

func TestTreeMultipleDisabledAndDynamicRoots(t *testing.T) {
	tree := Tree(&TreeNode{ID: "p", Label: "P", Children: []*TreeNode{{ID: "a", Label: "A"}, {ID: "b", Label: "B", Disabled: true}, {ID: "c", Label: "C"}}}, &TreeNode{ID: "z", Label: "Z"}).MultiSelect()
	tree.SetExpanded("p", true)
	calls := 0
	tree.OnSelectionChange(func(ids []string) {
		calls++
		if len(ids) > 0 {
			ids[0] = "external"
		}
	})
	h := sized(300, tree)
	click(t, h, "A")
	h.Key(key.NameDownArrow, key.ModShift)
	if !reflect.DeepEqual(tree.SelectedIDs(), []string{"a", "c"}) {
		t.Fatalf("range %v", tree.SelectedIDs())
	}
	click(t, h, "B")
	if tree.Value() != "c" {
		t.Fatal("disabled node selected")
	}
	tableClickModifiers(t, h, "Z", key.ModShortcut)
	if !reflect.DeepEqual(tree.SelectedIDs(), []string{"a", "c", "z"}) {
		t.Fatal("additive selection")
	}
	before := calls
	tree.SetRoots(&TreeNode{ID: "z", Label: "Z"}, &TreeNode{ID: "p", Label: "P", Children: []*TreeNode{{ID: "c", Label: "C"}}})
	if tree.Value() != "z" || !reflect.DeepEqual(tree.SelectedIDs(), []string{"z", "c"}) || calls != before {
		t.Fatal("dynamic roots lost valid selection")
	}
	tree.SetExpanded("p", false)
	if !reflect.DeepEqual(tree.SelectedIDs(), []string{"z", "c"}) {
		t.Fatal("collapsed selection lost")
	}
	tree.SetSelectedIDs([]string{"c"})
	h.Frame()
	if !tree.Expanded("p") || tree.Value() != "c" || calls != before {
		t.Fatal("program selection reveal/callback")
	}
}

func TestTreeMoveSubtreesIsAtomicAndOwned(t *testing.T) {
	tree := Tree(&TreeNode{ID: "a", Label: "A", Children: []*TreeNode{{ID: "child", Label: "Child"}}}, &TreeNode{ID: "b", Label: "B"}).MultiSelect()
	tree.SetSelectedIDs([]string{"a", "child"})
	before := tree.Roots()
	for _, move := range []struct {
		id, parent string
		index      int
	}{{"a", "child", 0}, {"missing", "", 0}, {"a", "missing", 0}, {"a", "", 9}} {
		if tree.MoveNode(move.id, move.parent, move.index) == nil {
			t.Fatal("invalid move accepted")
		}
		if !reflect.DeepEqual(before, tree.Roots()) {
			t.Fatal("invalid move partly applied")
		}
	}
	if err := tree.MoveNode("a", "b", 0); err != nil {
		t.Fatal(err)
	}
	roots := tree.Roots()
	if len(roots) != 1 || roots[0].ID != "b" || roots[0].Children[0].ID != "a" || !tree.Expanded("b") || !reflect.DeepEqual(tree.SelectedIDs(), []string{"a", "child"}) {
		t.Fatal("subtree or selection lost")
	}
	roots[0].Children[0].Label = "external"
	if tree.nodes["a"].Label != "A" {
		t.Fatal("Roots exposes owned state")
	}
}

func TestTreeDragReorderAndCancel(t *testing.T) {
	moved := ""
	tree := Tree(&TreeNode{ID: "a", Label: "A"}, &TreeNode{ID: "b", Label: "B"}, &TreeNode{ID: "c", Label: "C"}).Reorderable(func(id, parent string, index int) { moved = id })
	tree.SetValue("a")
	h := sized(300, tree)
	b := bounds(h, "A")
	x, y := float32(b.Min.X+50), float32(b.Min.Y+b.Dy()/2)
	h.Drag(x, y, x, y+56)
	if moved != "a" || tree.Roots()[2].ID != "a" || tree.Value() != "a" {
		t.Fatalf("drag %q %+v", moved, tree.Roots())
	}
	moved = ""
	tree.dragNode(0, el.DragEvent{Kind: el.DragStart, Y: 10})
	tree.dragNode(0, el.DragEvent{Kind: el.DragEnd, Y: 80, Canceled: true})
	if moved != "" {
		t.Fatal("cancel moved node")
	}
}

func TestListAndTreePageKeysStopAtEnabledBoundary(t *testing.T) {
	nav := base.List{Count: 5, Disabled: func(i int) bool { return i == 0 || i == 4 }}
	if i, _ := nav.Key(string(key.NamePageDown), 1); i != 3 {
		t.Fatalf("PageDown %d", i)
	}
	if i, _ := nav.Key(string(key.NamePageUp), 3); i != 1 {
		t.Fatalf("PageUp %d", i)
	}
}
