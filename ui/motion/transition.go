package motion

import (
	"github.com/dyike/keel/ui/internal/animation"
	"time"
)

// Kind selects the transition generator. The zero value is a spring.
type Kind = animation.Kind

const (
	KindSpring    = animation.KindSpring
	KindTween     = animation.KindTween
	KindKeyframes = animation.KindKeyframes
	KindInertia   = animation.KindInertia
)

// RepeatType controls how repeated tween and keyframe cycles restart.
type RepeatType = animation.RepeatType

const (
	RepeatLoop    = animation.RepeatLoop
	RepeatReverse = animation.RepeatReverse
	RepeatMirror  = animation.RepeatMirror
)

// Transition describes an animation segment. Keep its slices immutable in view state.
type Transition = animation.Transition

func Spring(stiffness, damping float32) Transition { return animation.Spring(stiffness, damping) }
func SpringDuration(visual time.Duration, bounce float32) Transition {
	return animation.SpringDuration(visual, bounce)
}
func Tween(duration time.Duration, ease Easing) Transition { return animation.Tween(duration, ease) }
func Keyframes(duration time.Duration, values, times []float32, ease Easing) Transition {
	return animation.Keyframes(duration, values, times, ease)
}
func Inertia(timeConstant time.Duration) Transition       { return animation.Inertia(timeConstant) }
func Stagger(index int, step time.Duration) time.Duration { return animation.Stagger(index, step) }
