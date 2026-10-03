package kit

import "github.com/dyike/keel/ui/el"

// TreeItemContext describes one visible row. Actions are for UI event handlers,
// not render callbacks. Disabled describes this node/the tree; ancestor-disabled
// containers are enforced by the outer element and actions at event time.
type TreeItemContext struct {
	ID, Label                                          string
	Index, Depth                                       int
	Expanded, Selected, Disabled, HasChildren, Loading bool
	Error                                              string
	Toggle, Retry                                      func()
}

// RenderItem replaces the content after the expand control. Nil callback/content
// uses the default label. Indentation, selection and disabled semantics remain
// on the row; child buttons operate independently. Reuse stateful child Views.
func (v *TreeView) RenderItem(fn func(TreeItemContext) el.View) *TreeView {
	v.renderItem = fn
	return v
}

// RowHeight sets the virtual row height in dp (minimum 20). Zero restores 28.
// Use enough height for custom content; rows are uniformly sized.
func (v *TreeView) RowHeight(dp float32) *TreeView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		if dp == 0 {
			dp = 28
		}
		v.list.rowH = max(20, dp)
		v.reveal = v.selected != ""
	}
	return v
}

// Indent sets the additional dp indent per level. Zero removes nesting indent.
func (v *TreeView) Indent(dp float32) *TreeView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.indent = dp
	}
	return v
}

// ScrollTo reveals an existing node by ID, expanding its ancestors without
// changing selection or sending selection/expansion notifications.
func (v *TreeView) ScrollTo(cx *el.Context, id string) bool {
	if v.nodes[id] == nil {
		return false
	}
	v.expandAncestors(id)
	v.flatten()
	v.list.ScrollTo(cx, v.index(id))
	return true
}
func (v *TreeView) expandAncestors(id string) {
	var walk func([]*TreeNode) bool
	walk = func(nodes []*TreeNode) bool {
		for _, n := range nodes {
			if n.ID == id || walk(n.Children) {
				if n.ID != id {
					v.open[n.ID] = true
				}
				return true
			}
		}
		return false
	}
	walk(v.roots)
}
func (v *TreeView) selectID(cx *el.Context, id string) {
	v.flatten()
	v.selectNode(cx, v.index(id), cx.ClickModifiers())
	cx.Focus(autoID("tree", v))
}
