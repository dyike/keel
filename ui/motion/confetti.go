package motion

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// ConfettiOptions controls one deterministic particle burst. Count has no
// framework cap; the pool grows to the requested size and is reused.
type ConfettiOptions struct {
	Count                    int
	Seed                     uint64
	Gravity                  float32
	SpeedMin, SpeedMax       float32
	LifetimeMin, LifetimeMax time.Duration
	SizeMin, SizeMax         float32
	Colors                   []color.NRGBA
}

type confettiParticle struct {
	x, y, vx, vy      float32
	rotation, angular float32
	size              float32
	lifetime          float32
	color             color.NRGBA
	round             bool
}

// Confetti is a reusable particle pool. Positions and velocities are in dp and
// dp/second. Keep it in view state and call Apply on every frame of its host.
type Confetti struct {
	particles []confettiParticle
	started   time.Time
	gravity   float32
	elapsed   float32
	decorate  func(core.C, func())
}

// Active reports whether any particle is alive.
func (c *Confetti) Active() bool { return len(c.particles) > 0 }

// Emit starts a burst at an element-local point in dp. Seed makes the sequence
// reproducible. Reduced motion clears any existing burst and emits nothing.
func (c *Confetti) Emit(now time.Time, x, y float32, options ConfettiOptions) {
	if el.ReducedMotion() || options.Count <= 0 {
		c.particles = c.particles[:0]
		return
	}
	options = normalizeConfetti(options)
	if cap(c.particles) < options.Count {
		c.particles = make([]confettiParticle, options.Count)
	} else {
		c.particles = c.particles[:options.Count]
	}
	rng := particleRNG{state: options.Seed}
	if rng.state == 0 {
		rng.state = 0x9e3779b97f4a7c15
	}
	for i := range c.particles {
		angle := (-.9 + rng.float32()*.8) * float32(math.Pi)
		speed := mix(options.SpeedMin, options.SpeedMax, rng.float32())
		lifetime := mix(float32(options.LifetimeMin.Seconds()), float32(options.LifetimeMax.Seconds()), rng.float32())
		c.particles[i] = confettiParticle{
			x: x, y: y,
			vx:       float32(math.Cos(float64(angle))) * speed,
			vy:       float32(math.Sin(float64(angle))) * speed,
			rotation: rng.float32() * 2 * float32(math.Pi),
			angular:  mix(-9, 9, rng.float32()),
			size:     mix(options.SizeMin, options.SizeMax, rng.float32()),
			lifetime: lifetime,
			color:    options.Colors[int(rng.next()%uint64(len(options.Colors)))],
			round:    rng.next()&1 == 0,
		}
	}
	c.started, c.gravity = now, options.Gravity
}

// Apply paints the particles after e's content. It owns e's Decorate callback.
func (c *Confetti) Apply(cx *el.Context, e *el.DivEl) *el.DivEl {
	if e == nil {
		return nil
	}
	if c.decorate == nil {
		c.decorate = c.paint
	}
	if c.update(cx.Now(), el.ReducedMotion()) {
		cx.Animating()
	}
	e.Decorate(c.decorate)
	return e
}

func normalizeConfetti(options ConfettiOptions) ConfettiOptions {
	if options.Gravity == 0 {
		options.Gravity = 520
	}
	if options.SpeedMin <= 0 {
		options.SpeedMin = 150
	}
	if options.SpeedMax < options.SpeedMin {
		options.SpeedMax = max(options.SpeedMin, 310)
	}
	if options.LifetimeMin <= 0 {
		options.LifetimeMin = 650 * time.Millisecond
	}
	if options.LifetimeMax < options.LifetimeMin {
		options.LifetimeMax = max(options.LifetimeMin, 1100*time.Millisecond)
	}
	if options.SizeMin <= 0 {
		options.SizeMin = 3
	}
	if options.SizeMax < options.SizeMin {
		options.SizeMax = max(options.SizeMin, 7)
	}
	if len(options.Colors) == 0 {
		options.Colors = defaultConfettiColors[:]
	}
	return options
}

var defaultConfettiColors = [...]color.NRGBA{
	{R: 0x7c, G: 0x3a, B: 0xed, A: 0xff},
	{R: 0x06, G: 0xb6, B: 0xd4, A: 0xff},
	{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff},
	{R: 0xef, G: 0x44, B: 0x44, A: 0xff},
	{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff},
}

func (c *Confetti) paint(gtx core.C, draw func()) {
	draw()
	if len(c.particles) == 0 {
		return
	}
	scale := gtx.Metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	for i := range c.particles {
		c.drawParticle(gtx, c.particles[i], c.elapsed, scale)
	}
}

func (c *Confetti) drawParticle(gtx core.C, particle confettiParticle, t, scale float32) {
	x := (particle.x + particle.vx*t) * scale
	y := (particle.y + particle.vy*t + .5*c.gravity*t*t) * scale
	size := particle.size * scale
	alpha := clamp01((particle.lifetime - t) / min(particle.lifetime, .24))
	col := particle.color
	col.A = uint8(float32(col.A)*alpha + .5)
	transform := f32.AffineId().Offset(f32.Pt(x, y)).Rotate(f32.Point{}, particle.rotation+particle.angular*t)
	stack := op.Affine(transform).Push(gtx.Ops)
	half := size / 2
	rect := image.Rect(roundInt(-half), roundInt(-half*.55), roundInt(half), roundInt(half*.55))
	if particle.round {
		paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
	} else {
		paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
	}
	stack.Pop()
}

func (c *Confetti) update(now time.Time, reduced bool) bool {
	if reduced {
		c.particles = c.particles[:0]
		return false
	}
	t := float32(now.Sub(c.started).Seconds())
	if t < 0 {
		t = 0
	}
	c.elapsed = t
	write := 0
	for i := range c.particles {
		if t < c.particles[i].lifetime {
			c.particles[write] = c.particles[i]
			write++
		}
	}
	c.particles = c.particles[:write]
	return write > 0
}

type particleRNG struct{ state uint64 }

func (r *particleRNG) next() uint64 {
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 0x2545f4914f6cdd1d
}

func (r *particleRNG) float32() float32 {
	return float32(r.next()>>40) / float32(1<<24)
}

func mix(a, b, p float32) float32 { return a + (b-a)*p }
