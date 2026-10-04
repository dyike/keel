package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"time"
)

type carouselScroll struct {
	active          bool
	vertical        bool
	offset, maximum float32
	sequence        uint64
}

// Scrollable enables scroll input, on by default. Continuous deltas snap after
// 140ms without input because Gio does not expose trackpad gesture-end events.
func (v *CarouselView) Scrollable(on bool) *CarouselView {
	v.scrollable = on
	if !on {
		v.scroll.active = false
	}
	return v
}

// WheelStep selects one item per scroll event instead of accumulating pixels.
// Gio cannot distinguish a wheel from a trackpad, so this mode is explicit.
func (v *CarouselView) WheelStep(on bool) *CarouselView {
	if v.wheelStep != on {
		v.scroll.active = false
	}
	v.wheelStep = on
	return v
}

func (v *CarouselView) cancelGestures() {
	v.motion = carouselMotion{}
	v.drag.active = false
	v.scroll.active = false
}

func (v *CarouselView) scrollInput(cx *el.Context, stage *el.DivEl, id string, g *carouselGeometry) {
	if !v.scrollable || v.disabled || v.drag.active || len(g.points) == 0 {
		return
	}
	offset := g.points[v.current]
	if v.motion.ready {
		offset = v.motion.offset
	}
	if v.scroll.active {
		offset = v.scroll.offset
	}
	r := el.ScrollRange{}
	if g.cycle > 0 || v.scroll.active || offset > 0 {
		r.Min = -1e6
	}
	if g.cycle > 0 || v.scroll.active || offset < g.maximum {
		r.Max = 1e6
	}
	if v.wheelStep {
		if v.CanPrevious() {
			r.Min = -1e6
		}
		if v.CanNext() {
			r.Max = 1e6
		}
	}
	x, y := r, el.ScrollRange{}
	if v.vertical {
		x, y = el.ScrollRange{}, r
	} else if v.wheelStep {
		y = r
	}
	stage.OnScroll(x, y, func(e el.ScrollEvent) {
		delta := e.X
		if v.vertical || v.wheelStep && math.Abs(float64(e.Y)) > math.Abs(float64(e.X)) {
			delta = e.Y
		}
		if delta == 0 {
			return
		}
		if v.wheelStep {
			if delta > 0 {
				v.Next()
			} else {
				v.Previous()
			}
			return
		}
		if !v.scroll.active {
			v.scroll.offset = g.points[v.current]
			if v.motion.ready {
				v.scroll.offset = v.motion.offset
			}
			v.motion.running, v.motion.pending = false, false
		}
		v.scroll.active = true
		v.scroll.vertical = g.vertical
		v.scroll.maximum = g.maximum
		v.scroll.offset = g.constrain(v.scroll.offset + delta)
		v.scroll.sequence++
	})
	if v.scroll.active {
		sequence := v.scroll.sequence
		cx.AfterEnabled(id, struct {
			ID       string
			Sequence uint64
		}{id, sequence}, 140*time.Millisecond, func() {
			if !v.scroll.active || v.scroll.sequence != sequence {
				return
			}
			best := g.nearest(v.scroll.offset, v.current)
			v.transitionTo(best, 0)
		})
	}
}
