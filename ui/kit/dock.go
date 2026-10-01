package kit

import (
	"slices"

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
)

// DockPanel is a tool window that lives in a Dock region. IDs must be unique.
// The panel gets the region's full height; long content should scroll itself.
type DockPanel struct {
	ID, Title string
	View      el.View
}

// DockLayout is everything about a Dock's arrangement, for saving and
// restoring it (it encodes as JSON). Panel lists are in tab order.
type DockLayout struct {
	Left, Right, Bottom             []string `json:",omitempty"`
	LeftActive, RightActive         string   `json:",omitempty"`
	BottomActive                    string   `json:",omitempty"`
	LeftSize, RightSize, BottomSize float32
	Hidden                          []string `json:",omitempty"`
}

// DockView arranges tool panels around a central view, like an IDE. Each
// region holds tabbed panels and resizes against the center; a panel's menu
// moves it to another region or closes it. Layout and SetLayout save and
// restore the arrangement.
type DockView struct {
	center   el.View
	panels   map[string]DockPanel
	layout   DockLayout
	painted  [3]float32
	total    [2]float32 // painted width and height of the whole dock, dp
	grab     float32
	menus    [3]*MenuView
	onLayout func(DockLayout)
}

func Dock(center el.View) *DockView {
	return &DockView{center: center, panels: map[string]DockPanel{},
		layout: DockLayout{LeftSize: 240, RightSize: 260, BottomSize: 180},
		menus:  [3]*MenuView{Menu(), Menu(), Menu()}}
}

// Panel adds a panel to a region; the first panel added to a region is active.
func (v *DockView) Panel(p DockPanel, side DockSide) *DockView {
	v.panels[p.ID] = p
	ids, active := v.side(side)
	*ids = append(*ids, p.ID)
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
	l.Left, l.Right, l.Bottom, l.Hidden = slices.Clone(l.Left), slices.Clone(l.Right), slices.Clone(l.Bottom), slices.Clone(l.Hidden)
	return l
}

// SetLayout restores an arrangement. Unknown panel IDs are dropped; panels
// it does not mention stay where they are.
func (v *DockView) SetLayout(l DockLayout) {
	known := func(ids []string) []string {
		return slices.DeleteFunc(slices.Clone(ids), func(id string) bool { _, ok := v.panels[id]; return !ok })
	}
	mentioned := map[string]bool{}
	for _, ids := range [][]string{l.Left, l.Right, l.Bottom, l.Hidden} {
		for _, id := range ids {
			mentioned[id] = true
		}
	}
	keep := func(ids []string) []string {
		var out []string
		for _, id := range ids {
			if !mentioned[id] {
				out = append(out, id)
			}
		}
		return out
	}
	old := v.layout
	v.layout = l
	v.layout.Left = append(known(l.Left), keep(old.Left)...)
	v.layout.Right = append(known(l.Right), keep(old.Right)...)
	v.layout.Bottom = append(known(l.Bottom), keep(old.Bottom)...)
	v.layout.Hidden = append(known(l.Hidden), keep(old.Hidden)...)
	for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
		v.fixActive(s)
	}
	if l.LeftSize <= 0 {
		v.layout.LeftSize = old.LeftSize
	}
	if l.RightSize <= 0 {
		v.layout.RightSize = old.RightSize
	}
	if l.BottomSize <= 0 {
		v.layout.BottomSize = old.BottomSize
	}
}

// Visible reports whether a panel is shown (not closed).
func (v *DockView) Visible(id string) bool { return !slices.Contains(v.layout.Hidden, id) }

// SetVisible closes a panel or reopens it in the region it was last in (left
// if none).
func (v *DockView) SetVisible(id string, on bool) {
	if _, ok := v.panels[id]; !ok || on == v.Visible(id) {
		return
	}
	if on {
		v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
		if v.where(id) < 0 {
			v.Move(id, DockLeft)
			return
		}
		_, active := v.side(DockSide(v.where(id)))
		*active = id
		return
	}
	v.layout.Hidden = append(v.layout.Hidden, id)
	if s := v.where(id); s >= 0 {
		v.fixActive(DockSide(s))
	}
}

// Move puts a panel at the end of a region and makes it that region's active tab.
func (v *DockView) Move(id string, to DockSide) {
	if _, ok := v.panels[id]; !ok {
		return
	}
	for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
		ids, _ := v.side(s)
		*ids = slices.DeleteFunc(*ids, func(x string) bool { return x == id })
		v.fixActive(s)
	}
	ids, active := v.side(to)
	*ids = append(*ids, id)
	*active = id
	v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
}

func (v *DockView) side(s DockSide) (*[]string, *string) {
	switch s {
	case DockRight:
		return &v.layout.Right, &v.layout.RightActive
	case DockBottom:
		return &v.layout.Bottom, &v.layout.BottomActive
	}
	return &v.layout.Left, &v.layout.LeftActive
}

