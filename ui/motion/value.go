package motion

import (
	"math"
	"time"

	"github.com/dyike/keel/ui/el"
)

// Value is a number that moves toward a target over frame time. Keep it in a
// view rather than constructing it in Render. Its zero value is ready to use.
type Value struct {
	cur, vel, from, to, initialVel float32
	start                          time.Time
	tr                             Transition
	sig                            transitionSignature
	init, done                     bool
}

// To declares the target for this frame and returns the value to paint.
// Retargeting a spring preserves its instantaneous velocity.
func (v *Value) To(cx *el.Context, target float32, tr Transition) float32 {
	value, animate := v.toAt(cx.Now(), target, tr, el.ReducedMotion())
	if animate {
		cx.Animating()
	}
	return value
}

// toAt is the allocation-free sampling kernel behind To. Its bool result says
// whether the caller should request another frame.
func (v *Value) toAt(now time.Time, target float32, tr Transition, reduced bool) (float32, bool) {
	tr = normalize(tr)
	sig := signature(tr)
	if !v.init {
		v.Set(target)
		v.tr, v.sig = tr, sig
		return target, false
	}
	if reduced {
		v.Set(target)
		v.tr, v.sig = tr, sig
		return target, false
	}
	if target != v.to || sig != v.sig {
		if !v.done {
			v.advance(now)
		}
		v.from, v.to, v.initialVel = v.cur, target, v.vel
		v.start, v.tr, v.sig, v.done = now, tr, sig, false
	}
	if !v.done {
		v.advance(now)
	}
	return v.cur, !v.done
}

// Set jumps to x and clears velocity.
func (v *Value) Set(x float32) {
	v.cur, v.vel, v.from, v.to, v.initialVel = x, 0, x, x, 0
	v.init, v.done = true, true
}

func (v *Value) Get() float32      { return v.cur }
func (v *Value) Velocity() float32 { return v.vel }
func (v *Value) Animating() bool   { return v.init && !v.done }

func (v *Value) advance(now time.Time) {
	elapsed := now.Sub(v.start) - v.tr.Delay
	if elapsed < 0 {
		v.cur, v.vel = v.from, 0
		return
	}
	seconds := float32(elapsed.Seconds())
	switch v.tr.Kind {
	case KindTween:
		v.sampleTween(seconds)
	case KindKeyframes:
		v.sampleKeyframes(seconds)
	case KindInertia:
		v.sampleInertia(seconds)
	default:
		v.sampleSpring(seconds)
	}
	if math.IsNaN(float64(v.cur)) || math.IsInf(float64(v.cur), 0) || math.IsNaN(float64(v.vel)) || math.IsInf(float64(v.vel), 0) {
		v.cur, v.vel, v.done = v.to, 0, true
	}
}

func (v *Value) sampleSpring(t float32) {
	k, c, m := v.tr.Stiffness, v.tr.Damping, v.tr.Mass
	x0, v0 := v.from-v.to, v.initialVel
	w0 := float32(math.Sqrt(float64(k / m)))
	z := c / (2 * float32(math.Sqrt(float64(k*m))))
	var x, velocity float32
	switch {
	case z < 1-1e-4:
		a := c / (2 * m)
		wd := w0 * float32(math.Sqrt(float64(1-z*z)))
		b := (v0 + a*x0) / wd
		e := float32(math.Exp(float64(-a * t)))
		co, si := float32(math.Cos(float64(wd*t))), float32(math.Sin(float64(wd*t)))
		x = e * (x0*co + b*si)
		velocity = e * ((-a*x0+b*wd)*co + (-a*b-x0*wd)*si)
	case z > 1+1e-4:
		disc := float32(math.Sqrt(float64(max(0, c*c-4*m*k))))
		r1, r2 := (-c+disc)/(2*m), (-c-disc)/(2*m)
		a := (v0 - r2*x0) / (r1 - r2)
		b := x0 - a
		e1, e2 := float32(math.Exp(float64(r1*t))), float32(math.Exp(float64(r2*t)))
		x, velocity = a*e1+b*e2, a*r1*e1+b*r2*e2
	default:
		b := v0 + w0*x0
		e := float32(math.Exp(float64(-w0 * t)))
		x = (x0 + b*t) * e
		velocity = (b - w0*(x0+b*t)) * e
	}
	v.cur, v.vel = v.to+x, velocity
	if abs(v.vel) < v.tr.RestSpeed && abs(v.to-v.cur) < v.tr.RestDelta {
		v.cur, v.vel, v.done = v.to, 0, true
	}
}

