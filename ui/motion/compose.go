package motion

import (
	"image/color"
	"math"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// Vec2 is a pair of independently animated values.
type Vec2 struct{ X, Y Value }

// Affine describes a paint transform. Rotate is in radians; origin values are
// fractions of the element box. Zero scale fields default to one.
type Affine struct {
	Scale, ScaleX, ScaleY float32
	Rotate                float32
	OriginX, OriginY      float32
}

// Transform decorates an element with a paint transform. Pass the embedded
// Styled field, for example motion.Transform(&box.Styled, affine).
func Transform[T any](styled *el.Styled[T], a Affine) *T {
	scale := a.Scale
	if scale == 0 {
		scale = 1
	}
	sx, sy := a.ScaleX, a.ScaleY
	if sx == 0 {
		sx = scale
	}
	if sy == 0 {
		sy = scale
	}
	ox, oy := a.OriginX, a.OriginY
	if ox == 0 && oy == 0 {
		ox, oy = .5, .5
	}
	return styled.Decorate(func(gtx core.C, draw func()) {
		origin := f32.Pt(float32(gtx.Constraints.Max.X)*ox, float32(gtx.Constraints.Max.Y)*oy)
		transform := f32.AffineId().Scale(origin, f32.Pt(sx, sy)).Rotate(origin, a.Rotate)
		defer op.Affine(transform).Push(gtx.Ops).Pop()
		draw()
	})
}

// MixColor interpolates RGB in squared sRGB space and alpha linearly.
func MixColor(a, b color.NRGBA, p float32) color.NRGBA {
	p = clamp01(p)
	mix := func(x, y uint8) uint8 {
		xx, yy := float32(x), float32(y)
		return uint8(math.Round(math.Sqrt(float64(xx*xx + (yy*yy-xx*xx)*p))))
	}
	alpha := float32(a.A) + (float32(b.A)-float32(a.A))*p
	return color.NRGBA{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: uint8(math.Round(float64(alpha)))}
}

// MapRange maps x between two ranges. Degenerate input ranges return outMin.
func MapRange(x, inMin, inMax, outMin, outMax float32, clamp bool) float32 {
	if inMin == inMax {
		return outMin
	}
	p := (x - inMin) / (inMax - inMin)
	if clamp {
		p = clamp01(p)
	}
	return outMin + (outMax-outMin)*p
}

type dragSample struct {
	at   time.Time
	x, y float32
}

// DragTracker estimates release velocity from a fixed, allocation-free sample
// window. Pass Context.Now and each DragEvent to Track.
type DragTracker struct {
	samples [8]dragSample
	n       int
}

func (d *DragTracker) Reset() { d.n = 0 }

func (d *DragTracker) Track(now time.Time, event el.DragEvent) {
	if event.Kind == el.DragStart {
		d.Reset()
	}
	if d.n > 0 && !now.After(d.samples[d.n-1].at) {
		d.samples[d.n-1] = dragSample{now, event.X, event.Y}
		return
	}
	if d.n == len(d.samples) {
		copy(d.samples[:], d.samples[1:])
		d.n--
	}
	d.samples[d.n] = dragSample{now, event.X, event.Y}
	d.n++
}

// Velocity returns dp per second over the most recent roughly 100ms.
func (d *DragTracker) Velocity() (x, y float32) {
	if d.n < 2 {
		return 0, 0
	}
	last := d.samples[d.n-1]
	first := d.samples[0]
	for i := d.n - 2; i >= 0; i-- {
		if last.at.Sub(d.samples[i].at) > 100*time.Millisecond {
			break
		}
		first = d.samples[i]
	}
	dt := float32(last.at.Sub(first.at).Seconds())
	if dt <= 0 {
		return 0, 0
	}
	return (last.x - first.x) / dt, (last.y - first.y) / dt
}
