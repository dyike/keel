package kit

import (
	"math"
	"slices"
	"time"

	"github.com/dyike/keel/ui/el"
)

const carouselTransition = 180 * time.Millisecond

type carouselMotion struct {
	geometry                *carouselGeometry
	ready, pending, running bool
	direction               int
	offset, from, to        float32
	started                 time.Time
}

func (a *carouselGeometry) same(b *carouselGeometry) bool {
	return b != nil && a.viewport == b.viewport && a.maximum == b.maximum && a.cycle == b.cycle && a.vertical == b.vertical && slices.Equal(a.points, b.points)
}

func (v *CarouselView) animateTo(direction int) {
	m := &v.motion
	if v.drag.active {
		m.offset = v.drag.offset
	}
	if v.scroll.active {
		m.offset = v.scroll.offset
	}
	m.pending = m.ready
	m.direction = direction
	v.drag.active = false
	v.scroll.active = false
}

// motionOffset samples only while painting the enabled stage, after its final
// geometry is known. Programmatic changes and resize start at the new target.
func (v *CarouselView) motionOffset(cx *el.Context, g *carouselGeometry, target float32, enabled bool) float32 {
	m := &v.motion
	if !m.ready || !g.same(m.geometry) || !enabled || el.ReducedMotion() {
		*m = carouselMotion{ready: true, geometry: g, offset: target}
		return target
	}
	if v.drag.active || v.scroll.active {
		m.offset = target
		m.running, m.pending = false, false
		return target
	}
	if m.pending {
		to := target
		if g.cycle > 0 {
			to += float32(math.Round(float64((m.offset-to)/g.cycle))) * g.cycle
			if m.direction > 0 && to < m.offset {
				to += g.cycle
			}
			if m.direction < 0 && to > m.offset {
				to -= g.cycle
			}
		}
		m.from, m.to, m.started = m.offset, to, cx.Now()
		m.running = m.from != m.to
		m.pending = false
	}
	if m.running {
		t := float32(cx.Now().Sub(m.started)) / float32(carouselTransition)
		if t >= 1 {
			m.running = false
			m.offset = target
		} else {
			t = max(0, t)
			eased := 1 - (1-t)*(1-t)*(1-t)
			m.offset = m.from + (m.to-m.from)*eased
			cx.Animating()
		}
	} else {
		m.offset = target
	}
	return m.offset
}
