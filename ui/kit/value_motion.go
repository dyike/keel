package kit

import (
	"github.com/dyike/keel/ui/el"
	"time"
)

// valueMotion retargets from the currently displayed value, using the frame
// clock. Initial values and reduced-motion changes are displayed immediately.
type valueMotion struct {
	initialized     bool
	value, from, to float32
	started         time.Time
}

func (m *valueMotion) sample(cx *el.Context, target float32, duration time.Duration) float32 {
	now := cx.Now()
	if !m.initialized || el.ReducedMotion() {
		m.initialized = true
		m.value = target
		m.from = target
		m.to = target
		m.started = now
		return target
	}
	p := max(0, min(1, float32(now.Sub(m.started))/float32(duration)))
	p = p * p * (3 - 2*p)
	m.value = m.from + (m.to-m.from)*p
	if p == 1 {
		m.value = m.to
	}
	if target != m.to {
		m.from = m.value
		m.to = target
		m.started = now
	}
	if m.value != m.to {
		cx.Animating()
	}
	return m.value
}

// ProgressDuration is the duration of a determinate progress value transition.
const ProgressDuration = 200 * time.Millisecond
