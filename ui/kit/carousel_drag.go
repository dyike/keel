package kit

import (
	"math"

	"github.com/dyike/keel/ui/el"
)

type carouselGeometry struct {
	points   []float32
	maximum  float32
	cycle    float32
	viewport float32
	vertical bool
}

type carouselDrag struct {
	active, moved, vertical        bool
	start, origin, offset, maximum float32
}

// Draggable enables pointer dragging, on by default. A release selects the
// nearest snap point. Disabling it cancels a pending drag without a callback.
func (v *CarouselView) Draggable(on bool) *CarouselView {
	v.draggable = on
	if !on {
		v.drag.active = false
	}
	return v
}

func (v *CarouselView) handleDrag(e el.DragEvent, g *carouselGeometry) {
	if v.disabled || !v.draggable || len(g.points) == 0 {
		return
	}
	position := e.X
	if g.vertical {
		position = e.Y
	}
	switch e.Kind {
	case el.DragStart:
		origin := g.points[v.current]
		if v.motion.ready {
			origin = v.motion.offset
		}
		v.motion.running, v.motion.pending = false, false
		if v.scroll.active {
			origin = v.scroll.offset
		}
		v.scroll.active = false
		v.drag = carouselDrag{active: true, vertical: g.vertical, start: position,
			origin: origin, offset: origin, maximum: g.maximum}
	case el.DragMove, el.DragEnd:
		if !v.drag.active {
			return
		}
		if e.Canceled {
			v.animateTo(0)
			return
		}
		delta := v.drag.start - position
		if math.Abs(float64(delta)) >= 4 {
			v.drag.moved = true
		}
		if v.drag.moved {
			v.drag.offset = g.constrain(v.drag.origin + delta)
		}
		if e.Kind != el.DragEnd {
			return
		}
		moved, offset := v.drag.moved, v.drag.offset
		if !moved {
			v.animateTo(0)
			return
		}
		best := g.nearest(offset, v.current)
		v.transitionTo(best, 0)
	}
}

// Circular coordinates keep gestures local while selecting the nearest copy of
// a snap point. Items themselves are never duplicated.
func (g *carouselGeometry) constrain(offset float32) float32 {
	if g.cycle == 0 {
		return min(max(offset, 0), g.maximum)
	}
	return offset
}
func (g *carouselGeometry) normalized(offset float32) float32 {
	if g.cycle == 0 {
		return offset
	}
	offset = float32(math.Mod(float64(offset), float64(g.cycle)))
	if offset < 0 {
		offset += g.cycle
	}
	return offset
}
func (g *carouselGeometry) nearest(offset float32, current int) int {
	offset = g.normalized(offset)
	distance := func(p float32) float32 {
		d := float32(math.Abs(float64(p - offset)))
		if g.cycle > 0 {
			d = min(d, g.cycle-d)
		}
		return d
	}
	best, d := current, distance(g.points[current])
	for i, p := range g.points {
		if next := distance(p); next < d {
			best, d = i, next
		}
	}
	return best
}
