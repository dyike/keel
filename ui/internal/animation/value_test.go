package animation

import (
	"math"
	"testing"
	"time"
)

func TestExternallySeededVelocityAtUnchangedTarget(t *testing.T) {
	now := time.Unix(1, 0)
	var v Value
	tr := Spring(100, 20)
	v.Sample(now, 10, tr, false)
	v.SetVelocity(10, 100)
	if x, active := v.Sample(now, 10, tr, false); x != 10 || !active || v.Velocity() != 100 {
		t.Fatalf("seeded start: x=%v active=%v velocity=%v", x, active, v.Velocity())
	}
	x, active := v.Sample(now.Add(50*time.Millisecond), 10, tr, false)
	if x <= 10 || !active {
		t.Fatalf("velocity ignored: x=%v active=%v", x, active)
	}
	if x, active := v.Sample(now.Add(10*time.Second), 10, tr, false); x != 10 || active {
		t.Fatalf("did not settle: %v %v", x, active)
	}
	v.SetVelocity(10, 100)
	v.Set(10)
	if _, active := v.Sample(now.Add(11*time.Second), 10, tr, false); active {
		t.Fatal("Set retained seeded velocity")
	}
}

func TestSpringMatchesNumericalIntegration(t *testing.T) {
	for _, tr := range []Transition{Spring(100, 4), Spring(100, 20), Spring(100, 40)} {
		tr = normalize(tr)
		v := Value{cur: 0, from: 0, to: 1, initialVel: 2, tr: tr, init: true}
		v.sampleSpring(.37)
		x, velocity := rk4(tr, 0, 2, 1, .37, 1e-5)
		if math.Abs(float64(v.cur-x)) > 5e-3 || math.Abs(float64(v.vel-velocity)) > 5e-3 {
			t.Fatalf("k=%g c=%g analytic=(%g,%g), numeric=(%g,%g)", tr.Stiffness, tr.Damping, v.cur, v.vel, x, velocity)
		}
	}
}

func rk4(tr Transition, x, velocity, target, total, step float32) (float32, float32) {
	accel := func(px, pv float32) float32 { return (-tr.Stiffness*(px-target) - tr.Damping*pv) / tr.Mass }
	for elapsed := float32(0); elapsed < total; {
		dt := min(step, total-elapsed)
		k1x, k1v := velocity, accel(x, velocity)
		k2x, k2v := velocity+k1v*dt/2, accel(x+k1x*dt/2, velocity+k1v*dt/2)
		k3x, k3v := velocity+k2v*dt/2, accel(x+k2x*dt/2, velocity+k2v*dt/2)
		k4x, k4v := velocity+k3v*dt, accel(x+k3x*dt, velocity+k3v*dt)
		x += dt * (k1x + 2*k2x + 2*k3x + k4x) / 6
		velocity += dt * (k1v + 2*k2v + 2*k3v + k4v) / 6
		elapsed += dt
	}
	return x, velocity
}

func TestCubicBezierAndKeyframes(t *testing.T) {
	if got := CubicBezier(.25, .1, .25, 1)(.5); abs(got-.8024) > 5e-4 {
		t.Fatalf("ease(.5)=%g", got)
	}
	tr := normalize(Keyframes(time.Second, []float32{0, 10, 20}, []float32{0, .25, 1}, Linear))
	v := Value{from: 0, to: 20, tr: tr}
	v.sampleKeyframes(.25)
	if v.cur != 10 {
		t.Fatalf("at .25 got %g", v.cur)
	}
	for _, test := range []struct {
		typ  RepeatType
		want float32
	}{{RepeatLoop, 10}, {RepeatReverse, 16.666666}, {RepeatMirror, 16.666666}} {
		tr.Repeat, tr.RepeatType = 1, test.typ
		v.tr = tr
		v.sampleKeyframes(1.25)
		if abs(v.cur-test.want) > 1e-4 {
			t.Fatalf("repeat type %d at 1.25 got %g, want %g", test.typ, v.cur, test.want)
		}
	}
}

func TestInertiaConstraintSpringDurationAndStagger(t *testing.T) {
	minimum, maximum := float32(-100), float32(100)
	tr := Inertia(100 * time.Millisecond)
	tr.Min, tr.Max = &minimum, &maximum
	v := Value{from: 0, to: 200, tr: normalize(tr)}
	v.sampleInertia(2)
	if abs(v.cur-100) > .01 || !v.done {
		t.Fatalf("constrained inertia value=%g done=%v", v.cur, v.done)
	}
	spring := normalize(SpringDuration(500*time.Millisecond, .25))
	if spring.Stiffness <= 0 || spring.Damping <= 0 || spring.Mass != 1 {
		t.Fatalf("invalid duration spring: %+v", spring)
	}
	if got := Stagger(3, 40*time.Millisecond); got != 120*time.Millisecond {
		t.Fatalf("stagger=%v", got)
	}
}
