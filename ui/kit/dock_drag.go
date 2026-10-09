package kit

import (
	"image"
	"math"
	"slices"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type dockRect struct{ x, y, w, h float32 }

func (r dockRect) contains(x, y float32) bool {
	return r.w > 0 && r.h > 0 && x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

type dockDrop struct {
	kind      uint8 // 1: tab, 2: split, 3: outer region
	target    string
	after     bool
	side      DockSide
	placement DockPlacement
	preview   dockRect
}
type dockDrag struct {
	id    string
	x, y  float32
	moved bool
	drop  dockDrop
}

func dockGeometry(cx *el.Context, gtx core.C, e el.Element) dockRect {
	origin, viewport := cx.PaintGeometry()
	scale := gtx.Metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	w, h := cx.LayoutSize(e)
	r := image.Rect(origin.X, origin.Y, origin.X+int(w*scale), origin.Y+int(h*scale)).Intersect(viewport)
	return dockRect{float32(r.Min.X) / scale, float32(r.Min.Y) / scale, float32(r.Dx()) / scale, float32(r.Dy()) / scale}
}
func (v *DockView) tabDrag(id string, e el.DragEvent) {
	rect, ok := v.tabRects[id]
	if !ok {
		if v.drag.id == id {
			v.drag = dockDrag{}
		}
		return
	}
	// Pointer coordinates are relative to the unclipped tab, unlike drop bounds.
	origin := v.tabOrigins[id]
	x, y := origin[0]+e.X, origin[1]+e.Y
	if e.Kind == el.DragStart {
		if !rect.contains(x, y) {
			return
		}
		v.drag = dockDrag{id: id, x: x, y: y}
		return
	}
	if v.drag.id != id {
		return
	}
	if !v.drag.moved && float32(math.Hypot(float64(x-v.drag.x), float64(y-v.drag.y))) >= 6 {
		v.drag.moved = true
	}
	if v.drag.moved {
		v.drag.drop = v.dropAt(id, x, y)
	}
	if e.Kind != el.DragEnd {
		return
	}
	drag := v.drag
	v.drag = dockDrag{}
	if e.Canceled || v.disabled || !drag.moved {
		return
	}
	// Dropped outside the dock: into a window of its own, when the app allows.
	if drag.drop.kind == 0 && v.onDetach != nil && v.bounds.w > 0 && !v.bounds.contains(x, y) {
		v.Detach(id)
		return
	}
	changed := false
	switch drag.drop.kind {
	case 1:
		changed = v.joinTab(id, drag.drop.target, drag.drop.after)
	case 2:
		changed = v.Split(id, drag.drop.target, drag.drop.placement)
	case 3:
		v.Move(id, drag.drop.side)
		changed = true
	}
	if changed {
		v.focusTab = id
		v.changed()
	}
}
func (v *DockView) joinTab(id, target string, after bool) bool {
	if id == target || v.where(id) < 0 || v.where(target) < 0 {
		return false
	}
	side := DockSide(v.where(target))
	source := findDockGroup(*v.tree(DockSide(v.where(id))), id)
	dest := findDockGroup(*v.tree(side), target)
	old := slices.Clone(dest.Panels)
	// The target keeps the destination leaf alive while removing the source.
	for _, s := range dockSides {
		*v.tree(s) = removeDockPanel(*v.tree(s), id)
	}
	index := slices.Index(dest.Panels, target)
	if after {
		index++
	}
	dest.Panels = slices.Insert(dest.Panels, index, id)
	wasActive := dest.Active == id
	dest.Active = id
	v.layout.Hidden = slices.DeleteFunc(v.layout.Hidden, func(s string) bool { return s == id })
	v.syncTrees()
	_, active := v.side(side)
	*active = id
	return source != dest || !slices.Equal(old, dest.Panels) || !wasActive
}
func (v *DockView) dropAt(id string, x, y float32) dockDrop {
	// Documents cover the center, so only a thin strip along its edges
	// brings back an empty side region there.
	if c := v.centerRect; len(v.shown(DockCenter)) > 0 && c.contains(x, y) {
		const strip = 24
		switch {
		case x-c.x < strip && len(v.shown(DockLeft)) == 0:
			return v.regionDrop(id, DockLeft, dockRect{c.x, c.y, min(c.w/2, 240), c.h})
		case c.x+c.w-x < strip && len(v.shown(DockRight)) == 0:
			w := min(c.w/2, 260)
			return v.regionDrop(id, DockRight, dockRect{c.x + c.w - w, c.y, w, c.h})
		case c.y+c.h-y < strip && len(v.shown(DockBottom)) == 0:
			h := min(c.h/2, 180)
			return v.regionDrop(id, DockBottom, dockRect{c.x, c.y + c.h - h, c.w, h})
		}
	}
	// Hit tabs in layout order, so hidden/offscreen tabs never become targets.
	for _, side := range dockSides {
		ids, _ := v.side(side)
		for _, target := range *ids {
			if target == id {
				continue
			}
			r := v.tabRects[target]
			if r.contains(x, y) {
				after := x >= r.x+r.w/2
				at := r.x
				if after {
					at += r.w
				}
				return dockDrop{kind: 1, target: target, after: after, preview: dockRect{at - 2, r.y, 4, r.h}}
			}
		}
	}
	for _, side := range dockSides {
		for _, target := range dockPanels(*v.tree(side)) {
			if target == id || !v.Visible(target) {
				continue
			}
			n := findDockGroup(*v.tree(side), target)
			r := v.groupRects[n]
			if !r.contains(x, y) {
				continue
			}
			drop := dockDrop{kind: 1, target: target, after: true, preview: r}
			// The header is always a tab destination; body edges create splits.
			if y-r.y < 40 {
				return drop
			}
			dx, dy := x-r.x, y-r.y
			if dx < min(60, r.w*.2) {
				drop.kind = 2
				drop.placement = DockPlacementLeft
				drop.preview.w = r.w / 2
			} else if dx > r.w-min(60, r.w*.2) {
				drop.kind = 2
				drop.placement = DockPlacementRight
				drop.preview.x += r.w / 2
				drop.preview.w = r.w / 2
			} else if dy < min(80, r.h*.25) {
				drop.kind = 2
				drop.placement = DockPlacementTop
				drop.preview.h = r.h / 2
			} else if dy > r.h-min(80, r.h*.25) {
				drop.kind = 2
				drop.placement = DockPlacementBottom
				drop.preview.y += r.h / 2
				drop.preview.h = r.h / 2
			}
			if drop.kind == 2 && dockGroupDepth(*v.tree(side), target, 0) >= 32 {
				return dockDrop{}
			}
			return drop
		}
	}
	// Center edges can bring back a completely closed/empty outer region.
	r := v.centerRect
	if r.contains(x, y) {
		if x-r.x < min(60, r.w*.25) {
			r.w = min(r.w/2, 240)
			return v.regionDrop(id, DockLeft, r)
		}
		if r.x+r.w-x < min(60, r.w*.25) {
			w := min(r.w/2, 260)
			r.x += r.w - w
			r.w = w
			return v.regionDrop(id, DockRight, r)
		}
		if r.y+r.h-y < min(60, r.h*.25) {
			h := min(r.h/2, 180)
			r.y += r.h - h
			r.h = h
			return v.regionDrop(id, DockBottom, r)
		}
		// The middle of an empty center opens the tab there as a document.
		if v.documents && len(v.shown(DockCenter)) == 0 {
			return v.regionDrop(id, DockCenter, r)
		}
	}
	return dockDrop{}
}
func (v *DockView) paintDrop(cx *el.Context, gtx core.C) {
	if !v.drag.moved || v.drag.drop.kind == 0 {
		return
	}
	r := v.drag.drop.preview
	origin, _ := cx.PaintGeometry()
	scale := gtx.Metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	bounds := image.Rect(int(r.x*scale)-origin.X, int(r.y*scale)-origin.Y, int((r.x+r.w)*scale)-origin.X, int((r.y+r.h)*scale)-origin.Y)
	color := theme.Primary
	color.A = 65
	paint.FillShape(gtx.Ops, color, clip.Rect(bounds).Op())
}

func (v *DockView) tabID(id string) string { return autoID("dock", v) + "/tab/" + id }

func (v *DockView) regionDrop(id string, side DockSide, preview dockRect) dockDrop {
	ids := v.shown(side)
	if len(ids) == 0 {
		return dockDrop{kind: 3, side: side, preview: preview}
	}
	for _, target := range ids {
		if target != id {
			n := findDockGroup(*v.tree(side), target)
			return dockDrop{kind: 1, target: target, after: true, preview: v.groupRects[n]}
		}
	}
	return dockDrop{}
}
