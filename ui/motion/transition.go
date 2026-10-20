package motion

import (
	"math"
	"time"
)

// Kind selects the transition generator. The zero value is a spring.
type Kind uint8

const (
	KindSpring Kind = iota
	KindTween
	KindKeyframes
	KindInertia
)

// RepeatType controls how repeated tween and keyframe cycles restart.
type RepeatType uint8

const (
	RepeatLoop RepeatType = iota
	RepeatReverse
	RepeatMirror
)

// Transition describes one animation segment. Keep Values and Times in view
// state; changing their backing array starts a new segment.
type Transition struct {
	Kind Kind

	Delay                    time.Duration
	Stiffness, Damping, Mass float32
	RestSpeed, RestDelta     float32
	Duration                 time.Duration
	Ease                     Easing
	Values, Times            []float32
	Repeat                   int // -1 repeats until no longer rendered
	RepeatType               RepeatType
	RepeatDelay              time.Duration
	Power, TimeConstant      float32
	Min, Max                 *float32
}

// Spring creates a physics spring. Mass defaults to one.
func Spring(stiffness, damping float32) Transition {
	return Transition{Kind: KindSpring, Stiffness: stiffness, Damping: damping, Mass: 1}
}

// SpringDuration approximates a duration/bounce spring. Bounce is clamped to
// 0..1; higher values lower the damping ratio.
func SpringDuration(visual time.Duration, bounce float32) Transition {
	seconds := max(visual.Seconds(), .001)
	bounce = clamp01(bounce)
	omega := float32(2 * math.Pi / seconds)
	zeta := max(.05, 1-bounce)
	return Transition{Kind: KindSpring, Stiffness: omega * omega, Damping: 2 * zeta * omega, Mass: 1}
}

// Tween creates a duration-based interpolation.
func Tween(duration time.Duration, ease Easing) Transition {
	return Transition{Kind: KindTween, Duration: duration, Ease: ease}
}

// Keyframes creates a transition through values. Times may be nil for evenly
// spaced frames. The slices are borrowed and must remain immutable.
func Keyframes(duration time.Duration, values, times []float32, ease Easing) Transition {
	return Transition{Kind: KindKeyframes, Duration: duration, Values: values, Times: times, Ease: ease}
}

// Inertia creates an exponential decay toward the target supplied to Value.To.
func Inertia(timeConstant time.Duration) Transition {
	return Transition{Kind: KindInertia, TimeConstant: float32(timeConstant.Seconds())}
}

// Stagger returns the delay for a zero-based item index.
func Stagger(index int, step time.Duration) time.Duration {
	if index <= 0 || step <= 0 {
		return 0
	}
	return time.Duration(index) * step
}

func normalize(t Transition) Transition {
	if t.Delay < 0 {
		t.Delay = 0
	}
	if t.Repeat < -1 {
		t.Repeat = 0
	}
	if t.RepeatDelay < 0 {
		t.RepeatDelay = 0
	}
	switch t.Kind {
	case KindTween, KindKeyframes:
		if t.Duration <= 0 {
			t.Duration = 300 * time.Millisecond
		}
		if t.Ease == nil {
			t.Ease = EaseInOut
		}
	case KindInertia:
		if t.TimeConstant <= 0 {
			t.TimeConstant = .7
		}
		if t.RestSpeed <= 0 {
			t.RestSpeed = .01
		}
		if t.RestDelta <= 0 {
			t.RestDelta = .005
		}
	default:
		t.Kind = KindSpring
		if t.Stiffness <= 0 {
			t.Stiffness = 100
		}
		if t.Damping < 0 {
			t.Damping = 10
		}
		if t.Mass <= 0 {
			t.Mass = 1
		}
		if t.RestSpeed <= 0 {
			t.RestSpeed = .01
		}
		if t.RestDelta <= 0 {
			t.RestDelta = .005
		}
	}
	return t
}

type transitionSignature struct {
	kind                                               Kind
	delay, duration, repeatDelay                       time.Duration
	a, b, c, d, e, f, g, min, max, ease1, ease2, ease3 float32
	repeat                                             int
	repeatType                                         RepeatType
	valuesLen, timesLen                                int
	valuesPtr, timesPtr                                *float32
}

func signature(t Transition) transitionSignature {
	var vp, tp *float32
	if len(t.Values) > 0 {
		vp = &t.Values[0]
	}
	if len(t.Times) > 0 {
		tp = &t.Times[0]
	}
	e1, e2, e3 := float32(0), float32(0), float32(0)
	if t.Ease != nil {
		e1, e2, e3 = t.Ease(.25), t.Ease(.5), t.Ease(.75)
	}
	minv, maxv := float32(math.Inf(-1)), float32(math.Inf(1))
	if t.Min != nil {
		minv = *t.Min
	}
	if t.Max != nil {
		maxv = *t.Max
	}
	return transitionSignature{
		kind: t.Kind, delay: t.Delay, duration: t.Duration, repeatDelay: t.RepeatDelay,
		a: t.Stiffness, b: t.Damping, c: t.Mass, d: t.RestSpeed, e: t.RestDelta,
		f: t.Power, g: t.TimeConstant, min: minv, max: maxv,
		ease1: e1, ease2: e2, ease3: e3, repeat: t.Repeat, repeatType: t.RepeatType,
		valuesLen: len(t.Values), timesLen: len(t.Times), valuesPtr: vp, timesPtr: tp,
	}
}
