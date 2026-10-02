package kit

import (
	"slices"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// DockAxis is the direction in which a split's two children are arranged.
type DockAxis uint8

const (
	DockAxisHorizontal DockAxis = iota
	DockAxisVertical
)

// DockPlacement selects the edge of the target group for a new split.
type DockPlacement uint8

const (
	DockPlacementLeft DockPlacement = iota
	DockPlacementRight
	DockPlacementTop
	DockPlacementBottom
)

// DockNode is either a tab group (Panels, Active), or a binary split
// (First, Second, Axis, Ratio). Ratio is the first child's share, in [0.05,0.95].
// Each panel occurs exactly once across the three region trees.
type DockNode struct {
	Panels        []string  `json:",omitempty"`
	Active        string    `json:",omitempty"`
	Axis          DockAxis  `json:",omitempty"`
	Ratio         float32   `json:",omitempty"`
	First, Second *DockNode `json:",omitempty"`
}

func cloneDockNode(n *DockNode) *DockNode {
	if n == nil {
		return nil
	}
	c := *n
	c.Panels = slices.Clone(n.Panels)
	c.First, c.Second = cloneDockNode(n.First), cloneDockNode(n.Second)
	return &c
}
func dockPanels(n *DockNode) []string {
	if n == nil {
		return nil
	}
	if n.First == nil {
		return slices.Clone(n.Panels)
	}
	return append(dockPanels(n.First), dockPanels(n.Second)...)
}
func firstDockGroup(n *DockNode) *DockNode {
	for n != nil && n.First != nil {
		n = n.First
	}
	return n
}
func findDockGroup(n *DockNode, id string) *DockNode {
	if n == nil {
		return nil
	}
	if slices.Contains(n.Panels, id) {
		return n
	}
	if found := findDockGroup(n.First, id); found != nil {
		return found
	}
	return findDockGroup(n.Second, id)
}
func removeDockPanel(n *DockNode, id string) *DockNode {
	if n == nil {
		return nil
	}
	if n.First == nil {
		n.Panels = slices.DeleteFunc(n.Panels, func(s string) bool { return s == id })
		if len(n.Panels) == 0 {
			return nil
		}
		return n
	}
	n.First, n.Second = removeDockPanel(n.First, id), removeDockPanel(n.Second, id)
	if n.First == nil {
		return n.Second
	}
	if n.Second == nil {
		return n.First
	}
	return n
}
func (v *DockView) tree(s DockSide) **DockNode {
	switch s {
	case DockRight:
		return &v.layout.RightTree
	case DockBottom:
		return &v.layout.BottomTree
	}
	return &v.layout.LeftTree
}
func (v *DockView) appendTree(s DockSide, id string) {
	root := v.tree(s)
	if *root == nil {
		*root = &DockNode{}
	}
	n := firstDockGroup(*root)
	n.Panels = append(n.Panels, id)
	if n.Active == "" {
		n.Active = id
	}
}
func (v *DockView) restoreNode(n *DockNode, seen map[*DockNode]bool, depth int) (*DockNode, bool) {
	if n == nil {
		return nil, true
	}
	if depth > 32 || seen[n] {
		return nil, false
	}
	seen[n] = true
	if n.First == nil && n.Second == nil {
		if n.Axis != 0 || n.Ratio != 0 {
			return nil, false
		}
		c := &DockNode{Active: n.Active}
		for _, id := range n.Panels {
			if _, ok := v.panels[id]; ok {
				c.Panels = append(c.Panels, id)
			}
		}
		if len(c.Panels) == 0 {
			return nil, true
		}
		return c, true
	}
	if n.First == nil || n.Second == nil || len(n.Panels) > 0 || n.Active != "" || n.Axis > DockAxisVertical || !finiteNumber(float64(n.Ratio)) || n.Ratio < .05 || n.Ratio > .95 {
		return nil, false
	}
	a, ok := v.restoreNode(n.First, seen, depth+1)
	if !ok {
		return nil, false
	}
	b, ok := v.restoreNode(n.Second, seen, depth+1)
	if !ok {
		return nil, false
	}
	if a == nil {
		return b, true
	}
	if b == nil {
		return a, true
	}
	return &DockNode{First: a, Second: b, Axis: n.Axis, Ratio: n.Ratio}, true
}
func (v *DockView) nodeShown(n *DockNode) []string {
	var ids []string
	if n != nil {
		for _, id := range n.Panels {
			if v.Visible(id) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
func (v *DockView) fixNode(n *DockNode) {
	if n == nil {
		return
	}
	if n.First != nil {
		v.fixNode(n.First)
		v.fixNode(n.Second)
		return
	}
	ids := v.nodeShown(n)
	if !slices.Contains(ids, n.Active) {
		n.Active = ""
		if len(ids) > 0 {
			n.Active = ids[0]
		}
	}
}

// Split moves id into its own group on the given edge of target's group.
// Other tabs remain together. Empty groups collapse. Programmatic changes
// do not call OnLayoutChange. Invalid arguments leave the layout unchanged.
func (v *DockView) Split(id, target string, placement DockPlacement) bool {
	if id == target || placement > DockPlacementBottom || v.where(id) < 0 || v.where(target) < 0 || dockGroupDepth(*v.tree(DockSide(v.where(target))), target, 0) >= 32 {
		return false
	}
	v.cancelResize()
	for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
		*v.tree(s) = removeDockPanel(*v.tree(s), id)
	}
	side := DockSide(v.where(target))
	n := findDockGroup(*v.tree(side), target)
	old := *n
	fresh := &DockNode{Panels: []string{id}, Active: id}
	*n = DockNode{First: &old, Second: fresh, Ratio: .5}
	if placement == DockPlacementTop || placement == DockPlacementBottom {
		n.Axis = DockAxisVertical
	}
	if placement == DockPlacementLeft || placement == DockPlacementTop {
		n.First, n.Second = n.Second, n.First
	}
	v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
	v.syncTrees()
	return true
}
func (v *DockView) syncTrees() {
	v.splitSizes = map[*DockNode]float32{}
	for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
		ids, _ := v.side(s)
		*ids = dockPanels(*v.tree(s))
		v.fixActive(s)
	}
	// Tree edits may remove groups; discard their menus as well.
	for n, m := range v.menus {
		present := false
		for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
			for _, id := range n.Panels {
				if findDockGroup(*v.tree(s), id) == n {
					present = true
				}
			}
		}
		if !present {
			m.SetValue(false)
			delete(v.menus, n)
		}
	}
}
func (v *DockView) region(cx *el.Context, s DockSide) el.Element {
	child := v.renderNode(cx, s, *v.tree(s))
	if child == nil {
		return nil
	}
	sz := v.fitted(s)
	box := el.Div().NoShrink().Items(el.Stretch).Child(child)
	if s == DockBottom {
		box.H(el.Dp(sz))
	} else {
		box.W(el.Dp(sz))
	}
	return box.Decorate(func(gtx core.C, draw func()) { v.painted[s] = sz; draw() })
}
func (v *DockView) renderNode(cx *el.Context, s DockSide, n *DockNode) el.Element {
	if n == nil {
		return nil
	}
	if n.First == nil {
		return v.group(cx, s, n)
	}
	a, b := v.renderNode(cx, s, n.First), v.renderNode(cx, s, n.Second)
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	// Zero basis lets flex divide the available space by the persisted ratio.
	left, right := el.Div().Flex(n.Ratio).Items(el.Stretch).Child(a), el.Div().Flex(1-n.Ratio).Items(el.Stretch).Child(b)
	row := el.Div().Grow().MinW(el.Dp(0)).MinH(el.Dp(0)).Items(el.Stretch)
	handle := el.Div().Role("separator").Name(locale.Current().Resize).Focusable(true).NoShrink().Bg(theme.Border).
		FocusStyle(func(st *el.Style) { st.Bg(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if e.Modifiers != 0 {
				return false
			}
			old := n.Ratio
			ratio := old
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameUpArrow:
				ratio -= .05
			case key.NameRightArrow, key.NameDownArrow:
				ratio += .05
			case key.NameHome:
				ratio = .05
			case key.NameEnd:
				ratio = .95
			default:
				return false
			}
			if e.State == el.KeyPress {
				n.Ratio = max(.05, min(.95, ratio))
				if n.Ratio != old {
					v.changed()
				}
			}
			return true
		})
	// Drag callbacks are rebuilt each frame, so state lives on the split itself
	// through a per-split interaction record (see splitHandle).
	v.splitHandle(n, handle)
	if n.Axis == DockAxisHorizontal {
		row.Row()
		left.W(el.Dp(0))
		right.W(el.Dp(0))
		handle.W(el.Dp(4)).Cursor(pointer.CursorColResize)
	} else {
		left.H(el.Dp(0))
		right.H(el.Dp(0))
		handle.H(el.Dp(4)).Cursor(pointer.CursorRowResize)
	}
	row.Child(left, handle, right)
	return row.Decorate(func(gtx core.C, draw func()) {
		w, h := cx.LayoutSize(row)
		axisSize := w - 4
		if n.Axis == DockAxisVertical {
			axisSize = h - 4
		}
		v.splitSizes[n] = axisSize
		draw()
	})
}

func (v *DockView) splitHandle(n *DockNode, h *el.DivEl) {
	h.OnDrag(func(e el.DragEvent) {
		pos := e.X
		if n.Axis == DockAxisVertical {
			pos = e.Y
		}
		switch e.Kind {
		case el.DragStart:
			v.splitResize = n
			v.splitStart = n.Ratio
			v.splitGrab = pos
		case el.DragMove:
			if v.splitResize == n && v.splitSizes[n] > 0 {
				n.Ratio = max(.05, min(.95, n.Ratio+(pos-v.splitGrab)/v.splitSizes[n]))
			}
		case el.DragEnd:
			if v.splitResize != n {
				return
			}
			v.splitResize = nil
			if e.Canceled {
				n.Ratio = v.splitStart
			} else if n.Ratio != v.splitStart {
				v.changed()
			}
		}
	})
}

func dockGroupDepth(n *DockNode, id string, depth int) int {
	if n == nil {
		return -1
	}
	if slices.Contains(n.Panels, id) {
		return depth
	}
	if d := dockGroupDepth(n.First, id, depth+1); d >= 0 {
		return d
	}
	return dockGroupDepth(n.Second, id, depth+1)
}
