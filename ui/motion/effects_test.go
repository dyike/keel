package motion

import (
	"image"
	"reflect"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type effectHarness struct {
	root   *el.RootWidget
	router input.Router
	ops    op.Ops
	scale  float32
}

func newEffectHarness(view el.View, scale float32) *effectHarness {
	return &effectHarness{root: el.Embed(view), scale: scale}
}

func (h *effectHarness) frame(now time.Time) bool {
	h.router.WakeupTime()
	h.ops.Reset()
	size := image.Pt(int(320*h.scale), int(240*h.scale))
	gtx := layout.Context{Ops: &h.ops, Now: now, Source: h.router.Source(), Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: h.scale, PxPerSp: h.scale}}
	h.root.Layout(gtx)
	h.router.Frame(&h.ops)
	_, wakeup := h.router.PeekWakeup()
	return wakeup
}

func (h *effectHarness) move(x, y float32) {
	h.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x*h.scale, y*h.scale)})
}

func (h *effectHarness) press(x, y float32) {
	position := f32.Pt(x*h.scale, y*h.scale)
	h.router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: position},
	)
}

type glowScrollView struct{ glow PointerGlow }

func (v *glowScrollView) Render(cx *el.Context) el.Element {
	card := el.Div().W(el.Dp(100)).H(el.Dp(50))
	v.glow.Apply(cx, "glow-card", card, PointerGlowStyle{})
	return el.Div().W(el.Dp(200)).H(el.Dp(80)).ScrollY().ScrollOffset(0, 50).Child(
		el.Div().H(el.Dp(80)).NoShrink(),
		card.NoShrink(),
	)
}

func TestPointerGlowOriginUnderHiDPIAndScroll(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	start := time.Unix(80, 0)
	v := new(glowScrollView)
	h := newEffectHarness(v, 2)
	h.frame(start)
	// Card natural y=80dp, controlled scroll=50dp, so visual y=30dp.
	// A global pointer at (20,40) is card-local (20,10), independent of 2x.
	h.move(20, 40)
	if !h.frame(start.Add(time.Millisecond)) {
		t.Fatal("hover enter did not request an effect frame")
	}
	x, y := v.glow.Origin()
	if abs(x-20) > .01 || abs(y-10) > .01 {
		t.Fatalf("origin=(%g,%g), want (20,10)", x, y)
	}
	if abs(v.glow.rippleX-20) > .01 || abs(v.glow.rippleY-10) > .01 {
		t.Fatalf("ripple origin=(%g,%g)", v.glow.rippleX, v.glow.rippleY)
	}
	h.press(70, 50)
	h.frame(start.Add(2 * time.Millisecond))
	if abs(v.glow.rippleX-70) > .01 || abs(v.glow.rippleY-20) > .01 {
		t.Fatalf("pressed ripple origin=(%g,%g), want (70,20)", v.glow.rippleX, v.glow.rippleY)
	}

	h.move(300, 220)
	h.frame(start.Add(20 * time.Millisecond))
	if wakeup := h.frame(start.Add(2 * time.Second)); wakeup || v.glow.Active() {
		t.Fatalf("finished pointer effect wakeup=%v active=%v", wakeup, v.glow.Active())
	}
}

func TestPointerGlowReducedMotionSkipsEffect(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(true)
	start := time.Unix(81, 0)
	v := new(glowScrollView)
	h := newEffectHarness(v, 1)
	h.frame(start)
	h.move(20, 40)
	h.frame(start.Add(time.Millisecond)) // el hover state itself invalidates once.
	if v.glow.Active() {
		t.Fatal("reduced pointer effect became active")
	}
	if wakeup := h.frame(start.Add(2 * time.Millisecond)); wakeup {
		t.Fatal("reduced pointer effect requested a follow-up frame")
	}
}

type confettiView struct{ confetti Confetti }

