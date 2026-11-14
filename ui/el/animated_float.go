package el

import "github.com/dyike/keel/ui/internal/animation"

// AnimatedFloat retains a scalar spring or tween across renders. Use it when
// the animated value affects layout, rather than only a paint transform.
// Keep it in view state and call Sample during Render under the frame lock.
type AnimatedFloat struct {
	value animation.Value
}

// Set initializes or immediately replaces the value without animation.
func (v *AnimatedFloat) Set(value float32) { v.value.Set(value) }

// Get returns the last sampled value.
func (v *AnimatedFloat) Get() float32 { return v.value.Get() }

// Sample moves toward target using the context's frame clock, preserving
// velocity when retargeted. It schedules frames only while moving and applies
// the target immediately when reduced motion is enabled.
func (v *AnimatedFloat) Sample(cx *Context, target float32, transition Animation) float32 {
	value, moving := v.value.Sample(cx.Now(), target, transition, ReducedMotion())
	if moving {
		cx.Animating()
	}
	return value
}
