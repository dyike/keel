package kit

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// RisingGlowView is a decorative field of fading, vertically cycling particles.
// It has no text or input semantics; place a heading above it.
type RisingGlowView struct {
	width, height float32
	count         int
	color         *color.NRGBA
	disabled      bool
	particles     []risingParticle
	start         time.Time
}

type risingParticle struct{ left, size, delay, duration float64 }

// RisingGlow fills the available width, with a height of 120dp and 20 particles.
func RisingGlow() *RisingGlowView { return &RisingGlowView{height: 120, count: 20} }

// Width sets dp width; zero fills the available width.
func (v *RisingGlowView) Width(dp float32) *RisingGlowView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}
func (v *RisingGlowView) Height(dp float32) *RisingGlowView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// ParticleCount accepts 0..512. Zero hides particles and stops animation.
func (v *RisingGlowView) ParticleCount(n int) *RisingGlowView {
	if n >= 0 && n <= 512 && n != v.count {
		v.count = n
		v.particles = nil
		v.start = time.Time{}
	}
	return v
}

// Color overrides the theme's primary text color.
func (v *RisingGlowView) Color(c color.NRGBA) *RisingGlowView { v.color = &c; return v }

// Enabled(false) shows a still representative pose, as reduced motion does.
// Re-enabling starts a fresh cycle.
func (v *RisingGlowView) Enabled(on bool) *RisingGlowView {
	if v.disabled == on {
		v.start = time.Time{}
	}
	v.disabled = !on
	return v
}

func (v *RisingGlowView) prepare() {
	if len(v.particles) == v.count {
		return
	}
	v.particles = make([]risingParticle, v.count)
	for i := range v.particles {
		v.particles[i] = risingParticle{left: rand.Float64(), size: 4 + 6*rand.Float64(), delay: 2 * rand.Float64(), duration: 3 + 2*rand.Float64()}
	}
}

func (v *RisingGlowView) Render(cx *el.Context) el.Element {
	v.prepare()
	graphic := el.Widget(core.Func(v.layout)).H(el.Dp(v.height))
	if v.width > 0 {
		graphic.W(el.Dp(v.width))
	} else {
		graphic.WFull()
	}
	return graphic
}

func (v *RisingGlowView) layout(gtx core.C) core.D {
	size := gtx.Constraints.Min
	if size.X <= 0 || size.Y <= 0 || len(v.particles) == 0 {
		return core.D{Size: size}
	}
	c := theme.PrimaryText
	if v.color != nil {
		c = *v.color
	}
	if c.A == 0 {
		return core.D{Size: size}
	}
	elapsed := 2.0 // Stable, visible pose when animation is off.
	if !v.disabled && !el.ReducedMotion() {
		if v.start.IsZero() || gtx.Now.Before(v.start) {
			v.start = gtx.Now
		}
		elapsed = gtx.Now.Sub(v.start).Seconds()
		el.AmbientPhase(gtx, time.Second) // Paint-only ticks; no timers or tree rebuild.
	}
	density := gtx.Metric.PxPerDp
	if density <= 0 {
		density = 1
	}
	defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
	for _, p := range v.particles {
		y, alpha := p.sample(elapsed, float64(size.Y)/float64(density))
		if alpha <= 0 {
			continue
		}
		diameter := max(1, int(math.Round(p.size*float64(density))))
		offset := op.Affine(f32.AffineId().Offset(f32.Pt(float32(p.left)*float32(size.X), float32(y)*density))).Push(gtx.Ops)
		tint := c
		tint.A = uint8(math.Round(float64(c.A) * alpha))
		paint.FillShape(gtx.Ops, tint, clip.Ellipse(image.Rect(0, 0, diameter, diameter)).Op(gtx.Ops))
		offset.Pop()
	}
	return core.D{Size: size}
}

// The reference keyframes are y=[-10,-height,-10], opacity=[0,1,0],
// with CSS ease-in-out applied separately to each half of the cycle.
func (p risingParticle) sample(elapsed, height float64) (top, alpha float64) {
	if elapsed < p.delay {
		return height - p.size - 10, 0
	}
	phase := math.Mod(elapsed-p.delay, p.duration) / p.duration
	t := phase * 2
	if t > 1 {
		t = 2 - t
	}
	alpha = risingEase(t)
	return height - p.size - 10 - (height-10)*alpha, alpha
}

// Invert cubic-bezier(.42,0,.58,1) using bounded bisection.
func risingEase(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	lo, hi := 0.0, 1.0
	for range 16 {
		t := (lo + hi) / 2
		bx := 3*(1-t)*(1-t)*t*.42 + 3*(1-t)*t*t*.58 + t*t*t
		if bx < x {
			lo = t
		} else {
			hi = t
		}
	}
	t := (lo + hi) / 2
	return t * t * (3 - 2*t)
}
