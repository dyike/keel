package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TreeNode is one node of a Tree. IDs must be unique within the tree.
type TreeNode struct {
	ID, Label string
	Children  []*TreeNode
}

type treeRow struct {
	node   *TreeNode
	depth  int
	parent int // index of the parent row, -1 at the top
}

// TreeView shows nested nodes that expand and collapse. With focus: ↑ ↓
// move, → expands or enters the first child, ← collapses or goes to the
// parent, Home / End jump, Enter activates. Only visible rows are built.
type TreeView struct {
	roots              []*TreeNode
	open               map[string]bool
	selected           string
	disabled           bool
	rows               []treeRow
	list               *VirtualListView
	onChange, onActive func(id string)
}

func Tree(roots ...*TreeNode) *TreeView {
	v := &TreeView{roots: roots, open: map[string]bool{}}
	v.list = VirtualList(0, 28, v.row)
	return v
}

func (v *TreeView) Height(dp float32) *TreeView             { v.list.Height(dp); return v }
func (v *TreeView) Fill() *TreeView                         { v.list.Fill(); return v }
func (v *TreeView) OnChange(fn func(id string)) *TreeView   { v.onChange = fn; return v }
func (v *TreeView) OnActivate(fn func(id string)) *TreeView { v.onActive = fn; return v }
func (v *TreeView) SetRoots(roots ...*TreeNode)             { v.roots = roots }
func (v *TreeView) SetDisabled(on bool)                     { v.disabled = on }
func (v *TreeView) Expanded(id string) bool                 { return v.open[id] }
func (v *TreeView) SetExpanded(id string, open bool)        { v.open[id] = open }

// Value is the selected node's ID, or "".
func (v *TreeView) Value() string { return v.selected }

// SetValue selects a node without calling OnChange and expands its ancestors
// so it shows.
func (v *TreeView) SetValue(id string) {
	v.selected = id
	var walk func(nodes []*TreeNode) bool
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

func (v *TreeView) flatten() {
	v.rows = v.rows[:0]
	var add func(nodes []*TreeNode, depth, parent int)
	add = func(nodes []*TreeNode, depth, parent int) {
		for _, n := range nodes {
			v.rows = append(v.rows, treeRow{n, depth, parent})
			if v.open[n.ID] && len(n.Children) > 0 {
				add(n.Children, depth+1, len(v.rows)-1)
			}
		}
	}
	add(v.roots, 0, -1)
	v.list.SetCount(len(v.rows))
}

func (v *TreeView) index(id string) int {
	for i, r := range v.rows {
		if r.node.ID == id {
			return i
		}
	}
	return -1
}

func (v *TreeView) choose(cx *el.Context, i int) {
	if i < 0 || i >= len(v.rows) {
		return
	}
	v.list.ScrollTo(cx, i)
	id := v.rows[i].node.ID
	if id == v.selected {
		return
	}
	v.selected = id
	if v.onChange != nil {
		v.onChange(id)
	}
}

func (v *TreeView) activate() {
	if v.selected != "" && v.onActive != nil {
		v.onActive(v.selected)
	}
}

func (v *TreeView) key(cx *el.Context, e el.KeyEvent) bool {
	i := v.index(v.selected)
	switch key.Name(e.Name) {
	case key.NameReturn:
		if e.State == el.KeyPress {
			v.activate()
		}
		return true
	case key.NameRightArrow, key.NameLeftArrow:
		if e.State != el.KeyPress || i < 0 {
			return true
		}
		r := v.rows[i]
		has := len(r.node.Children) > 0
		switch {
		case key.Name(e.Name) == key.NameRightArrow && has && !v.open[r.node.ID]:
			v.open[r.node.ID] = true
		case key.Name(e.Name) == key.NameRightArrow && has:
			v.choose(cx, i+1)
		case key.Name(e.Name) == key.NameLeftArrow && has && v.open[r.node.ID]:
			v.open[r.node.ID] = false
		case key.Name(e.Name) == key.NameLeftArrow:
			v.choose(cx, r.parent)
		}
		return true
	}
	j, ok := listKeys(e.Name, i, len(v.rows), 10)
	if ok && e.State == el.KeyPress && len(v.rows) > 0 {
		v.choose(cx, j)
	}
	return ok
}

func (v *TreeView) row(cx *el.Context, i int) el.Element {
	r := v.rows[i]
	n, on := r.node, r.node.ID == v.selected
	state := ""
	if len(n.Children) > 0 {
		state = "collapsed"
		if v.open[n.ID] {
			state = "expanded"
		}
	}
	row := el.Div().Role("treeitem").Name(n.Label).Value(state).Selected(on).
		Row().Items(el.Center).Gap(4).Pl(float32(8 + 16*r.depth)).Pr(8).Rounded(4).Mx(4)
	if on {
		row.Bg(theme.Highlight).TextColor(theme.PrimaryText)
	}
	twisty := el.Div().Size(el.Dp(16)).NoShrink()
	if state != "" {
		icon := IconChevronRight
		if state == "expanded" {
			icon = IconChevronDown
		}
		twisty.Child(Icon(icon).Size(16).Color(theme.Muted).Render(cx))
		if !v.disabled {
			twisty.CursorPointer().Focusable(false).OnClick(func() { v.open[n.ID] = !v.open[n.ID] })
		}
	}
	if !v.disabled {
		row.CursorPointer().
			OnClick(func() { v.choose(cx, i); cx.Focus(autoID("tree", v)) }).
			OnDoubleClick(func() {
				if len(n.Children) > 0 {
					v.open[n.ID] = !v.open[n.ID]
				}
				v.activate()
			})
		if !on {
			row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	return row.Child(twisty, el.Text(n.Label).MaxLines(1).Grow())
}

func (v *TreeView) Render(cx *el.Context) el.Element {
	v.flatten()
	return el.Div().ID(autoID("tree", v)).Role("tree").Disabled(v.disabled).
		Rounded(6).Border(1, theme.Border).Bg(theme.Surface).Py(4).Items(el.Stretch).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool { return v.key(cx, e) }).
		Child(v.list.Render(cx))
}
