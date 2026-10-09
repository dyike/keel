package kit

import (
	"slices"

	"gioui.org/op"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// StatusItem is one entry of a status bar that may move into its overflow
// menu when the bar is too narrow.
type StatusItem struct {
	// Label names the item in the overflow menu, and is the text shown in
	// the bar when View is nil.
	Label string
	// View is shown in the bar; nil shows Label.
	View el.View
	// Action runs when the item is clicked, in the bar (for a Label-only
	// item) or in the overflow menu.
	Action func()
	// Priority orders hiding: lower priorities move to the menu first, and
	// later items before earlier ones at the same priority.
	Priority int
}

// StatusBarView is a single-line 24dp status strip; its parent places it.
// Items added with Add move into a "…" menu, lowest priority first, when
// they do not fit; views given to Left and Right always stay.
type StatusBarView struct {
	left, right []el.View
	items       []statusEntry
	more        *MenuView
	avail       float32    // dp available to the bar's content
	pinned      [2]float32 // dp the Left and Right views take
}

type statusEntry struct {
	StatusItem
	right bool
	width float32 // last measured width in dp, 0 before the first frame
}

func StatusBar() *StatusBarView { return &StatusBarView{more: Menu()} }
func (v *StatusBarView) Left(views ...el.View) *StatusBarView {
	v.left = append([]el.View(nil), views...)
	return v
}
func (v *StatusBarView) Right(views ...el.View) *StatusBarView {
	v.right = append([]el.View(nil), views...)
	return v
}

// Add appends items to the left group, after the Left views.
func (v *StatusBarView) Add(items ...StatusItem) *StatusBarView {
	for _, it := range items {
		v.items = append(v.items, statusEntry{StatusItem: it})
	}
	return v
}

// AddRight appends items to the right group, before the Right views.
func (v *StatusBarView) AddRight(items ...StatusItem) *StatusBarView {
	for _, it := range items {
		v.items = append(v.items, statusEntry{StatusItem: it, right: true})
	}
	return v
}

// Hidden returns the labels of the items now in the overflow menu.
func (v *StatusBarView) Hidden() []string {
	var out []string
	for i, h := range v.hidden() {
		if h {
			out = append(out, v.items[i].Label)
		}
	}
	return out
}

const statusMoreWidth = 28

// hidden decides which items move to the menu from the measured widths.
func (v *StatusBarView) hidden() []bool {
	hide := make([]bool, len(v.items))
	if v.avail <= 0 {
		return hide
	}
	used := v.pinned[0] + v.pinned[1]
	for _, it := range v.items {
		used += it.width + 8
	}
	if used <= v.avail {
		return hide
	}
	used += statusMoreWidth
	order := make([]int, len(v.items))
	for i := range order {
		order[i] = i
	}
	// Lowest priority first; among equals, the later item first.
	slices.SortStableFunc(order, func(a, b int) int {
		if v.items[a].Priority != v.items[b].Priority {
			return v.items[a].Priority - v.items[b].Priority
		}
		return b - a
	})
	n := 0
	for ; n < len(order) && used > v.avail; n++ {
		hide[order[n]] = true
		used -= v.items[order[n]].width + 8
	}
	// Hiding a wide item may leave room for a narrower one hidden before it.
	for k := n - 2; k >= 0; k-- {
		if i := order[k]; used+v.items[i].width+8 <= v.avail {
			hide[i] = false
			used += v.items[i].width + 8
		}
	}
	return hide
}

func (v *StatusBarView) Render(cx *el.Context) el.Element {
	id := autoID("statusbar", v)
	hide := v.hidden()
	// The groups shrink so long pinned text is cut off rather than pushing
	// the bar wider; overflowing items keep their width and move instead.
	left := el.Div().Row().Gap(theme.SpaceMd).MaxLines(1).Items(el.Center).MinW(el.Dp(0))
	right := el.Div().Row().Gap(theme.SpaceMd).MaxLines(1).Items(el.Center).MinW(el.Dp(0))
	if len(v.left) > 0 {
		left.Child(v.pinnedGroup(cx, id+"/left", v.left, 0))
	} else {
		v.pinned[0] = 0
	}
	for i := range v.items {
		if hide[i] {
			continue
		}
		group := left
		if v.items[i].right {
			group = right
		}
		group.Child(v.itemElement(cx, id, i))
	}
	if len(v.right) > 0 {
		right.Child(v.pinnedGroup(cx, id+"/right", v.right, 1))
	} else {
		v.pinned[1] = 0
	}
	row := el.Div().ID(id + "/row").W(el.Full).H(el.Full).Row().Px(theme.SpaceMd).Gap(theme.SpaceMd).Items(el.Center).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			if w := float32(gtx.Constraints.Max.X)/px - 16; v.avail != w {
				v.avail = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
	})
	row.Child(left, el.Div().Grow().MinW(el.Dp(0)))
	if slices.Contains(hide, true) {
		v.more.items = nil
		for i, h := range hide {
			if h {
				it := v.items[i]
				v.more.Item(it.Label, "", it.Action)
			}
		}
		v.more.Trigger(Button("", v.more.Toggle).Name(locale.Current().More).Icon(IconMore).Variant(ButtonGhost).Size(20))
		row.Child(v.more.Render(cx))
	}
	row.Child(right)
	return el.Div().W(el.Full).H(el.Dp(24)).Role("status").TextSize(theme.TextSm).TextColor(theme.Muted).Bg(theme.Surface).Child(
		el.Div().Absolute().Top(0).Left(0).Right(0).H(el.Dp(1)).Bg(theme.Border),
		row,
	)
}

// pinnedGroup shows the Left or Right views, which never overflow, and
// records their width.
func (v *StatusBarView) pinnedGroup(cx *el.Context, id string, views []el.View, side int) el.Element {
	g := el.Div().ID(id).Row().Gap(theme.SpaceMd).Items(el.Center).MinW(el.Dp(0))
	for _, e := range views {
		if e != nil {
			g.Child(e.Render(cx))
		}
	}
	return g.Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			if w := float32(gtx.Constraints.Max.X) / px; v.pinned[side] != w {
				v.pinned[side] = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
	})
}

// itemElement shows one item and records its width.
func (v *StatusBarView) itemElement(cx *el.Context, id string, i int) el.Element {
	it := v.items[i]
	var content el.Element
	if it.View != nil {
		content = it.View.Render(cx)
	} else if it.Action != nil {
		content = el.Div().Role("button").Name(it.Label).Px(theme.SpaceXs).Rounded(theme.RadiusSm).Focusable(true).
			Hover(func(s *el.Style) { s.Bg(theme.Subtle).TextColor(theme.Text) }).OnClick(it.Action).Child(el.Text(it.Label))
	} else {
		content = el.Text(it.Label)
	}
	return el.Div().ID(id + "/item/" + it.Label).NoShrink().Child(content).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 && i < len(v.items) {
			if w := float32(gtx.Constraints.Max.X) / px; v.items[i].width != w {
				v.items[i].width = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
	})
}