func (v *DockView) size(s DockSide) *float32 {
	switch s {
	case DockRight:
		return &v.layout.RightSize
	case DockBottom:
		return &v.layout.BottomSize
	}
	return &v.layout.LeftSize
}

func (v *DockView) where(id string) int {
	for _, s := range []DockSide{DockLeft, DockRight, DockBottom} {
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
	_, active := v.side(s)
	if shown := v.shown(s); !slices.Contains(shown, *active) {
		*active = ""
		if len(shown) > 0 {
			*active = shown[0]
		}
	}
}

func (v *DockView) changed() {
	if v.onLayout != nil {
		v.onLayout(v.Layout())
	}
}

func (v *DockView) region(cx *el.Context, s DockSide) el.Element {
	ids := v.shown(s)
	if len(ids) == 0 {
		return nil
	}
	_, active := v.side(s)
	text := locale.Current()
	tabs := el.Div().Role("tablist").Row().Grow().W(el.Dp(0)).Gap(2)
	for _, id := range ids {
		id := id
		on := id == *active
		t := el.Div().Role("tab").Name(v.panels[id].Title).Selected(on).Px(10).Py(6).Rounded(4).TextSize(13).
			CursorPointer().Focusable(true).FocusStyle(func(st *el.Style) { st.BorderColor(theme.Primary) }).
			OnClick(func() { *active = id; v.changed() }).Child(el.Text(v.panels[id].Title).MaxLines(1))
		if on {
			t.Bg(theme.Surface).TextColor(theme.PrimaryText)
		} else {
			t.TextColor(theme.Muted).Hover(func(st *el.Style) { st.Bg(theme.SubtleHover) })
		}
		tabs.Child(t)
	}
	m := v.menus[s]
	m.items = m.items[:0]
	cur := *active
	for _, to := range []struct {
		side  DockSide
		label string
	}{{DockLeft, text.DockLeft}, {DockRight, text.DockRight}, {DockBottom, text.DockBottom}} {
		if to.side != s {
			to := to
			m.Item(to.label, "", func() { v.Move(cur, to.side); v.changed() })
		}
	}
	m.Separator().Item(text.Close, "", func() { v.SetVisible(cur, false); v.changed() })
	m.Trigger(Button("", m.Toggle).Name(text.Name(text.More, v.panels[cur].Title)).Icon(IconChevronDown).Variant(ButtonGhost).Size(24))
	head := el.Div().Row().Items(el.Center).Gap(4).Px(4).Py(4).Bg(theme.Subtle).Child(tabs, m.Render(cx))
	// A bounded body, not a scroll view: panels like Tree and Table fill it
	// and scroll themselves; wrap long plain content in a ScrollY element.
	body := el.Div().Grow().H(el.Dp(0)).Items(el.Stretch).P(8)
	if p := v.panels[cur]; p.View != nil {
		body.Child(p.View.Render(cx))
	}
	box := el.Div().Role("region").Name(v.panels[cur].Title).NoShrink().Items(el.Stretch).Bg(theme.Surface).Child(head, body)
	sz := v.fitted(s)
	if s == DockBottom {
		box.H(el.Dp(sz))
	} else {
		box.W(el.Dp(sz))
	}
	return box.Decorate(func(gtx core.C, draw func()) {
		v.painted[s] = sz
		draw()
	})
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
	h := el.Div().Role("separator").Name(locale.Current().Resize).NoShrink().Bg(theme.Border).Cursor(cursor).
		Hover(func(st *el.Style) { st.Bg(theme.Primary) }).
		OnDrag(func(e el.DragEvent) {
			pos := e.X
			if s == DockBottom {
				pos = e.Y
			}
			switch e.Kind {
			case el.DragStart:
				v.grab = pos
			case el.DragEnd:
				v.changed()
			default:
				*v.size(s) = max(80, v.painted[s]+sign*(pos-v.grab))
			}
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
			sz = max(h-120, 40)
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
		sz *= max(w-128, 80) / (l + r)
	}
	return sz
}

func (v *DockView) Render(cx *el.Context) el.Element {
	middle := el.Div().Grow().W(el.Dp(0)).Items(el.Stretch)
	center := el.Div().Grow().H(el.Dp(0)).Items(el.Stretch)
	if v.center != nil {
		center.Child(v.center.Render(cx))
	}
	middle.Child(center)
	if b := v.region(cx, DockBottom); b != nil {
		middle.Child(v.handle(DockBottom), b)
	}
	row := el.Div().Row().Grow().Items(el.Stretch).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			v.total = [2]float32{float32(gtx.Constraints.Max.X) / px, float32(gtx.Constraints.Max.Y) / px}
		}
		draw()
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