func (v *confettiView) Render(cx *el.Context) el.Element {
	host := el.Div().W(el.Dp(300)).H(el.Dp(200))
	return v.confetti.Apply(cx, host)
}

func testConfettiOptions(count int, seed uint64) ConfettiOptions {
	return ConfettiOptions{
		Count: count, Seed: seed, Gravity: 500, SpeedMin: 120, SpeedMax: 240,
		LifetimeMin: 100 * time.Millisecond, LifetimeMax: 100 * time.Millisecond,
		SizeMin: 3, SizeMax: 6,
	}
}

func TestConfettiLifetimeStopsFrames(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	start := time.Unix(82, 0)
	v := new(confettiView)
	v.confetti.Emit(start, 150, 80, testConfettiOptions(256, 42))
	h := newEffectHarness(v, 1)
	if !h.frame(start) {
		t.Fatal("live confetti did not request a frame")
	}
	if wakeup := h.frame(start.Add(101 * time.Millisecond)); wakeup || v.confetti.Active() {
		t.Fatalf("expired confetti wakeup=%v active=%v particles=%d", wakeup, v.confetti.Active(), len(v.confetti.particles))
	}
}

func TestConfettiReducedMotionAndDeterministicSeed(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	start := time.Unix(83, 0)
	theme.SetReducedMotion(false)
	options := testConfettiOptions(37, 99)
	var a, b Confetti
	a.Emit(start, 10, 20, options)
	b.Emit(start, 10, 20, options)
	if !reflect.DeepEqual(a.particles, b.particles) {
		t.Fatal("equal seeds produced different particles")
	}
	if len(a.particles) != 37 {
		t.Fatalf("particle count=%d", len(a.particles))
	}
	many := options
	many.Count = 4097
	b.Emit(start, 10, 20, many)
	if len(b.particles) != 4097 {
		t.Fatalf("uncapped particle count=%d", len(b.particles))
	}
	theme.SetReducedMotion(true)
	a.Emit(start, 10, 20, options)
	if a.Active() {
		t.Fatal("reduced motion emitted particles")
	}
}

func TestEffectDrawingAllocs(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Now: time.Unix(84, 0), Constraints: layout.Exact(image.Pt(300, 200)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	style := normalizePointerGlow(PointerGlowStyle{})
	glow := PointerGlow{x: 120, y: 60}
	particle := confettiParticle{x: 100, y: 60, vx: 10, vy: -20, rotation: .2, angular: 1, size: 5, lifetime: 1, color: defaultConfettiColors[0]}
	confetti := Confetti{gravity: 500}
	glow.drawSpotlight(gtx, style, 1, 1)
	confetti.drawParticle(gtx, particle, .2, 1)
	if got := testing.AllocsPerRun(1000, func() {
		ops.Reset()
		glow.drawSpotlight(gtx, style, 1, 1)
		confetti.drawParticle(gtx, particle, .2, 1)
	}); got != 0 {
		t.Fatalf("effect drawing allocs=%g", got)
	}

	start := gtx.Now
	confetti.Emit(start, 100, 60, testConfettiOptions(256, 17))
	confetti.update(start.Add(50*time.Millisecond), false)
	ops.Reset()
	confetti.paint(gtx, noopDraw)
	if got := testing.AllocsPerRun(1000, func() {
		ops.Reset()
		confetti.update(start.Add(50*time.Millisecond), false)
		confetti.paint(gtx, noopDraw)
	}); got != 0 {
		t.Fatalf("steady confetti frame allocs=%g", got)
	}
}

func noopDraw() {}

func BenchmarkConfettiUpdate(b *testing.B) {
	start := time.Unix(85, 0)
	var c Confetti
	c.Emit(start, 100, 80, testConfettiOptions(1000, 7))
	particles := len(c.particles)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.particles = c.particles[:particles]
		c.update(start.Add(50*time.Millisecond), false)
	}
}

// Ensure the drawing helpers accept the same context type Decorate provides.
var _ core.C = layout.Context{}
