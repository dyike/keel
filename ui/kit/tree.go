package kit

import (
	"slices"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// TreeNode is one node of a Tree. IDs must be unique within the tree.
type TreeNode struct {
	ID, Label string
	Children  []*TreeNode
	Disabled  bool
	// Lazy marks a branch whose children have not yet been loaded.
	Lazy bool
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
	renderItem         func(TreeItemContext) el.View
	indent             float32
	onExpand           func(string, bool)
	onLoad             func(string, uint64)
	loads              map[string]treeLoad
	loadToken          uint64
	selection          base.Selection[string]
	typeahead          base.Typeahead
	drag               *treeDrag
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
	v := &TreeView{open: map[string]bool{}, loads: map[string]treeLoad{}, indent: 16}
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
func (v *TreeView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		clear(v.loads)
	}
}
func (v *TreeView) Expanded(id string) bool          { return v.open[id] }
func (v *TreeView) SetExpanded(id string, open bool) { v.expand(id, open, false) }

// SetRoots deep-copies the nodes and preserves selection/expansion by ID.
// Nil nodes are skipped. Empty or duplicate IDs (including cycles) panic before
// modifying the tree. Removed selections and expansion entries are discarded.
func (v *TreeView) SetRoots(roots ...*TreeNode) {
	owned, nodes, err := cloneTree(roots)
	if err != nil {
		panic(err.Error())
	}
	clear(v.loads)
	v.install(owned, nodes)
	v.expandAncestors(v.selected)
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
	v.expandAncestors(id)
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
		has := len(r.node.Children) > 0 || r.node.Lazy
		switch {
		case key.Name(e.Name) == key.NameRightArrow && has && !v.open[r.node.ID]:
			v.expand(r.node.ID, true, true)
		case key.Name(e.Name) == key.NameRightArrow && has:
			for next := i + 1; next < len(v.rows) && v.rows[next].depth > r.depth; next++ {
				if !v.rows[next].node.Disabled {
					v.selectNode(cx, next, e.Modifiers)
					break
				}
			}
		case key.Name(e.Name) == key.NameLeftArrow && has && v.open[r.node.ID]:
			v.expand(r.node.ID, false, true)
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
	if len(n.Children) > 0 || n.Lazy {
		state = "collapsed"
		if v.open[n.ID] {
			state = "expanded"
		}
	}
	row := el.Div().Role("treeitem").Name(n.Label).Value(state).Selected(on).Disabled(n.Disabled).
		H(el.Dp(v.list.rowH)).Row().Items(el.Center).Gap(theme.SpaceXs).Pl(8 + v.indent*float32(r.depth)).Pr(theme.SpaceMd).Rounded(theme.RadiusSm).Mx(4)
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
			twisty.Name(locale.Current().MoreOptions + " " + n.Label).CursorPointer().Focusable(false).OnClick(func() { v.expand(n.ID, !v.open[n.ID], true) })
		}
	}

	if !v.disabled && !n.Disabled {
		activate := el.Div().ID("activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).OnClick(func() { v.selectID(cx, n.ID) }).OnDoubleClick(func() {
			if current := v.nodes[n.ID]; current != nil && (len(current.Children) > 0 || current.Lazy) {
				v.expand(n.ID, !v.open[n.ID], true)
			}
			v.activate()
		})
		if v.reorderable {
			activate.OnDrag(func(e el.DragEvent) { v.dragNode(v.index(n.ID), e) })
		}
		row.Child(activate).CursorPointer()
		if !on {
			row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	var content el.View
	if v.renderItem != nil {
		enabled := func() bool {
			return !v.disabled && v.nodes[n.ID] != nil && !v.nodes[n.ID].Disabled && cx.Enabled(autoID("tree", v))
		}
		content = v.renderItem(TreeItemContext{ID: n.ID, Label: n.Label, Index: i, Depth: r.depth, Expanded: v.open[n.ID], Selected: on, Disabled: v.disabled || n.Disabled, HasChildren: len(n.Children) > 0 || n.Lazy, Loading: v.NodeLoading(n.ID), Error: v.NodeError(n.ID), Toggle: func() {
			if enabled() {
				v.expand(n.ID, !v.open[n.ID], true)
			}
		}, Retry: func() {
			if enabled() {
				v.ReloadNode(n.ID)
			}
		}})
	}
	row.Child(twisty)
	v.dropMarker(row, i)
	if content != nil {
		return row.Child(el.Div().ID("content").Grow().MinW(el.Dp(0)).Child(content.Render(cx)))
	}
	row.Child(el.Text(n.Label).MaxLines(1).Grow())
	if v.NodeLoading(n.ID) {
		row.Child(Spinner().Render(cx))
	} else if message := v.NodeError(n.ID); message != "" {
		row.Child(el.Text(message).TextColor(theme.Danger).MaxLines(1), Button(locale.Current().Retry, func() { v.ReloadNode(n.ID) }).Size(24).Render(cx))
	}
	return row
}

func (v *TreeView) Render(cx *el.Context) el.Element {
	v.flatten()
	v.dragFrame(cx)
	if v.reveal {
		v.list.ScrollTo(cx, v.index(v.selected))
		v.reveal = false
	}
	return listFrame(el.Div().ID(autoID("tree", v)).Role("tree").Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }), v.plain).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool { return v.key(cx, e) }).
		Child(v.list.Render(cx))
}
