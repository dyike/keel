package kit

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestRisingGlowReferenceMotion(t *testing.T) {
	p := risingParticle{size: 6, delay: 1, duration: 4}
	for _, tc := range []struct{ at, y, a float64 }{{0, 84, 0}, {1, 84, 0}, {2, 39, .5}, {3, -6, 1}, {4, 39, .5}, {5, 84, 0}} {
		y, a := p.sample(tc.at, 100)
		if math.Abs(y-tc.y) > .01 || math.Abs(a-tc.a) > .001 {
			t.Fatalf("t=%v: y=%v alpha=%v", tc.at, y, a)
		}
	}
}
func TestRisingGlowConfiguration(t *testing.T) {
	v := RisingGlow().ParticleCount(80)
	v.prepare()
	first := v.particles[0]
	v.ParticleCount(80)
	v.prepare()
	if first != v.particles[0] {
		t.Fatal("particles randomized on unchanged count")
	}
	for _, p := range v.particles {
		if p.left < 0 || p.left >= 1 || p.size < 4 || p.size >= 10 || p.delay < 0 || p.delay >= 2 || p.duration < 3 || p.duration >= 5 {
			t.Fatal(p)
		}
	}
	v.Width(float32(math.NaN())).Height(float32(math.Inf(1))).ParticleCount(513)
	if v.width != 0 || v.height != 120 || v.count != 80 {
		t.Fatal("invalid config accepted")
	}
	v.ParticleCount(0)
	v.prepare()
	if len(v.particles) != 0 {
		t.Fatal("zero count retained particles")
	}
}

func TestRisingGlowSchedulingAndInteraction(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	now := time.Unix(100, 0)
	v := RisingGlow().Width(200).Height(100)
	renders := 0
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		renders++
		return el.Div().Child(v.Render(cx), Button("Pause", func() { v.Enabled(false) }).Render(cx))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	h.Router.WakeupTime()
	now = now.Add(time.Second)
	h.Frame()
	if _, ok := h.Router.PeekWakeup(); !ok {
		t.Fatal("animation did not request a frame")
	}
	if renders != 1 {
		t.Fatalf("ambient frame rebuilt tree %d times", renders)
	}
	click(t, h, "Pause")
	now = now.Add(2 * time.Second) // Let the button press transition settle.
	h.Frame()
	h.Router.WakeupTime()
	h.Frame()
	if _, ok := h.Router.PeekWakeup(); ok {
		t.Fatal("paused effect requested a frame")
	}
	v.Enabled(true)
	theme.SetReducedMotion(true)
	h.Router.WakeupTime()
	h.Frame()
	if _, ok := h.Router.PeekWakeup(); ok {
		t.Fatal("reduced motion requested a frame")
	}
}

func TestRisingGlowHiddenDoesNotAnimate(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	for _, mode := range []string{"clipped", "empty", "transparent"} {
		v := RisingGlow()
		if mode == "empty" {
			v.ParticleCount(0)
		}
		if mode == "transparent" {
			v.Color(color.NRGBA{})
		}
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			box := el.Div()
			if mode == "clipped" {
				box.Child(el.Div().H(el.Dp(500)).NoShrink())
			}
			return box.Child(v.Render(cx))
		}))
		h := uitest.New(root)
		h.Router.WakeupTime()
		h.Frame()
		if _, ok := h.Router.PeekWakeup(); ok {
			t.Errorf("%s requested a frame", mode)
		}
	}
}

func BenchmarkRisingGlowPaint(b *testing.B) {
	for _, n := range []int{20, 80, 512} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			v := RisingGlow().ParticleCount(n)
			root := el.Root(v)
			root.RetainFrames()
			now := time.Unix(100, 0)
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				gtx.Metric = unit.Metric{PxPerDp: 1.25, PxPerSp: 1.25}
				gtx.Constraints.Max = image.Pt(480, 150)
				root.Layout(gtx)
			})
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				now = now.Add(el.AmbientFrameInterval)
				h.Frame()
			}
		})
	}
}
