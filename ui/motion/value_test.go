package motion

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type valueView struct {
	v      Value
	target float32
	tr     Transition
	got    float32
}

func (v *valueView) Render(cx *el.Context) el.Element {
	v.got = v.v.To(cx, v.target, v.tr)
	return el.Div().Size(el.Dp(10))
}

type fakeFrames struct {
	root   *el.RootWidget
	ops    op.Ops
	router input.Router
}

func newFakeFrames(view el.View) *fakeFrames { return &fakeFrames{root: el.Embed(view)} }

func (f *fakeFrames) at(now time.Time) bool {
	f.router.WakeupTime()
	f.ops.Reset()
	gtx := layout.Context{Ops: &f.ops, Now: now, Source: f.router.Source(),
		Constraints: layout.Exact(image.Pt(100, 100)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	f.root.Layout(gtx)
	f.router.Frame(&f.ops)
	_, wakeup := f.router.PeekWakeup()
	return wakeup
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

func TestFrameRateIndependent(t *testing.T) {
	start := time.Unix(10, 0)
	makeView := func() (*valueView, *fakeFrames) {
		v := &valueView{target: 1, tr: Spring(180, 18)}
		v.v.Set(0)
		return v, newFakeFrames(v)
	}
	a, af := makeView()
	b, bf := makeView()
	af.at(start)
	bf.at(start)
	for i := 1; i <= 60; i++ {
		af.at(start.Add(time.Duration(i) * time.Second / 60))
	}
	for i := 1; i <= 120; i++ {
		bf.at(start.Add(time.Duration(i) * time.Second / 120))
	}
	if d := abs(a.got - b.got); d > 1e-5 {
		t.Fatalf("60Hz=%g 120Hz=%g delta=%g", a.got, b.got, d)
	}
}

func TestSpringRetargetPreservesVelocity(t *testing.T) {
	start := time.Unix(20, 0)
	v := &valueView{target: 1, tr: Spring(120, 12)}
	v.v.Set(0)
	f := newFakeFrames(v)
	f.at(start)
	f.at(start.Add(150 * time.Millisecond))
	before := v.v.Velocity()
	v.target = -1
	f.at(start.Add(150 * time.Millisecond))
	if d := abs(v.v.Velocity() - before); d > 1e-4 {
		t.Fatalf("velocity changed at retarget: before=%g after=%g", before, v.v.Velocity())
	}
}

func TestSettlesAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	start := time.Unix(30, 0)
	v := &valueView{target: 1, tr: Spring(150, 20)}
	v.v.Set(0)
	f := newFakeFrames(v)
	f.at(start)
	wakeup := f.at(start.Add(5 * time.Second))
	if v.v.Animating() || v.got != 1 {
		t.Fatalf("spring did not settle: value=%g velocity=%g", v.got, v.v.Velocity())
	}
	if wakeup {
		t.Fatal("settled spring requested another frame")
	}
	v.target = 2
	theme.SetReducedMotion(true)
	wakeup = f.at(start.Add(6 * time.Second))
	if v.got != 2 || v.v.Animating() {
		t.Fatalf("reduced motion value=%g animating=%v", v.got, v.v.Animating())
	}
	if wakeup {
		t.Fatal("reduced motion requested another frame")
	}
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

func TestTweenRetargetDoesNotJump(t *testing.T) {
	tr := Tween(time.Second, EaseInOut)
	var v Value
	v.Set(0)
	start := time.Unix(35, 0)
	v.toAt(start, 1, tr, false)
	before, _ := v.toAt(start.Add(300*time.Millisecond), 1, tr, false)
	after, _ := v.toAt(start.Add(300*time.Millisecond), 2, tr, false)
	if abs(after-before) > 1e-6 {
		t.Fatalf("tween jumped on retarget: before=%g after=%g", before, after)
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

func TestMapRangeDragAndColor(t *testing.T) {
	if got := MapRange(2, 0, 1, 10, 20, true); got != 20 {
		t.Fatalf("clamped map=%g", got)
	}
	var d DragTracker
	start := time.Unix(0, 0)
	d.Track(start, el.DragEvent{Kind: el.DragStart})
	d.Track(start.Add(100*time.Millisecond), el.DragEvent{Kind: el.DragMove, X: 20, Y: -10})
	x, y := d.Velocity()
	if abs(x-200) > .001 || abs(y+100) > .001 {
		t.Fatalf("velocity=(%g,%g)", x, y)
	}
	got := MixColor(colorNRGBA(0, 0, 0, 0), colorNRGBA(100, 0, 0, 200), .25)
	if got.R != 50 || got.A != 50 {
		t.Fatalf("mixed color=%v", got)
	}
}

type presenceView struct {
	p     Presence[int]
	keys  []int
	items []PresenceItem[int]
}

func (v *presenceView) Render(cx *el.Context) el.Element {
	v.items = append(v.items[:0], v.p.Items(cx, v.keys,
		Tween(100*time.Millisecond, Linear), Tween(100*time.Millisecond, Linear))...)
	return el.Div()
}

func TestPresenceStableExit(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	start := time.Unix(40, 0)
	v := &presenceView{keys: []int{1, 2}}
	f := newFakeFrames(v)
	f.at(start)
	f.at(start.Add(100 * time.Millisecond))
	v.keys = []int{2, 3}
	f.at(start.Add(110 * time.Millisecond))
	if len(v.items) != 3 || v.items[0].Key != 1 || !v.items[0].Exiting || !v.items[0].Disabled || v.items[1].Key != 2 || v.items[2].Key != 3 {
		t.Fatalf("unexpected transition items: %+v", v.items)
	}
	f.at(start.Add(220 * time.Millisecond))
	if len(v.items) != 2 || v.items[0].Key != 2 || v.items[1].Key != 3 {
		t.Fatalf("exit was not removed in stable order: %+v", v.items)
	}
}

type layoutView struct {
	group   LayoutGroup
	reverse bool
}

func (v *layoutView) Render(cx *el.Context) el.Element {
	order := [2]int{0, 1}
	if v.reverse {
		order = [2]int{1, 0}
	}
	row := el.Div().Row()
	for _, id := range order {
		tile := el.Div().Size(el.Dp(20)).NoShrink()
		row.Child(v.group.Animate(cx, fmt.Sprint(id), tile, Spring(200, 20)))
	}
	return row
}

func TestLayoutGroupDetectsPositionChange(t *testing.T) {
	start := time.Unix(45, 0)
	v := new(layoutView)
	f := newFakeFrames(v)
	f.at(start)
	v.reverse = true
	if !f.at(start.Add(time.Millisecond)) {
		t.Fatal("layout change did not request a compensation frame")
	}
	if v.group.states["0"].x.Get() == 0 || v.group.states["1"].x.Get() == 0 {
		t.Fatalf("missing FLIP offsets: zero=%g one=%g", v.group.states["0"].x.Get(), v.group.states["1"].x.Get())
	}
}

func colorNRGBA(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }

func TestSpringSampleAllocs(t *testing.T) {
	tr := Spring(100, 10)
	var v Value
	v.Set(0)
	now := time.Unix(50, 0)
	if got := testing.AllocsPerRun(1000, func() {
		now = now.Add(time.Millisecond)
		v.toAt(now, 1, tr, false)
	}); got != 0 {
		t.Fatalf("Value.To kernel allocs=%g", got)
	}
}

func BenchmarkValueTo(b *testing.B) {
	tr := Spring(100, 10)
	var v Value
	v.Set(0)
	now := time.Unix(60, 0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		now = now.Add(time.Millisecond)
		v.toAt(now, 1, tr, false)
	}
}

func BenchmarkPresenceItems(b *testing.B) {
	keys := []int{1, 2, 3, 4}
	tr := Tween(100*time.Millisecond, Linear)
	var p Presence[int]
	now := time.Unix(70, 0)
	p.itemsAt(now, keys, tr, tr, false)
	p.itemsAt(now.Add(time.Second), keys, tr, tr, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.itemsAt(now.Add(time.Second), keys, tr, tr, false)
	}
}
