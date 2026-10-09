package kit

import (
	"image"
	"slices"
	"time"

	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
)

type tableHeaderGeometry struct {
	origin f32.Point
	bounds struct{ Min, Max f32.Point }
}
type tableColumnDrag struct {
	column, target, position int
	start, pointer           f32.Point
	tick                     int
	last                     time.Time
	moved, valid, after      bool
}

// OnColumnMove runs after a header drop changes the display order. Column is a
// source index; from/to are display positions including hidden columns.
// MoveColumn and SetLayoutState do not invoke it.
func (v *TableView) OnColumnMove(fn func(column, from, to int)) *TableView {
	v.onColumnMove = fn
	return v
}

func (v *TableView) canMoveColumn(column, position int) bool {
	from := slices.Index(v.columns, column)
	if from < 0 || position < 0 || position >= len(v.columns) {
		return false
	}
	for _, c := range v.columns[min(from, position) : max(from, position)+1] {
		if v.cols[c].noMove {
			return false
		}
	}
	return true
}

func (v *TableView) recordHeaderGeometry(cx *el.Context, c int, cell el.Element) {
	origin, viewport := cx.PaintGeometry()
	scale := cx.PixelScale()
	w, h := cx.LayoutSize(cell)
	rect := image.Rect(origin.X, origin.Y, origin.X+int(w*scale+.5), origin.Y+int(h*scale+.5)).Intersect(viewport)
	v.headerGeometry[c] = tableHeaderGeometry{
		origin: f32.Pt(float32(origin.X)/scale, float32(origin.Y)/scale),
		bounds: struct{ Min, Max f32.Point }{Min: f32.Pt(float32(rect.Min.X)/scale, float32(rect.Min.Y)/scale), Max: f32.Pt(float32(rect.Max.X)/scale, float32(rect.Max.Y)/scale)},
	}
}

func (v *TableView) dragColumn(c int, e el.DragEvent) {
	geom, ok := v.headerGeometry[c]
	if !ok || v.disabled || v.hidden[c] || v.cols[c].noMove {
		v.columnDrag = nil
		return
	}
	p := geom.origin.Add(f32.Pt(e.X, e.Y))
	if e.Kind == el.DragStart {
		v.columnDrag = &tableColumnDrag{column: c, start: p, pointer: p, target: -1}
		return
	}
	drag := v.columnDrag
	if drag == nil || drag.column != c {
		return
	}
	if e.Canceled {
		v.columnDrag = nil
		return
	}
	if abs32(p.X-drag.start.X) >= 3 {
		drag.moved = true
	}
	drag.pointer = p
	v.updateColumnDrop(drag)
	if e.Kind != el.DragEnd {
		return
	}
	v.columnDrag = nil
	if !drag.valid {
		return
	}
	from := slices.Index(v.columns, c)
	v.MoveColumn(c, drag.position)
	if v.onColumnMove != nil {
		v.onColumnMove(c, from, drag.position)
	}
}

func (v *TableView) updateColumnDrop(drag *tableColumnDrag) {
	p := drag.pointer
	drag.valid = false
	if drag.moved {
		// Frozen headers paint over the scrolling strip, so prefer their hit areas.
		visible := v.visibleColumns()
		left, right := v.frozenCounts(visible)
		hits := append([]int{}, visible[:left]...)
		hits = append(hits, visible[len(visible)-right:]...)
		hits = append(hits, visible[left:len(visible)-right]...)
		for _, target := range hits {
			g, ok := v.headerGeometry[target]
			if !ok || p.X < g.bounds.Min.X || p.X >= g.bounds.Max.X || p.Y < g.bounds.Min.Y || p.Y >= g.bounds.Max.Y {
				continue
			}
			after := p.X >= g.origin.X+v.widths[target]/2
			position := slices.Index(v.columns, target)
			if after {
				position++
			}
			from := slices.Index(v.columns, drag.column)
			if position > from {
				position--
			}
			drag.target, drag.after, drag.position = target, after, position
			drag.valid = position != from && v.canMoveColumn(drag.column, position)
			break
		}
	}
}

// Use the unfrozen header strip as the scroll zone. Keeping the pointer still
// must keep scrolling; timer ownership stops ticks when the table is hidden.
func (v *TableView) scrollColumnDrag(cx *el.Context) {
	d := v.columnDrag
	if d == nil || !d.moved || v.disabled {
		return
	}
	left, right := float32(1e9), float32(-1e9)
	top, bottom := float32(1e9), float32(-1e9)
	for _, g := range v.headerGeometry {
		if g.bounds.Max.X <= g.bounds.Min.X {
			continue
		}
		left = min(left, g.bounds.Min.X)
		right = max(right, g.bounds.Max.X)
		top = min(top, g.bounds.Min.Y)
		bottom = max(bottom, g.bounds.Max.Y)
	}
	cols := v.visibleColumns()
	l, r := v.frozenCounts(cols)
	for _, c := range cols[:l] {
		left = max(left, v.headerGeometry[c].bounds.Max.X)
	}
	for _, c := range cols[len(cols)-r:] {
		right = min(right, v.headerGeometry[c].bounds.Min.X)
	}
	p := d.pointer
	if p.Y < top || p.Y >= bottom || p.X < left || p.X >= right || right <= left {
		d.last = time.Time{}
		return
	}
	zone := min(float32(32), (right-left)/3)
	speed := float32(0)
	if p.X < left+zone {
		speed = -360 * (left + zone - p.X) / zone
	}
	if p.X > right-zone {
		speed = 360 * (p.X - right + zone) / zone
	}
	id := autoID("table", v)
	offset, viewport, content := cx.ScrollStateX(id)
	if speed == 0 || speed < 0 && offset <= 0 || speed > 0 && offset >= content-viewport {
		d.last = time.Time{}
		return
	}
	if d.last.IsZero() {
		d.last = cx.Now()
	}
	cx.AfterEnabled(id, struct {
		Drag *tableColumnDrag
		Tick int
	}{d, d.tick}, 16*time.Millisecond, func() {
		if v.columnDrag != d || v.disabled {
			return
		}
		dt := min(cx.Now().Sub(d.last), 50*time.Millisecond)
		d.last = cx.Now()
		d.tick++
		cx.ScrollToX(id, offset+speed*float32(dt.Seconds()))
	})
}
