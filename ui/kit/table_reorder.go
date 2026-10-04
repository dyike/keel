package kit

import (
	"image"
	"slices"

	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
)

type tableHeaderGeometry struct {
	origin f32.Point
	bounds struct{ Min, Max f32.Point }
}
type tableColumnDrag struct {
	column, target, position int
	start                    f32.Point
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
		v.columnDrag = &tableColumnDrag{column: c, start: p, target: -1}
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
			from := slices.Index(v.columns, c)
			if position > from {
				position--
			}
			drag.target, drag.after, drag.position = target, after, position
			drag.valid = position != from && v.canMoveColumn(c, position)
			break
		}
	}
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
