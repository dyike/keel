package kit

import (
	"math"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

type carouselScroll struct {
	active          bool
	vertical        bool
	offset, maximum float32
	sequence        uint64
	ended           uint64 // core.ScrollGesture.Ended when this scroll began
}

// Scrollable enables scroll input, on by default. Trackpad scrolling snaps
// when the fingers lift and ignores the momentum that follows. A mouse wheel
// steps one item per notch along the carousel's axis. Where the platform
// reports neither (only macOS does), continuous deltas snap after 140ms
// without input.
func (v *CarouselView) Scrollable(on bool) *CarouselView {
	v.scrollable = on
	if !on {
		v.scroll.active = false
	}
	return v
}

// WheelStep selects one item per scroll event instead of accumulating pixels,
// for any device, and lets a horizontal carousel step on vertical scrolling
// too. Without it, stepping applies only to a detected mouse wheel.
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
		gesture := core.CurrentScrollGesture()
		if delta == 0 || gesture.Momentum {
			return // inertia after a lift; the lift already chose the item
		}
		if v.wheelStep || gesture.Device == core.ScrollDeviceWheel {
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
			v.scroll.ended = gesture.Ended
		}
		v.scroll.active = true
		v.scroll.vertical = g.vertical
		v.scroll.maximum = g.maximum
		v.scroll.offset = g.constrain(v.scroll.offset + delta)
		v.scroll.sequence++
	})
	if v.scroll.active {
		delay := 140 * time.Millisecond
		if gesture := core.CurrentScrollGesture(); gesture.Device == core.ScrollDeviceTrackpad && gesture.Phases {
			if gesture.Ended != v.scroll.ended {
				delay = 0 // the fingers lifted
			} else if gesture.Active {
				delay = time.Second // fingers still down: wait for the lift, unless it never comes
				// Look again soon: some platforms report the lift only when asked.
				cx.AfterEnabled(id, struct {
					ID       string
					Sequence uint64
					Poll     bool
				}{id, v.scroll.sequence, true}, 50*time.Millisecond, func() {})
			}
		}
		sequence := v.scroll.sequence
		cx.AfterEnabled(id, struct {
			ID       string
			Sequence uint64
			Lifted   bool
		}{id, sequence, delay == 0}, delay, func() {
			if !v.scroll.active || v.scroll.sequence != sequence {
				return
			}
			best := g.nearest(v.scroll.offset, v.current)
			v.transitionTo(best, 0)
		})
	}
}