func (v *Value) sampleTween(t float32) {
	p, direction, done := repeatPhase(t, float32(v.tr.Duration.Seconds()), float32(v.tr.RepeatDelay.Seconds()), v.tr.Repeat, v.tr.RepeatType)
	eased := v.tr.Ease(p)
	v.cur = v.from + (v.to-v.from)*eased
	if done {
		v.vel, v.done = 0, true
		return
	}
	d := max(float32(v.tr.Duration.Seconds()), 1e-6)
	h := float32(1e-3)
	derivative := (v.tr.Ease(clamp01(p+h)) - v.tr.Ease(clamp01(p-h))) / (2 * h)
	v.vel = (v.to - v.from) * derivative / d * direction
}

func (v *Value) sampleKeyframes(t float32) {
	p, direction, done := repeatPhase(t, float32(v.tr.Duration.Seconds()), float32(v.tr.RepeatDelay.Seconds()), v.tr.Repeat, v.tr.RepeatType)
	v.cur = keyframeAt(v.from, v.to, v.tr.Values, v.tr.Times, v.tr.Ease, p)
	if done {
		v.vel, v.done = 0, true
		return
	}
	d := max(float32(v.tr.Duration.Seconds()), 1e-6)
	h := float32(1e-3)
	a := keyframeAt(v.from, v.to, v.tr.Values, v.tr.Times, v.tr.Ease, clamp01(p-h))
	b := keyframeAt(v.from, v.to, v.tr.Values, v.tr.Times, v.tr.Ease, clamp01(p+h))
	v.vel = (b - a) / (2 * h * d) * direction
}

func (v *Value) sampleInertia(t float32) {
	target := v.to
	if v.tr.Min != nil {
		target = max(target, *v.tr.Min)
	}
	if v.tr.Max != nil {
		target = min(target, *v.tr.Max)
	}
	tau := max(v.tr.TimeConstant, 1e-4)
	e := float32(math.Exp(float64(-t / tau)))
	v.cur = target + (v.from-target)*e
	v.vel = -(v.from - target) * e / tau
	if abs(v.vel) < v.tr.RestSpeed && abs(target-v.cur) < v.tr.RestDelta {
		v.cur, v.vel, v.done = target, 0, true
	}
}

func repeatPhase(t, duration, repeatDelay float32, repeat int, typ RepeatType) (phase, direction float32, done bool) {
	if duration <= 0 {
		return 1, 1, true
	}
	cycleDuration := duration + max(0, repeatDelay)
	cycle := int(t / cycleDuration)
	within := t - float32(cycle)*cycleDuration
	if repeat >= 0 && cycle > repeat {
		cycle, within, done = repeat, duration, true
	}
	p := min(within/duration, 1)
	direction = 1
	if (typ == RepeatReverse || typ == RepeatMirror) && cycle%2 == 1 {
		p, direction = 1-p, -1
	}
	return p, direction, done
}

func keyframeAt(from, to float32, values, times []float32, ease Easing, p float32) float32 {
	count := len(values)
	if count == 0 {
		return from + (to-from)*ease(p)
	}
	if count == 1 {
		return values[0]
	}
	segment := min(int(p*float32(count-1)), count-2)
	start, end := float32(segment)/float32(count-1), float32(segment+1)/float32(count-1)
	if len(times) == count && validTimes(times) {
		segment = count - 2
		for i := 0; i < count-1; i++ {
			if p <= times[i+1] {
				segment = i
				break
			}
		}
		start, end = times[segment], times[segment+1]
	}
	local := float32(1)
	if end > start {
		local = clamp01((p - start) / (end - start))
	}
	return values[segment] + (values[segment+1]-values[segment])*ease(local)
}

func validTimes(times []float32) bool {
	if len(times) < 2 || times[0] != 0 || times[len(times)-1] != 1 {
		return false
	}
	for i := 1; i < len(times); i++ {
		if times[i] < times[i-1] {
			return false
		}
	}
	return true
}
