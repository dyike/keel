package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// VirtualListView scrolls through many rows of one height, building only the
// rows near the viewport, so a list of 100 000 rows costs what 30 do.
//
//	kit.VirtualList(len(logs), 24, func(cx *el.Context, i int) el.Element { return el.Text(logs[i]) })
type VirtualListView struct {
	count                         int
	rowH                          float32
	height                        float32
	fill                          bool
	horizontal, axisChanged       bool
	width, lastOffset, axisOffset float32
	row                           func(cx *el.Context, i int) el.Element
	itemKey                       func(int) string
	reveal                        int // row to scroll to once the viewport exists, -1 none
	revealAlign                   ScrollAlign
}

func VirtualList(count int, rowHeight float32, row func(cx *el.Context, i int) el.Element) *VirtualListView {
	return &VirtualListView{count: max(count, 0), rowH: max(rowHeight, 1), height: 320, row: row, reveal: -1}
}

// Height sets the viewport height in dp, 320 by default.
func (v *VirtualListView) Height(dp float32) *VirtualListView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
		if !v.horizontal {
			v.fill = false
		}
	}
	return v
}

// Fill makes the viewport grow to the space its parent gives it instead.
func (v *VirtualListView) Fill() *VirtualListView { v.fill = true; return v }
func (v *VirtualListView) SetCount(n int)         { v.count = max(n, 0) }
func (v *VirtualListView) Count() int             { return v.count }

// ItemKey uses stable, unique, nonempty data keys for row element identity.
// The default is the row index. The callback is called only for built rows.
func (v *VirtualListView) ItemKey(fn func(int) string) *VirtualListView { v.itemKey = fn; return v }

// ID is the scroll container's element ID.
func (v *VirtualListView) ID() string { return autoID("vlist", v) }

// ScrollTo scrolls as little as needed to show row i; call it from Render or
// a callback. A list that is not on screen yet (another tab) scrolls when it
// first shows.
func (v *VirtualListView) ScrollTo(cx *el.Context, i int) { v.ScrollToAlign(cx, i, ScrollNearest) }

// ScrollToAlign scrolls row i to the top, center or bottom of the viewport,
// as far as the content allows, or minimally with ScrollNearest.
func (v *VirtualListView) ScrollToAlign(cx *el.Context, i int, align ScrollAlign) {
	if i < 0 || i >= v.count {
		return
	}
	v.reveal, v.revealAlign = i, align
	v.applyReveal(cx)
}

func (v *VirtualListView) applyReveal(cx *el.Context) {
	if v.reveal < 0 {
		return
	}
	if _, view, _ := virtualState(cx, v.ID(), v.horizontal); view == 0 {
		// Not painted yet: a zero timer asks for the next frame, even with
		// reduced motion, where the viewport will exist.
		cx.After(revealKey{v.ID()}, 0, func() {})
		return
	}
	if v.count == 0 {
		v.reveal = -1
		return
	}
	i := min(v.reveal, v.count-1)
	start, end := float32(i)*v.rowH, float32(i+1)*v.rowH
	if v.revealAlign == ScrollNearest {
		virtualReveal(cx, v.ID(), v.horizontal, start, end)
	} else {
		off, view, _ := virtualState(cx, v.ID(), v.horizontal)
		virtualScroll(cx, v.ID(), v.horizontal, alignedOffset(v.revealAlign, off, view, float32(v.count)*v.rowH, start, end))
	}
	v.reveal = -1
}

// Range reports the rows built in the last Render, [first, last).
func (v *VirtualListView) visible(cx *el.Context) (first, last int) {
	off, view, _ := virtualState(cx, v.ID(), v.horizontal)
	if view == 0 {
		view = virtualExtent(v.horizontal, v.width, v.height)
	}
	if v.axisChanged && v.reveal < 0 {
		off = v.axisOffset
	}
	// Clamp against the current row count, not the previous frame's content.
	// Otherwise a shortened list retains an oversized leading spacer forever.
	off = min(max(off, 0), max(0, float32(v.count)*v.rowH-view))
	// Build a screen above and below as well: a wheel scroll is applied while
	// painting, after this tree is built, and must not reveal blank space.
	spare := int(view/v.rowH) + 1
	first = max(0, int(off/v.rowH)-spare)
	last = min(v.count, int((off+view)/v.rowH)+1+spare)
	return first, max(first, last)
}

func (v *VirtualListView) Render(cx *el.Context) el.Element {
	id := v.ID()
	if v.axisChanged && v.reveal >= 0 {
		v.axisOffset = float32(v.reveal) * v.rowH
	}
	v.applyReveal(cx)
	first, last := v.visible(cx)
	box := el.Div().ID(id).Items(el.Stretch)
	if v.horizontal {
		box.Row().ScrollX().H(el.Dp(v.height))
		if v.fill {
			box.Grow().MinW(el.Dp(v.rowH))
		} else {
			box.W(el.Dp(virtualExtent(true, v.width, v.height)))
		}
	} else {
		box.ScrollY()
		if v.width > 0 {
			box.W(el.Dp(v.width))
		}
		if v.fill {
			box.Grow().MinH(el.Dp(v.rowH))
		} else {
			box.H(el.Dp(v.height))
		}
	}
	box.Child(virtualSpacer(v.horizontal, float32(first)*v.rowH))
	for i := first; i < last; i++ {
		// Stable IDs keep each row's state as the window slides.
		key := strconv.Itoa(i)
		if v.itemKey != nil {
			key = v.itemKey(i)
		}
		item := el.Div().ID(id + "/" + key).NoShrink().Items(el.Stretch)
		if v.horizontal {
			item.W(el.Dp(v.rowH)).H(el.Full)
		} else {
			item.H(el.Dp(v.rowH))
		}
		if v.row != nil {
			item.Child(v.row(cx, i))
		}
		box.Child(item)
	}
	box.Child(virtualSpacer(v.horizontal, float32(v.count-last)*v.rowH))
	box.Decorate(func(gtx core.C, draw func()) {
		draw()
		if v.axisChanged && gtx.Enabled() {
			if v.reveal < 0 {
				virtualScroll(cx, id, v.horizontal, v.axisOffset)
			}
			v.axisChanged = false
		}
		v.lastOffset, _, _ = virtualState(cx, id, v.horizontal)
	})
	return box
}

type revealKey struct{ id string }
