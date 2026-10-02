package kit

import (
	"slices"

	"gioui.org/io/key"
	"gioui.org/op"

	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// DockSide is where a Dock region sits around the center.
type DockSide uint8

const (
	DockLeft DockSide = iota
	DockRight
	DockBottom
	// DockCenter holds documents in the middle, in tab groups that split
	// like the side regions. While it has any, they replace the center view.
	DockCenter
)

// dockSides are every region, sides first.
var dockSides = []DockSide{DockLeft, DockRight, DockBottom, DockCenter}

// DockPanel is a tool window that lives in a Dock region. IDs must be unique.
// The panel gets the region's full height; long content should scroll itself.
type DockPanel struct {
	ID, Title string
	View      el.View
}

// DockLayout is everything about a Dock's arrangement, for saving and
// restoring it (it encodes as JSON). Panel lists are in tab order.
type DockLayout struct {
	LeftTree, RightTree, BottomTree *DockNode `json:",omitempty"`
	CenterTree                      *DockNode `json:",omitempty"`
	Center                          []string  `json:",omitempty"`
	CenterActive                    string    `json:",omitempty"`
	// Detached panels are open in windows of their own (see OnDetach).
	Detached                        []string `json:",omitempty"`
	Version                         int      `json:"version,omitempty"`
	Left, Right, Bottom             []string `json:",omitempty"`
	LeftActive, RightActive         string   `json:",omitempty"`
	BottomActive                    string   `json:",omitempty"`
	LeftSize, RightSize, BottomSize float32
	Hidden                          []string `json:",omitempty"`
	// Zoomed is the panel maximized over the whole dock, or "".
	Zoomed string `json:",omitempty"`
}

// DockView arranges tool panels around a central view, like an IDE. Each
// region holds tabbed panels and resizes against the center; a panel's menu
// moves it to another region or closes it. Layout and SetLayout save and
// restore the arrangement.
type DockView struct {
	drag                  dockDrag
	groupRects            map[*DockNode]dockRect
	tabRects              map[string]dockRect
	tabOrigins            map[string][2]float32
	centerRect            dockRect
	focusTab              string
	splitSizes            map[*DockNode]float32
	center                el.View
	panels                map[string]DockPanel
	layout                DockLayout
	painted               [4]float32
	centerSize            float32  // unused: the center has no size of its own
	bounds                dockRect // the whole dock in window dp, for dragging a tab out
	documents             bool     // the app uses DockCenter, so tabs may be dropped there
	onDetach              func(p DockPanel, reattach func())
	viewport              [2]float32
	total                 [2]float32 // painted width and height of the whole dock, dp
	grab                  float32
	menus                 map[*DockNode]*MenuView
	onLayout              func(DockLayout)
	disabled              bool
	resizing              bool
	resizeSide            DockSide
	resizeStart           float32
	splitResize           *DockNode
	splitStart, splitGrab float32
}

func Dock(center el.View) *DockView {
	return &DockView{groupRects: map[*DockNode]dockRect{}, tabRects: map[string]dockRect{}, tabOrigins: map[string][2]float32{}, splitSizes: map[*DockNode]float32{}, center: center, panels: map[string]DockPanel{},
		layout: DockLayout{Version: 2, LeftSize: 240, RightSize: 260, BottomSize: 180},
		menus:  map[*DockNode]*MenuView{}}
}

// Panel adds a panel to a region; the first panel added to a region is active.
func (v *DockView) Panel(p DockPanel, side DockSide) *DockView {
	if p.ID == "" || side > DockCenter {
		return v
	}
	_, exists := v.panels[p.ID]
	v.panels[p.ID] = p
	if side == DockCenter {
		v.documents = true
	}
	if exists {
		return v
	}
	ids, active := v.side(side)
	*ids = append(*ids, p.ID)
	v.appendTree(side, p.ID)
	if *active == "" {
		*active = p.ID
	}
	return v
}

// OnLayoutChange runs after the user moves, closes, switches or resizes a
// panel, with the new layout to save.
func (v *DockView) OnLayoutChange(fn func(DockLayout)) *DockView { v.onLayout = fn; return v }

// Layout returns a copy of the arrangement.
func (v *DockView) Layout() DockLayout {
	l := v.layout
	l.LeftTree, l.RightTree, l.BottomTree = cloneDockNode(l.LeftTree), cloneDockNode(l.RightTree), cloneDockNode(l.BottomTree)
	l.CenterTree, l.Center, l.Detached = cloneDockNode(l.CenterTree), slices.Clone(l.Center), slices.Clone(l.Detached)
	l.Left, l.Right, l.Bottom, l.Hidden = slices.Clone(l.Left), slices.Clone(l.Right), slices.Clone(l.Bottom), slices.Clone(l.Hidden)
	return l
}

// SetLayout restores an arrangement. Unknown panel IDs are dropped; panels
// it does not mention stay where they are.
func (v *DockView) SetLayout(l DockLayout) bool {
	if l.Version < 0 || l.Version > 2 {
		return false
	}
	for _, size := range []float32{l.LeftSize, l.RightSize, l.BottomSize} {
		if !finiteNumber(float64(size)) {
			return false
		}
	}
	// Trees are authoritative when present; legacy lists migrate to one group.
	var trees [4]*DockNode
	for i, node := range []*DockNode{l.LeftTree, l.RightTree, l.BottomTree, l.CenterTree} {
		var ok bool
		trees[i], ok = v.restoreNode(node, map[*DockNode]bool{}, 0)
		if !ok {
			return false
		}
	}
	if l.LeftTree != nil {
		l.Left = dockPanels(trees[0])
	}
	if l.RightTree != nil {
		l.Right = dockPanels(trees[1])
	}
	if l.BottomTree != nil {
		l.Bottom = dockPanels(trees[2])
	}
	if l.CenterTree != nil {
		l.Center = dockPanels(trees[3])
	}
	placed := map[string]bool{}
	known := func(ids []string) ([]string, bool) {
		var out []string
		for _, id := range ids {
			if _, ok := v.panels[id]; !ok {
				continue
			}
			if placed[id] {
				return nil, false
			}
			placed[id] = true
			out = append(out, id)
		}
		return out, true
	}
	next := l
	next.Version = 2
	var ok bool
	if next.Left, ok = known(l.Left); !ok {
		return false
	}
	if next.Right, ok = known(l.Right); !ok {
		return false
	}
	if next.Bottom, ok = known(l.Bottom); !ok {
		return false
	}
	if next.Center, ok = known(l.Center); !ok {
		return false
	}
	keep := func(old []string) []string {
		var out []string
		for _, id := range old {
			if !placed[id] {
				out = append(out, id)
			}
		}
		return out
	}
	next.Left = append(next.Left, keep(v.layout.Left)...)
	next.Right = append(next.Right, keep(v.layout.Right)...)
	next.Bottom = append(next.Bottom, keep(v.layout.Bottom)...)
	next.Center = append(next.Center, keep(v.layout.Center)...)
	next.Detached = nil // windows are not restored: detached panels come back
	hidden := map[string]bool{}
	next.Hidden = nil
	for _, id := range l.Hidden {
		if _, ok := v.panels[id]; ok && !hidden[id] {
			hidden[id] = true
			next.Hidden = append(next.Hidden, id)
		}
	}
	// Unmentioned panels retain both their region and previous hidden state.
	for _, id := range v.layout.Hidden {
		if !placed[id] && !hidden[id] {
			hidden[id] = true
			next.Hidden = append(next.Hidden, id)
		}
	}
	if next.LeftSize <= 0 {
		next.LeftSize = v.layout.LeftSize
	}
	if next.RightSize <= 0 {
		next.RightSize = v.layout.RightSize
	}
	if next.BottomSize <= 0 {
		next.BottomSize = v.layout.BottomSize
	}
	for i, ids := range [][]string{next.Left, next.Right, next.Bottom, next.Center} {
		if trees[i] == nil && len(ids) > 0 {
			trees[i] = &DockNode{Panels: slices.Clone(ids), Active: []string{next.LeftActive, next.RightActive, next.BottomActive, next.CenterActive}[i]}
		} else if trees[i] != nil {
			for _, id := range ids {
				if !slices.Contains(dockPanels(trees[i]), id) {
					firstDockGroup(trees[i]).Panels = append(firstDockGroup(trees[i]).Panels, id)
				}
			}
		}
	}
	next.LeftTree, next.RightTree, next.BottomTree, next.CenterTree = trees[0], trees[1], trees[2], trees[3]
	if _, ok := v.panels[next.Zoomed]; !ok || hidden[next.Zoomed] {
		next.Zoomed = ""
	}
	v.cancelResize()
	v.drag = dockDrag{}
	v.layout = next
	v.splitSizes = map[*DockNode]float32{}
	v.menus = map[*DockNode]*MenuView{}
	for _, side := range dockSides {
		v.fixActive(side)
	}
	return true
}

// Visible reports whether a panel is shown in the dock: not closed and not
// detached into a window of its own.
func (v *DockView) Visible(id string) bool {
	_, ok := v.panels[id]
	return ok && !slices.Contains(v.layout.Hidden, id) && !slices.Contains(v.layout.Detached, id)
}

// SetVisible closes a panel or reopens it in the region it was last in (left
// if none).
func (v *DockView) SetVisible(id string, on bool) {
	if _, ok := v.panels[id]; !ok || on == v.Visible(id) {
		return
	}
	v.cancelResize()
	v.drag = dockDrag{}
	if on {
		v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
		if v.where(id) < 0 {
			v.Move(id, DockLeft)
			return
		}
		_, active := v.side(DockSide(v.where(id)))
		*active = id
		if n := findDockGroup(*v.tree(DockSide(v.where(id))), id); n != nil {
			n.Active = id
		}
		return
	}
	v.layout.Hidden = append(v.layout.Hidden, id)
	if v.layout.Zoomed == id {
		v.layout.Zoomed = ""
	}
	if s := v.where(id); s >= 0 {
		v.fixActive(DockSide(s))
	}
}

// Move puts a panel at the end of a region and makes it that region's active tab.
func (v *DockView) Move(id string, to DockSide) {
	if _, ok := v.panels[id]; !ok || to > DockCenter {
		return
	}
	v.cancelResize()
	v.drag = dockDrag{}
	if v.layout.Zoomed == id {
		v.layout.Zoomed = ""
	}
	for _, s := range dockSides {
		*v.tree(s) = removeDockPanel(*v.tree(s), id)
		ids, _ := v.side(s)
		*ids = slices.DeleteFunc(*ids, func(x string) bool { return x == id })
		v.fixActive(s)
	}
	ids, active := v.side(to)
	*ids = append(*ids, id)
	v.appendTree(to, id)
	findDockGroup(*v.tree(to), id).Active = id
	*active = id
	v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
	v.syncTrees()
}

func (v *DockView) side(s DockSide) (*[]string, *string) {
	switch s {
	case DockRight:
		return &v.layout.Right, &v.layout.RightActive
	case DockBottom:
		return &v.layout.Bottom, &v.layout.BottomActive
	case DockCenter:
		return &v.layout.Center, &v.layout.CenterActive
	}
	return &v.layout.Left, &v.layout.LeftActive
}

func (v *DockView) size(s DockSide) *float32 {
	switch s {
	case DockRight:
		return &v.layout.RightSize
	case DockBottom:
		return &v.layout.BottomSize
	case DockCenter:
		return &v.centerSize
	}
	return &v.layout.LeftSize
}

func (v *DockView) where(id string) int {
	for _, s := range dockSides {
		if ids, _ := v.side(s); slices.Contains(*ids, id) {
			return int(s)
		}
	}
	return -1
}

// shown lists a region's visible panels.
func (v *DockView) shown(s DockSide) []string {
	ids, _ := v.side(s)
	var out []string
	for _, id := range *ids {
		if v.Visible(id) {
			out = append(out, id)
		}
	}
	return out
}

func (v *DockView) fixActive(s DockSide) {
	v.fixNode(*v.tree(s))
	_, active := v.side(s)
	if shown := v.shown(s); !slices.Contains(shown, *active) {
		*active = ""
		if len(shown) > 0 {
			*active = shown[0]
		}
	}
}

func (v *DockView) changed() {
	if v.disabled {
		return
	}
	if v.onLayout != nil {
		v.onLayout(v.Layout())
	}
}

func (v *DockView) group(cx *el.Context, s DockSide, n *DockNode) el.Element {
	ids := v.nodeShown(n)
	if len(ids) == 0 {
		return nil
	}
	active := &n.Active
	text := locale.Current()
	tabs := el.Div().Role("tablist").Row().Grow().W(el.Dp(0)).ScrollX().Gap(theme.SpaceXxs)
	for _, id := range ids {
		id := id
		on := id == *active
		t := el.Div().ID(v.tabID(id)).NoShrink().Role("tab").Name(v.panels[id].Title).Selected(on).Px(10).Py(theme.SpaceSm).Rounded(theme.RadiusSm).TextSize(theme.TextMd).
			CursorPointer().Focusable(true).FocusStyle(func(st *el.Style) { st.BorderColor(theme.Primary) }).
			OnClick(func() {
				if *active != id {
					*active = id
					_, selected := v.side(s)
					*selected = id
					v.changed()
				}
			}).OnDoubleClick(func() { v.toggleZoom(id) }).Child(el.Text(v.panels[id].Title).MaxLines(1))
		if on {
			t.Bg(theme.Surface).TextColor(theme.PrimaryText)
		} else {
			t.TextColor(theme.Muted).Hover(func(st *el.Style) { st.Bg(theme.SubtleHover) })
		}
		t.OnDrag(func(e el.DragEvent) { v.tabDrag(id, e) }).Decorate(func(gtx core.C, draw func()) {
			v.tabRects[id] = dockGeometry(cx, gtx, t)
			origin, _ := cx.PaintGeometry()
			scale := gtx.Metric.PxPerDp
			if scale <= 0 {
				scale = 1
			}
			v.tabOrigins[id] = [2]float32{float32(origin.X) / scale, float32(origin.Y) / scale}
			draw()
		})
		tabs.Child(t)
	}
	m := v.menus[n]
	if m == nil {
		m = Menu()
		v.menus[n] = m
	}
	m.items = m.items[:0]
	cur := *active
	for _, to := range []struct {
		side  DockSide
		label string
	}{{DockLeft, text.DockLeft}, {DockRight, text.DockRight}, {DockBottom, text.DockBottom}, {DockCenter, text.DockCenter}} {
		if to.side != s {
			to := to
			m.Item(to.label, "", func() { v.Move(cur, to.side); v.changed() })
		}
	}
	if len(ids) > 1 {
		target := ids[0]
		if target == cur {
			target = ids[1]
		}
		m.Separator().Item(text.DockSplitRight, "", func() {
			if v.Split(cur, target, DockPlacementRight) {
				v.changed()
			}
		}).
			Item(text.DockSplitBelow, "", func() {
				if v.Split(cur, target, DockPlacementBottom) {
					v.changed()
				}
			})
	}
	zoom := text.DockZoom
	if v.layout.Zoomed == cur {
		zoom = text.DockRestore
	}
	m.Separator().Item(zoom, "", func() { v.toggleZoom(cur) })
	if v.onDetach != nil {
		m.Separator().Item(text.DockDetach, "", func() { v.Detach(cur) })
	}
	m.Separator().Item(text.Close, "", func() { v.SetVisible(cur, false); v.changed() })
	m.Trigger(Button("", m.Toggle).Name(text.Name(text.More, v.panels[cur].Title)).Icon(IconChevronDown).Variant(ButtonGhost).Size(24))
	head := el.Div().Row().Items(el.Center).Gap(theme.SpaceXs).Px(theme.SpaceXs).Py(theme.SpaceXs).Bg(theme.Subtle).Child(tabs, m.Render(cx))
	// A bounded body, not a scroll view: panels like Tree and Table fill it
	// and scroll themselves; wrap long plain content in a ScrollY element.
	body := el.Div().Grow().H(el.Dp(0)).Items(el.Stretch).P(theme.SpaceMd)
	if p := v.panels[cur]; p.View != nil {
		body.Child(p.View.Render(cx))
	}
	box := el.Div().ID(autoID("dock-group", n)).Role("region").Name(v.panels[cur].Title).NoShrink().Items(el.Stretch).Bg(theme.Surface).Child(head, body)
	return box.Grow().MinW(el.Dp(0)).MinH(el.Dp(0)).Decorate(func(gtx core.C, draw func()) { v.groupRects[n] = dockGeometry(cx, gtx, box); draw() })
}

// handle resizes region s; growing a right or bottom region means dragging
// towards the center, so the sign flips there.
func (v *DockView) handle(s DockSide) el.Element {
	cursor, sign := pointer.CursorColResize, float32(1)
	if s == DockRight {
		sign = -1
	}
	if s == DockBottom {
		cursor, sign = pointer.CursorRowResize, -1
	}
	h := el.Div().Focusable(true).FocusStyle(func(st *el.Style) { st.Bg(theme.Primary) }).Role("separator").Name(locale.Current().Resize).NoShrink().Bg(theme.Border).Cursor(cursor).
		Hover(func(st *el.Style) { st.Bg(theme.Primary) }).
		OnDrag(func(e el.DragEvent) {
			pos := e.X
			if s == DockBottom {
				pos = e.Y
			}
			switch e.Kind {
			case el.DragStart:
				v.grab = pos
				v.resizing = true
				v.resizeSide = s
				v.resizeStart = *v.size(s)
			case el.DragEnd:
				v.resizing = false
				if e.Canceled {
					*v.size(s) = v.resizeStart
				} else if *v.size(s) != v.resizeStart {
					v.changed()
				}
			default:
				*v.size(s) = max(80, v.painted[s]+sign*(pos-v.grab))
			}
		})
	h.OnKey(func(e el.KeyEvent) bool {
		if e.Modifiers != 0 {
			return false
		}
		delta := float32(0)
		next := *v.size(s)
		switch key.Name(e.Name) {
		case key.NameLeftArrow, key.NameUpArrow:
			delta = -10 * sign
		case key.NameRightArrow, key.NameDownArrow:
			delta = 10 * sign
		case key.NameHome:
			next = 80
		case key.NameEnd:
			axis := 0
			if s == DockBottom {
				axis = 1
			}
			next = max(80, v.total[axis]-120)
		default:
			return false
		}
		if e.State == el.KeyPress {
			old := *v.size(s)
			*v.size(s) = max(80, next+delta)
			if old != *v.size(s) {
				v.changed()
			}
		}
		return true
	})
	if s == DockBottom {
		return h.H(el.Dp(4))
	}
	return h.W(el.Dp(4))
}

// fitted is a region's size, shrunk so the center keeps at least 120dp when
// the window is too small for the regions as set.
func (v *DockView) fitted(s DockSide) float32 {
	sz := *v.size(s)
	if s == DockBottom {
		if h := v.total[1]; h > 0 && sz > h-120 {
			sz = max(h-124, 0)
		}
		return sz
	}
	l, r := float32(0), float32(0)
	if len(v.shown(DockLeft)) > 0 {
		l = v.layout.LeftSize
	}
	if len(v.shown(DockRight)) > 0 {
		r = v.layout.RightSize
	}
	if w := v.total[0]; w > 0 && l+r > w-128 {
		sz *= max(w-128, 0) / (l + r)
	}
	return sz
}

func (v *DockView) Render(cx *el.Context) el.Element {
	id := autoID("dock", v)
	width, height := cx.ViewportSize()
	viewport := [2]float32{width, height}
	if v.total == [2]float32{} || v.viewport != viewport {
		v.total, v.viewport = viewport, viewport
	}
	if (v.resizing || v.splitResize != nil || v.drag.id != "") && !cx.Enabled(id) {
		v.cancelResize()
		v.drag = dockDrag{}
	}
	if v.focusTab != "" {
		cx.Focus(v.tabID(v.focusTab))
		v.focusTab = ""
	}
	if z := v.layout.Zoomed; z != "" && v.Visible(z) && v.where(z) >= 0 && findDockGroup(*v.tree(DockSide(v.where(z))), z) != nil {
		// One panel over the whole dock; its menu, a double click on its
		// tab or Esc brings the others back.
		s := DockSide(v.where(z))
		n := findDockGroup(*v.tree(s), z)
		n.Active = z
		_, active := v.side(s)
		*active = z
		return el.Div().ID(id).Disabled(v.disabled).Row().Grow().Items(el.Stretch).OnKey(func(e el.KeyEvent) bool {
			if e.Name != string(key.NameEscape) {
				return false
			}
			if e.State == el.KeyPress {
				v.toggleZoom(z)
			}
			return true
		}).Child(v.group(cx, s, n))
	}
	middle := el.Div().Grow().W(el.Dp(0)).Items(el.Stretch)
	center := el.Div().Grow().H(el.Dp(0)).Items(el.Stretch)
	if docs := v.renderNode(cx, DockCenter, v.layout.CenterTree); docs != nil {
		center.Child(docs)
	} else if v.center != nil {
		center.Child(v.center.Render(cx))
	}
	center.Decorate(func(gtx core.C, draw func()) { v.centerRect = dockGeometry(cx, gtx, center); draw() })
	middle.Child(center)
	if b := v.region(cx, DockBottom); b != nil {
		middle.Child(v.handle(DockBottom), b)
	}
	row := el.Div().ID(id).Disabled(v.disabled).Row().Grow().Items(el.Stretch).OnKey(func(e el.KeyEvent) bool {
		if e.Name == string(key.NameEscape) && v.drag.id != "" {
			if e.State == el.KeyPress {
				v.drag = dockDrag{}
			}
			return true
		}
		return false
	}).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			actual := [2]float32{float32(gtx.Constraints.Max.X) / px, float32(gtx.Constraints.Max.Y) / px}
			if v.total != actual {
				v.total = actual
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		clear(v.groupRects)
		clear(v.tabRects)
		clear(v.tabOrigins)
		if px := gtx.Metric.PxPerDp; px > 0 {
			origin, _ := cx.PaintGeometry()
			v.bounds = dockRect{float32(origin.X) / px, float32(origin.Y) / px, float32(gtx.Constraints.Max.X) / px, float32(gtx.Constraints.Max.Y) / px}
		}
		draw()
		v.paintDrop(cx, gtx)
	})
	if l := v.region(cx, DockLeft); l != nil {
		row.Child(l, v.handle(DockLeft))
	}
	row.Child(middle)
	if r := v.region(cx, DockRight); r != nil {
		row.Child(v.handle(DockRight), r)
	}
	return row
}

// Zoom maximizes a shown panel over the whole dock, hiding the center and the
// other panels until Zoom(""), its menu, a double click on its tab or Esc.
// The zoomed panel is part of Layout.
func (v *DockView) Zoom(id string) {
	if id != "" && (!v.Visible(id) || v.where(id) < 0) {
		return
	}
	v.cancelResize()
	v.drag = dockDrag{}
	v.layout.Zoomed = id
}

// Zoomed is the maximized panel, or "".
func (v *DockView) Zoomed() string { return v.layout.Zoomed }

func (v *DockView) toggleZoom(id string) {
	if v.layout.Zoomed == id {
		v.Zoom("")
	} else {
		v.Zoom(id)
	}
	v.focusTab = id
	v.changed()
}

func (v *DockView) cancelResize() {
	if v.splitResize != nil {
		v.splitResize.Ratio = v.splitStart
		v.splitResize = nil
	}
	if v.resizing {
		*v.size(v.resizeSide) = v.resizeStart
		v.resizing = false
	}
}
func (v *DockView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.drag = dockDrag{}
		v.cancelResize()
		for _, m := range v.menus {
			m.SetValue(false)
		}
	}
}
