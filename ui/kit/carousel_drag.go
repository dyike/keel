package kit

import (
	"math"

	"github.com/dyike/keel/ui/el"
)

type carouselGeometry struct {
	points   []float32
	maximum  float32
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
		v.scroll.active = false
		v.drag = carouselDrag{active: true, vertical: g.vertical, start: position,
			origin: g.points[v.current], offset: g.points[v.current], maximum: g.maximum}
	case el.DragMove, el.DragEnd:
		if !v.drag.active {
			return
		}
		if e.Canceled {
			v.drag.active = false
			return
		}
		delta := v.drag.start - position
		if math.Abs(float64(delta)) >= 4 {
			v.drag.moved = true
		}
		if v.drag.moved {
			v.drag.offset = min(max(v.drag.origin+delta, 0), g.maximum)
		}
		if e.Kind != el.DragEnd {
			return
		}
		moved, offset := v.drag.moved, v.drag.offset
		v.drag.active = false
		if !moved {
			return
		}
		best := v.current
		distance := float32(math.Abs(float64(g.points[best] - offset)))
		for i, p := range g.points {
			d := float32(math.Abs(float64(p - offset)))
			if d < distance {
				best, distance = i, d
			}
		}
		v.goTo(best)
	}
}
