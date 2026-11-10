package el

import (
	"image"
	"math"
	"reflect"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/internal/animation"
)

// SharedLayout connects successive elements with the same identity across
// parents in one Root. The incoming element starts at the previous element's
// visible position, size and rotation. Its destination content scales with the
// box; layout uses its final size immediately. Clips, opacity and paint order
// remain those of the incoming parent. Pointer, caret and anchor coordinates
// follow the moving box. Base solid colors, corners, border width and opacity
// interpolate too. Text with ContentTransition transfers its displayed glyphs.
//
// Use a comparable, non-nil identity, preferably a private application type.
// Only one visible element per identity may be built at a time. The first
// appearance, an absence of one frame, resizing the viewport, density changes
// and reduced motion snap. A stable owner follows scrolling without starting a
// transition. Omit the declaration to release the root-owned geometry; no old
// element tree or drawing resources are retained.
func (s *Styled[T]) SharedLayout(identity any, a Animation) *T {
	if identity == nil || !reflect.ValueOf(identity).Comparable() || identity != identity {
		panic("el: shared layout identity must be non-nil and comparable")
	}
	if a.Kind != animation.KindSpring && a.Kind != animation.KindTween || a.Repeat != 0 {
		panic("el: shared layout requires a non-repeating spring or tween")
	}
	s.node().sharedLayout = &sharedLayoutSpec{identity, defaultAnimation(a)}
	return s.self
}

type sharedLayoutSpec struct {
	identity  any
	animation Animation
}

type sharedLayoutState struct {
	owner          stateKey
	frame, visible uint64
	metric         unit.Metric
	viewport       image.Point
	shown          f32.Affine2D
	values         [6]animation.Value // residual from the current destination
	look, lookFrom sharedAppearance
	lookProgress   animation.Value
	lookActive     bool
	content        *textContentState
	contentSize    image.Point
	contentTime    time.Time
	contentFrom    []contentGlyph
	contentHandoff bool
}

// sharedComponents decomposes a unit-box transform about its center into
// position, angle, scales and shear. Scales are clamped away from zero during
// spring overshoot. A half turn does not collapse the box as linear
// interpolation of matrix coefficients would.
func sharedComponents(m f32.Affine2D) ([6]float32, bool) {
	a, b, _, c, d, _ := m.Elems()
	sx := float32(math.Hypot(float64(a), float64(c)))
	det := a*d - b*c
	if sx < 1e-6 || det <= 1e-6 {
		return [6]float32{}, false
	}
	sy := det / sx
	center := m.Transform(f32.Pt(.5, .5))
	v := [6]float32{center.X, center.Y, float32(math.Atan2(float64(c), float64(a))), sx, sy, (a*b + c*d) / (sx * sy)}
	for _, x := range v {
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			return [6]float32{}, false
		}
	}
	return v, true
}

func sharedTransform(v [6]float32) f32.Affine2D {
	sx, sy := max(v[3], .0001), max(v[4], .0001)
	sin, cos := math.Sincos(float64(v[2]))
	s, c := float32(sin), float32(cos)
	a, b, d, e := c*sx, (c*v[5]-s)*sy, s*sx, (s*v[5]+c)*sy
	return f32.NewAffine2D(a, b, v[0]-(a+b)/2, d, e, v[1]-(d+e)/2)
}

func (e *engine) sampleSharedLayout(n *Node, local f32.Affine2D, origin image.Point) (f32.Affine2D, bool) {
	if !e.frameSource.Enabled() || n.size.X <= 0 || n.size.Y <= 0 {
		return local, false
	}
	spec := n.sharedLayout
	parent := e.poseWorld.Mul(f32.AffineId().Offset(f32.Pt(float32(origin.X), float32(origin.Y))))
	unit := parent.Mul(local).Mul(f32.AffineId().Scale(f32.Point{}, f32.Pt(float32(n.size.X), float32(n.size.Y))))
	target, valid := sharedComponents(unit)
	if !valid {
		return local, false
	}
	if e.sharedLayouts == nil {
		e.sharedLayouts = make(map[any]*sharedLayoutState)
	}
	s := e.sharedLayouts[spec.identity]
	if s != nil && s.frame == e.paintSerial && s.owner != n.key {
		panic("el: shared layout identity appears more than once in a frame")
	}
	look := e.sharedAppearance(n)
	snap := ReducedMotion() || spec.animation.Duration < 0
	if s == nil || s.frame+1 < e.paintSerial || s.metric != e.m || s.viewport != e.gtx.Constraints.Max || snap {
		s = &sharedLayoutState{owner: n.key, metric: e.m, viewport: e.gtx.Constraints.Max, shown: unit, look: look}
		e.sharedLayouts[spec.identity] = s
	} else if s.owner != n.key {
		s.handoffContent(n)
		s.lookFrom = s.look
		s.lookActive = s.look != look
		s.lookProgress.Set(0)
		from, ok := sharedComponents(s.shown)
		if ok {
			for i := range from {
				delta := from[i] - target[i]
				if i == 2 {
					delta = float32(math.Remainder(float64(delta), 2*math.Pi))
				}
				s.values[i].SetVelocity(delta, s.values[i].Velocity())
			}
		}
		s.owner = n.key
	}
	if n.content == nil {
		s.content, s.contentFrom, s.contentHandoff = nil, nil, false
	}
	s.frame = e.paintSerial
	active := false
	if s.lookActive {
		p, moving := s.lookProgress.Sample(e.gtx.Now, 1, spec.animation, snap)
		s.look, s.lookActive = mixSharedAppearance(s.lookFrom, look, p), moving
	} else {
		s.look = look
	}
	for i := range target {
		v := &s.values[i]
		if !v.Animating() && v.Get() == 0 && v.Velocity() == 0 {
			continue
		}
		x, moving := v.Sample(e.gtx.Now, 0, spec.animation, snap)
		target[i] += x
		active = active || moving
	}
	if !active {
		// Preserve exact resting pixels, including fractional native text phase.
		s.shown = unit
		return local, s.lookActive
	}
	s.shown = sharedTransform(target)
	return parent.Invert().Mul(sharedBoxTransform(s.shown, n.size)), true
}

func (e *engine) markSharedLayout(n *Node) {
	if s := e.sharedLayouts[n.sharedLayout.identity]; s != nil && s.owner == n.key {
		s.visible = e.paintSerial
	}
}

func (e *engine) sweepSharedLayouts() {
	for key, s := range e.sharedLayouts {
		if s.visible != e.paintSerial {
			delete(e.sharedLayouts, key)
		}
	}
	if len(e.sharedLayouts) == 0 {
		e.sharedLayouts = nil
	}
}

// Avoid constructing a tiny reciprocal-scale matrix: Affine2D stores the
// diagonal minus one, which loses precision near zero (and native text phase).
func sharedBoxTransform(m f32.Affine2D, size image.Point) f32.Affine2D {
	a, b, x, c, d, y := m.Elems()
	w, h := float32(size.X), float32(size.Y)
	return f32.NewAffine2D(a/w, b/h, x, c/w, d/h, y)
}
