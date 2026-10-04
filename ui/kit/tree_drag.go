package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Dragging a tree row shows where it will land: the top of a row inserts
// before it, the bottom after it, and the middle of a row with children
// moves into it. Resting over a collapsed row expands it; near the top or
// bottom of the list it scrolls.

const (
	treeDropBefore = -1
	treeDropInto   = 0
	treeDropAfter  = 1

	treeExpandDelay = 600 * time.Millisecond
	treeScrollZone  = 32  // dp from an edge where dragging scrolls
	treeScrollSpeed = 480 // dp per second at the very edge
)

type treeDrag struct {
	id           string
	start, at    float32 // pointer y in list content coordinates, dp
	moved, valid bool
	target, zone int // target row and treeDrop* zone
	expandID     string
	tick         int
	last         time.Time
}

// dragNode follows one drag of row i.
func (v *TreeView) dragNode(i int, e el.DragEvent) {
	if i < 0 || i >= len(v.rows) {
		return
	}
	switch e.Kind {
	case el.DragStart:
		if v.disabled || v.rows[i].node.Disabled {
			return
		}
		y := float32(i)*v.list.rowH + e.Y
		v.drag = &treeDrag{id: v.rows[i].node.ID, start: y, at: y, target: -1}
		return
	}
	d := v.drag
	if d == nil || d.id != v.rows[i].node.ID {
		return
	}
	if e.Canceled || v.disabled {
		v.drag = nil
		return
	}
	// The row moves with the content, so its index places the pointer.
	d.at = float32(v.index(d.id))*v.list.rowH + e.Y
	if abs32(d.at-d.start) >= 4 {
		d.moved = true
	}
	v.updateDrop(d)
	if e.Kind != el.DragEnd {
		return
	}
	v.drag = nil
	if !d.moved || !d.valid {
		return
	}
	parent, index := v.dropPlace(d)
	if d.zone == treeDropInto {
		v.open[parent] = true
	}
	if err := v.MoveNode(d.id, parent, index); err == nil && v.onReorder != nil {
		v.onReorder(d.id, parent, index)
	}
}

// updateDrop finds the row and zone under the pointer and whether the
// dragged node may land there.
func (v *TreeView) updateDrop(d *treeDrag) {
	d.valid, d.target = false, -1
	if !d.moved || len(v.rows) == 0 {
		return
	}
	t := min(max(int(d.at/v.list.rowH), 0), len(v.rows)-1)
	frac := d.at/v.list.rowH - float32(t)
	n := v.rows[t].node
	zone := treeDropAfter
	if frac < .5 {
		zone = treeDropBefore
	}
	if len(n.Children) > 0 || n.Lazy {
		switch {
		case frac < .25:
			zone = treeDropBefore
		case frac > .75:
			zone = treeDropAfter
		default:
			zone = treeDropInto
		}
	}
	d.target, d.zone = t, zone
	d.valid = v.canDrop(d)
}

func (v *TreeView) canDrop(d *treeDrag) bool {
	n := v.rows[d.target].node
	if n.ID == d.id || n.Disabled || v.within(n.ID, d.id) {
		return false
	}
	parent, _ := v.dropPlace(d)
	if parent != "" && (v.nodes[parent] == nil || v.nodes[parent].Disabled || v.within(parent, d.id)) {
		return false
	}
	return true
}

// within reports whether id is inside the subtree of ancestor.
func (v *TreeView) within(id, ancestor string) bool {
	for {
		_, _, parent := v.siblings(id)
		if parent == "" {
			return false
		}
		if parent == ancestor {
			return true
		}
		id = parent
	}
}

// dropPlace is the parent and final sibling index the drop moves to.
func (v *TreeView) dropPlace(d *treeDrag) (string, int) {
	n := v.rows[d.target].node
	if d.zone == treeDropInto {
		return n.ID, len(n.Children)
	}
	siblings, index, parent := v.siblings(n.ID)
	if d.zone == treeDropAfter {
		index++
	}
	if source, old, _ := v.siblings(d.id); source == siblings && old < index {
		index-- // MoveNode takes the final position, after removing the node
	}
	return parent, index
}

// dragFrame runs every Render while a drag is on: it expands a collapsed
// row the pointer rests on and scrolls near the list's edges.
func (v *TreeView) dragFrame(cx *el.Context) {
	d := v.drag
	if d == nil || !d.moved || v.disabled {
		return
	}
	if d.target >= 0 && d.zone == treeDropInto && d.valid {
		n := v.rows[d.target].node
		if !v.open[n.ID] {
			if d.expandID != n.ID {
				d.expandID = n.ID
			}
			cx.AfterEnabled(autoID("tree", v), struct {
				Drag *treeDrag
				ID   string
			}{d, n.ID}, treeExpandDelay, func() {
				if v.drag == d && d.expandID == n.ID {
					v.expand(n.ID, true, true)
				}
			})
		}
	} else {
		d.expandID = ""
	}
	id := v.list.ID()
	offset, viewport, content := cx.ScrollState(id)
	y := d.at - offset
	speed := float32(0)
	if zone := min(float32(treeScrollZone), viewport/3); zone > 0 {
		if y < zone {
			speed = -treeScrollSpeed * min(1, (zone-y)/zone)
		} else if y > viewport-zone {
			speed = treeScrollSpeed * min(1, (y-viewport+zone)/zone)
		}
	}
	if speed == 0 || speed < 0 && offset <= 0 || speed > 0 && offset >= content-viewport {
		d.last = time.Time{}
		return
	}
	if d.last.IsZero() {
		d.last = cx.Now()
	}
	// Keep scrolling while the pointer rests; the pointer's content position
	// moves with the scroll, so the drop target follows.
	cx.AfterEnabled(id, struct {
		Drag *treeDrag
		Tick int
	}{d, d.tick}, 16*time.Millisecond, func() {
		if v.drag != d || v.disabled {
			return
		}
		dt := min(cx.Now().Sub(d.last), 50*time.Millisecond)
		d.last = cx.Now()
		d.tick++
		step := speed * float32(dt.Seconds())
		next := min(max(offset+step, 0), content-viewport)
		d.at += next - offset
		cx.ScrollTo(id, next)
		v.updateDrop(d)
	})
}

// dropMarker shows where a drop on row i lands, or nothing.
func (v *TreeView) dropMarker(row *el.DivEl, i int) {
	d := v.drag
	if d == nil || !d.moved || !d.valid || d.target != i {
		return
	}
	switch d.zone {
	case treeDropInto:
		row.Bg(theme.Highlight).Border(1, theme.Primary)
	case treeDropBefore:
		row.Child(el.Div().Absolute().Top(0).Left(0).Right(0).H(el.Dp(2)).Bg(theme.Primary))
	default:
		row.Child(el.Div().Absolute().Bottom(0).Left(0).Right(0).H(el.Dp(2)).Bg(theme.Primary))
	}
}
