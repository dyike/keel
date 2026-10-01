package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/el"
)

// VirtualListView scrolls through many rows of one height, building only the
// rows near the viewport, so a list of 100 000 rows costs what 30 do.
//
//	kit.VirtualList(len(logs), 24, func(cx *el.Context, i int) el.Element { return el.Text(logs[i]) })
type VirtualListView struct {
	count  int
	rowH   float32
	height float32
	fill   bool
	row    func(cx *el.Context, i int) el.Element
	reveal int // row to scroll to once the viewport exists, -1 none
}

func VirtualList(count int, rowHeight float32, row func(cx *el.Context, i int) el.Element) *VirtualListView {
	return &VirtualListView{count: max(count, 0), rowH: max(rowHeight, 1), height: 320, row: row, reveal: -1}
}

// Height sets the viewport height in dp, 320 by default.
func (v *VirtualListView) Height(dp float32) *VirtualListView {
	if dp > 0 {
		v.height, v.fill = dp, false
	}
	return v
}

// Fill makes the viewport grow to the space its parent gives it instead.
func (v *VirtualListView) Fill() *VirtualListView { v.fill = true; return v }
func (v *VirtualListView) SetCount(n int)         { v.count = max(n, 0) }
func (v *VirtualListView) Count() int             { return v.count }

// ID is the scroll container's element ID.
func (v *VirtualListView) ID() string { return autoID("vlist", v) }

// ScrollTo scrolls as little as needed to show row i; call it from Render or
// a callback. A list that is not on screen yet (another tab) scrolls when it
// first shows.
func (v *VirtualListView) ScrollTo(cx *el.Context, i int) {
	if i < 0 || i >= v.count {
		return
	}
	v.reveal = i
	v.applyReveal(cx)
}

func (v *VirtualListView) applyReveal(cx *el.Context) {
	if v.reveal < 0 {
		return
	}
	if _, view, _ := cx.ScrollState(v.ID()); view == 0 {
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
	cx.ScrollIntoView(v.ID(), float32(i)*v.rowH, float32(i+1)*v.rowH)
	v.reveal = -1
}

// Range reports the rows built in the last Render, [first, last).
func (v *VirtualListView) visible(cx *el.Context) (first, last int) {
	off, view, _ := cx.ScrollState(v.ID())
	if view == 0 {
		view = v.height
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
	v.applyReveal(cx)
	first, last := v.visible(cx)
	box := el.Div().ID(id).ScrollY().Items(el.Stretch)
	if v.fill {
		box.Grow().MinH(el.Dp(v.rowH))
	} else {
		box.H(el.Dp(v.height))
	}
	box.Child(el.Div().H(el.Dp(float32(first) * v.rowH)))
	for i := first; i < last; i++ {
		// Stable IDs keep each row's state as the window slides.
		box.Child(el.Div().ID(id + "/" + strconv.Itoa(i)).H(el.Dp(v.rowH)).NoShrink().Items(el.Stretch).Child(v.row(cx, i)))
	}
	box.Child(el.Div().H(el.Dp(float32(v.count-last) * v.rowH)))
	return box
}

type revealKey struct{ id string }
