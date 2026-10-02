package kit

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestTreeOwnsNodesAndPreservesIdentity(t *testing.T) {
	leaf := &TreeNode{ID: "leaf", Label: "original"}
	root := &TreeNode{ID: "root", Label: "root", Children: []*TreeNode{leaf}}
	tr := Tree(root, nil).Height(80)
	calls := 0
	tr.OnChange(func(string) { calls++ })
	tr.SetValue("leaf")
	leaf.Label = "mutated"
	root.Children = nil
	h := page(tr)
	if !shown(h, "original") || tr.Value() != "leaf" {
		t.Fatal("source tree mutations escaped")
	}
	tr.SetRoots(&TreeNode{ID: "new", Label: "new", Children: []*TreeNode{{ID: "leaf", Label: "moved"}}})
	h.Frame()
	if tr.Value() != "leaf" || !tr.Expanded("new") || tr.Expanded("root") || !shown(h, "moved") || calls != 0 {
		t.Fatal("SetRoots lost identity or called OnChange")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("cycle accepted")
			}
		}()
		cycle := &TreeNode{ID: "cycle"}
		cycle.Children = []*TreeNode{cycle}
		tr.SetRoots(cycle)
	}()
	if tr.Value() != "leaf" || tr.nodes["leaf"].Label != "moved" {
		t.Fatal("invalid roots modified tree")
	}
	tr.SetRoots(&TreeNode{ID: "other", Label: "other"})
	if tr.Value() != "" || tr.Expanded("new") || calls != 0 {
		t.Fatal("removed identity retained")
	}
	tr.SetValue("missing")
	if tr.Value() != "" {
		t.Fatal("unknown node selected")
	}
}

func TestTreeProgrammaticSelectionRevealsDistantNode(t *testing.T) {
	roots := make([]*TreeNode, 1000)
	for i := range roots {
		roots[i] = &TreeNode{ID: fmt.Sprint(i), Label: fmt.Sprintf("node-%d", i)}
	}
	tr := Tree(roots...).Height(100)
	h := page(tr)
	tr.SetValue("999")
	h.Frame()
	h.Frame()
	if !shown(h, "node-999") {
		t.Fatal("programmatic selection did not reveal distant node")
	}
}

func TestVirtualListStableKeysKeepEditorState(t *testing.T) {
	keys := []string{"a", "b", "c"}
	v := VirtualList(3, 40, func(cx *el.Context, i int) el.Element { return el.Input().Name(keys[i]) }).Height(160).ItemKey(func(i int) string { return keys[i] })
	h := page(v)
	clickClass(t, h, "Editor", "b")
	h.Type("draft")
	keys = []string{"c", "a", "b"}
	h.Frame()
	if desc(h, "b") != "draft" || desc(h, "a") == "draft" {
		t.Fatalf("state moved with index: a=%q b=%q", desc(h, "a"), desc(h, "b"))
	}
}
