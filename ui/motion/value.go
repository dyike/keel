package motion

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/animation"
	"time"
)

// Value is a number that moves toward a target over frame time. Keep it in view
// state, rather than constructing it in Render. Its zero value is ready to use.
type Value struct{ value animation.Value }

// To returns the value to display. Retargeting a spring preserves velocity.
func (v *Value) To(cx *el.Context, target float32, tr Transition) float32 {
	value, active := v.toAt(cx.Now(), target, tr, el.ReducedMotion())
	if active {
		cx.Animating()
	}
	return value
}
func (v *Value) toAt(now time.Time, target float32, tr Transition, reduced bool) (float32, bool) {
	return v.value.Sample(now, target, tr, reduced)
}

// Set jumps to x and clears velocity.
func (v *Value) Set(x float32)     { v.value.Set(x) }
func (v *Value) Get() float32      { return v.value.Get() }
func (v *Value) Velocity() float32 { return v.value.Velocity() }
func (v *Value) Animating() bool   { return v.value.Animating() }
