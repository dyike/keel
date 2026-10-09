package kit

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// Typing letters jumps to the item whose label starts with them; the same
// letter again cycles, and disabled items are skipped.
func TestTypeaheadInListMenuSelectAndTree(t *testing.T) {
	l := List("Apple", "Banana", "Blueberry", "Cherry")
	l.SetItemDisabled(3, true)
	h := sized(300, l)
	click(t, h, "Apple")
	h.Key("B", 0)
	h.Key("L", 0)
	if l.Value() != 2 {
		t.Fatalf("list bl = %d, want Blueberry", l.Value())
	}
	h.Key("C", 0)
	if l.Value() != 2 {
		t.Fatalf("list c chose disabled Cherry: %d", l.Value())
	}

	var ran []string
	m := Menu().Item("Copy", "", func() { ran = append(ran, "copy") }).
		Item("Cut", "", func() { ran = append(ran, "cut") }).
		Item("Paste", "", func() { ran = append(ran, "paste") })
	m.Trigger(Button("More", m.Toggle))
	h = uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().P(20).Child(m.Render(cx)) })))
	click(t, h, "More")
	h.Frame()
	h.Key("P", 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if len(ran) != 1 || ran[0] != "paste" {
		t.Fatalf("menu p ran %v", ran)
	}
	click(t, h, "More")
	h.Frame()
	h.Key("C", 0) // from Copy, the next item starting with c
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran[len(ran)-1] != "cut" {
		t.Fatalf("menu c again ran %v", ran)
	}

	s := Select("Fruit", "Apple", "Banana", "Cherry")
	h = page(s)
	clickRole(t, h, "select", "Fruit")
	h.Frame()
	h.Key("C", 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s.Value() != "Cherry" {
		t.Fatalf("select c chose %q", s.Value())
	}

	tree := Tree(&TreeNode{ID: "src", Label: "src"}, &TreeNode{ID: "docs", Label: "docs"}, &TreeNode{ID: "test", Label: "test"})
	h = sized(300, tree)
	click(t, h, "src")
	h.Key("T", 0)
	if tree.Value() != "test" {
		t.Fatalf("tree t chose %q", tree.Value())
	}
}
