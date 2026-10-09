package kit

import (
	"fmt"
	"slices"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
)

func (v *TreeView) SetNodeDisabled(id string, on bool) {
	if n := v.nodes[id]; n != nil {
		n.Disabled = on
		if on {
			delete(v.loads, id)
		}
	}
}
func (v *TreeView) MultiSelect() *TreeView {
	v.multi = true
	v.SetSelectedIDs([]string{v.selected})
	return v
}
func (v *TreeView) OnSelectionChange(fn func([]string)) *TreeView { v.onSelection = fn; return v }

// SelectedIDs includes collapsed selections in tree order, as a fresh snapshot.
func (v *TreeView) SelectedIDs() []string {
	if !v.multi {
		if v.selected != "" {
			return []string{v.selected}
		}
		return nil
	}
	var ids []string
	var walk func([]*TreeNode)
	walk = func(nodes []*TreeNode) {
		for _, n := range nodes {
			if v.selection.Has(n.ID) {
				ids = append(ids, n.ID)
			}
			walk(n.Children)
		}
	}
	walk(v.roots)
	return ids
}
func (v *TreeView) SetSelectedIDs(ids []string) {
	v.SetValue("")
	var keep []string
	for _, id := range ids {
		if v.nodes[id] != nil {
			keep = append(keep, id)
			v.SetValue(id)
			if !v.multi {
				break
			}
		}
	}
	v.selection.Set(keep...)
}
func (v *TreeView) selectedNode(id string) bool {
	if v.multi {
		return v.selection.Has(id)
	}
	return id == v.selected
}
func (v *TreeView) selectNode(cx *el.Context, i int, mods key.Modifiers) {
	if i < 0 || i >= len(v.rows) || v.rows[i].node.Disabled {
		return
	}
	before := v.SelectedIDs()
	if v.multi {
		order := make([]string, len(v.rows))
		for p, r := range v.rows {
			order[p] = r.node.ID
		}
		v.selection.Click(order, i, mods.Contain(key.ModShift), mods.Contain(key.ModShortcut), func(p int) bool { return v.rows[p].node.Disabled })
	}
	v.choose(cx, i)
	if !slices.Equal(before, v.SelectedIDs()) && v.onSelection != nil {
		v.onSelection(v.SelectedIDs())
	}
}

// Roots returns an owned snapshot for persistence after a move.
func (v *TreeView) Roots() []*TreeNode {
	var clone func([]*TreeNode) []*TreeNode
	clone = func(nodes []*TreeNode) []*TreeNode {
		out := make([]*TreeNode, len(nodes))
		for i, n := range nodes {
			copy := *n
			copy.Children = clone(n.Children)
			out[i] = &copy
		}
		return out
	}
	return clone(v.roots)
}
func (v *TreeView) siblings(id string) (*[]*TreeNode, int, string) {
	var find func(*[]*TreeNode, string) (*[]*TreeNode, int, string)
	find = func(nodes *[]*TreeNode, parent string) (*[]*TreeNode, int, string) {
		for i, n := range *nodes {
			if n.ID == id {
				return nodes, i, parent
			}
			if list, pos, p := find(&n.Children, n.ID); list != nil {
				return list, pos, p
			}
		}
		return nil, -1, ""
	}
	return find(&v.roots, "")
}

// MoveNode moves a subtree into parent ("" means roots). index is its final
// sibling position; len(target) means append. Invalid moves leave the tree intact.
// Programmatic moves do not invoke callbacks.
func (v *TreeView) MoveNode(id, parent string, index int) error {
	node := v.nodes[id]
	if node == nil {
		return fmt.Errorf("kit.Tree: unknown node %q", id)
	}
	target := &v.roots
	if parent != "" {
		p := v.nodes[parent]
		if p == nil {
			return fmt.Errorf("kit.Tree: unknown parent %q", parent)
		}
		target = &p.Children
	}
	var contains func(*TreeNode) bool
	contains = func(n *TreeNode) bool {
		if n.ID == parent {
			return true
		}
		for _, c := range n.Children {
			if contains(c) {
				return true
			}
		}
		return false
	}
	if contains(node) {
		return fmt.Errorf("kit.Tree: cannot move a node into its descendant")
	}
	if index < 0 || index > len(*target) {
		return fmt.Errorf("kit.Tree: invalid insertion index")
	}
	source, from, _ := v.siblings(id)
	*source = slices.Delete(*source, from, from+1)
	index = min(index, len(*target))
	*target = slices.Insert(*target, index, node)
	if parent != "" {
		v.open[parent] = true
	}
	selection := v.selection
	v.SetValue(v.selected)
	v.selection = selection
	v.flatten()
	return nil
}
func (v *TreeView) Reorderable(fn func(id, parent string, index int)) *TreeView {
	v.reorderable = true
	v.onReorder = fn
	return v
}
