package kit

import (
	"slices"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TreeNode is one node of a Tree. IDs must be unique within the tree.
type TreeNode struct {
	ID, Label string
	Children  []*TreeNode
	Disabled  bool
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
	multi, reorderable bool
	selection          base.Selection[string]
	typeahead          base.Typeahead
	dragID             string
	dragY              float32
	onSelection        func([]string)
	onReorder          func(string, string, int)
	roots              []*TreeNode
	nodes              map[string]*TreeNode
	reveal             bool
	open               map[string]bool
	selected           string
	disabled, plain    bool
	rows               []treeRow
	list               *VirtualListView
	onChange, onActive func(id string)
}

func Tree(roots ...*TreeNode) *TreeView {
	v := &TreeView{open: map[string]bool{}}
	v.list = VirtualList(0, 28, v.row).ItemKey(func(i int) string { return v.rows[i].node.ID })
	v.SetRoots(roots...)
	return v
}

func (v *TreeView) Height(dp float32) *TreeView { v.list.Height(dp); return v }
func (v *TreeView) Fill() *TreeView             { v.list.Fill(); return v }

// Plain drops the frame and background, like ListView.Plain.
func (v *TreeView) Plain() *TreeView                        { v.plain = true; return v }
func (v *TreeView) OnChange(fn func(id string)) *TreeView   { v.onChange = fn; return v }
func (v *TreeView) OnActivate(fn func(id string)) *TreeView { v.onActive = fn; return v }
func (v *TreeView) SetDisabled(on bool)                     { v.disabled = on }
func (v *TreeView) Expanded(id string) bool                 { return v.open[id] }
func (v *TreeView) SetExpanded(id string, open bool) {
	if _, ok := v.nodes[id]; ok {
		v.open[id] = open
	}
}

// SetRoots deep-copies the nodes and preserves selection/expansion by ID.
// Nil nodes are skipped. Empty or duplicate IDs (including cycles) panic before
// modifying the tree. Removed selections and expansion entries are discarded.
func (v *TreeView) SetRoots(roots ...*TreeNode) {
	nodes := make(map[string]*TreeNode)
	var clone func([]*TreeNode) []*TreeNode
	clone = func(source []*TreeNode) []*TreeNode {
		var out []*TreeNode
		for _, node := range source {
			if node == nil {
				continue
			}
			if node.ID == "" {
				panic("kit.Tree: empty node ID")
			}
			if _, ok := nodes[node.ID]; ok {
				panic("kit.Tree: duplicate node ID " + node.ID)
			}
			copy := *node
			nodes[node.ID] = &copy
			copy.Children = clone(node.Children)
			out = append(out, &copy)
		}
		return out
	}
	owned := clone(roots)
	v.roots, v.nodes = owned, nodes
	for id := range v.open {
		if _, ok := nodes[id]; !ok {
			delete(v.open, id)
		}
	}
	selection := v.selection
	selection.Keep(func(id string) bool { return nodes[id] != nil })
	v.SetValue(v.selected)
	v.selection = selection
}

// Value is the selected node's ID, or "".
func (v *TreeView) Value() string { return v.selected }

// SetValue selects a node without calling OnChange and expands its ancestors
// so it shows.
func (v *TreeView) SetValue(id string) {
	if _, ok := v.nodes[id]; !ok {
		id = ""
	}
	v.selected, v.reveal = id, id != ""
	if id != "" {
		v.selection.Set(id)
	} else {
		v.selection.Set()
	}
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
	if v.selected != "" && v.nodes[v.selected] != nil && !v.nodes[v.selected].Disabled && v.onActive != nil {
		v.onActive(v.selected)
	}
}

func (v *TreeView) key(cx *el.Context, e el.KeyEvent) bool {
	i := v.index(v.selected)
	if v.multi && key.Name(e.Name) == "A" && e.Modifiers.Contain(key.ModShortcut) {
		if e.State == el.KeyPress {
			before := v.SelectedIDs()
			var all []string
			for _, row := range v.rows {
				if !row.node.Disabled {
					all = append(all, row.node.ID)
				}
			}
			v.selection.Set(all...)
			if !slices.Equal(before, v.SelectedIDs()) && v.onSelection != nil {
				v.onSelection(v.SelectedIDs())
			}
		}
		return true
	}
	switch key.Name(e.Name) {
	case key.NameReturn:
		if e.State == el.KeyPress {
			v.activate()
		}
		return true
	case key.NameRightArrow, key.NameLeftArrow:
		if e.State != el.KeyPress || i < 0 || v.rows[i].node.Disabled {
			return true
		}
		r := v.rows[i]
		has := len(r.node.Children) > 0
		switch {
		case key.Name(e.Name) == key.NameRightArrow && has && !v.open[r.node.ID]:
			v.open[r.node.ID] = true
		case key.Name(e.Name) == key.NameRightArrow && has:
			for next := i + 1; next < len(v.rows) && v.rows[next].depth > r.depth; next++ {
				if !v.rows[next].node.Disabled {
					v.selectNode(cx, next, e.Modifiers)
					break
				}
			}
		case key.Name(e.Name) == key.NameLeftArrow && has && v.open[r.node.ID]:
			v.open[r.node.ID] = false
		case key.Name(e.Name) == key.NameLeftArrow:
			for parent := r.parent; parent >= 0; parent = v.rows[parent].parent {
				if !v.rows[parent].node.Disabled {
					v.selectNode(cx, parent, e.Modifiers)
					break
				}
			}
		}
		return true
	}
	nav := base.List{Count: len(v.rows), Disabled: func(i int) bool { return v.rows[i].node.Disabled }}
	if s, ok := base.Text(e.Name, e.Modifiers&typeaheadBlockers != 0); ok {
		if e.State == el.KeyPress {
			if j, found := v.typeahead.Find(time.Now(), s, i, nav, func(j int) string { return v.rows[j].node.Label }); found {
				v.selectNode(cx, j, 0)
			}
		}
		return true
	}
	j, ok := nav.Key(e.Name, i)
	if ok && e.State == el.KeyPress && len(v.rows) > 0 {
		v.selectNode(cx, j, e.Modifiers)
	}
	return ok
}

func (v *TreeView) row(cx *el.Context, i int) el.Element {
	r := v.rows[i]
	n, on := r.node, v.selectedNode(r.node.ID)
	state := ""
	if len(n.Children) > 0 {
		state = "collapsed"
		if v.open[n.ID] {
			state = "expanded"
		}
	}
	row := el.Div().Role("treeitem").Name(n.Label).Value(state).Selected(on).Disabled(n.Disabled).
		Row().Items(el.Center).Gap(theme.SpaceXs).Pl(float32(8 + 16*r.depth)).Pr(theme.SpaceMd).Rounded(theme.RadiusSm).Mx(4)
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
		if !v.disabled && !n.Disabled {
			twisty.CursorPointer().Focusable(false).OnClick(func() { v.open[n.ID] = !v.open[n.ID] })
		}
	}
	if !v.disabled && !n.Disabled {
		row.CursorPointer().
			OnClick(func() { v.selectNode(cx, i, cx.ClickModifiers()); cx.Focus(autoID("tree", v)) }).
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
	if v.reorderable && !v.disabled && !n.Disabled {
		row.OnDrag(func(e el.DragEvent) { v.dragNode(i, e) })
	}
	return row.Child(twisty, el.Text(n.Label).MaxLines(1).Grow())
}

func (v *TreeView) Render(cx *el.Context) el.Element {
	v.flatten()
	if v.reveal {
		v.list.ScrollTo(cx, v.index(v.selected))
		v.reveal = false
	}
	return listFrame(el.Div().ID(autoID("tree", v)).Role("tree").Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }), v.plain).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool { return v.key(cx, e) }).
		Child(v.list.Render(cx))
}
