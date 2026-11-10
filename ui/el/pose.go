package el

import (
	"image"
	"math"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/ui/internal/animation"
)

// Animation describes a frame-time spring, tween, keyframe or inertia segment.
// It is also accepted by ui/motion.Value. Keep its slices immutable.
type Animation = animation.Transition

// Easing maps normalized time to progress.
type Easing = animation.Easing

var (
	Linear    = animation.Linear
	EaseIn    = animation.EaseIn
	EaseOut   = animation.EaseOut
	EaseInOut = animation.EaseInOut
	Smooth    = animation.SpringDuration(500*time.Millisecond, 0)
	Snappy    = animation.SpringDuration(500*time.Millisecond, .15)
	Bouncy    = animation.SpringDuration(500*time.Millisecond, .3)
	// Instant applies externally sampled or directly manipulated poses immediately.
	Instant = Animation{Duration: -1}
)

// Tween moves along an easing curve for duration. A nil ease uses EaseInOut.
func Tween(duration time.Duration, ease Easing) Animation { return animation.Tween(duration, ease) }

// Spring uses physical stiffness and damping; mass defaults to one.
func Spring(stiffness, damping float32) Animation { return animation.Spring(stiffness, damping) }

// CubicBezier constructs an easing curve with CSS control points.
func CubicBezier(x1, y1, x2, y2 float32) Easing { return animation.CubicBezier(x1, y1, x2, y2) }

// Pose describes a visual transform around the element's center. X/Y are dp;
// XPercent/YPercent are percentages of its laid-out size. Rotate is degrees.
// Scale zero means one. Fade zero is opaque, one is transparent. Poses do not
// change layout. Pointer routing, semantics and anchored overlays follow them.
type Pose struct {
	X, Y, XPercent, YPercent float32
	Scale, Rotate, Fade      float32
}

func (p Pose) values() [7]float32 {
	s := p.Scale
	if s == 0 {
		s = 1
	}
	return [7]float32{p.X, p.Y, p.XPercent, p.YPercent, s, p.Rotate, p.Fade}
}

type poseSpec struct {
	target    Pose
	animation Animation
}
type poseState struct{ values [7]animation.Value }

// AnimatePose declares a visual target without reading frame time in Render.
// The first declaration shows its target immediately; later targets animate
// from the current value and velocity. A window-owned root can advance it in
// paint-only frames. Use stable IDs when sibling order changes. Omitting the
// declaration drops its state. It composes with Translate, Opacity and Decorate.
func (s *Styled[T]) AnimatePose(p Pose, a Animation) *T {
	validatePose(p)
	s.node().pose = &poseSpec{p, defaultAnimation(a)}
	return s.self
}

func validatePose(p Pose) {
	for _, v := range p.values() {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			panic("el: pose values must be finite")
		}
	}
	if p.Scale < 0 || p.Fade < 0 || p.Fade > 1 {
		panic("el: invalid pose scale or fade")
	}
}

func defaultAnimation(a Animation) Animation {
	if a.Kind == animation.KindSpring && a.Stiffness == 0 && a.Damping == 0 && a.Mass == 0 {
		// Keep explicit delay and rest thresholds while applying the default spring.
		a.Stiffness, a.Damping, a.Mass = Smooth.Stiffness, Smooth.Damping, Smooth.Mass
	}
	return a
}

func (e *engine) samplePose(n *Node) (f32.Affine2D, float32, bool) {
	st := e.store.get(n.key)
	v := n.pose.target.values()
	if st.pose == nil {
		st.pose = new(poseState)
		for i, target := range v {
			st.pose.values[i].Set(target)
		}
	}
	active := false
	for i, target := range v {
		if !st.pose.values[i].Animating() && st.pose.values[i].Get() == target {
			continue
		}
		var moving bool
		v[i], moving = st.pose.values[i].Sample(e.gtx.Now, target, n.pose.animation, ReducedMotion() || n.pose.animation.Duration < 0)
		active = active || moving
	}
	m := e.poseTransform(n, v)
	return m, min(max(v[6], 0), 1), active
}

func (e *engine) poseTransform(n *Node, v [7]float32) f32.Affine2D {
	scale := e.m.PxPerDp
	if scale == 0 {
		scale = 1
	}
	center := f32.Pt(float32(n.size.X)/2, float32(n.size.Y)/2)
	// Spring overshoot must not create a singular coordinate system.
	s := max(v[4], .0001)
	m := f32.AffineId().Scale(center, f32.Pt(s, s)).Rotate(center, v[5]*math.Pi/180)
	m = m.Offset(f32.Pt(v[0]*scale+v[2]*float32(n.size.X)/100, v[1]*scale+v[3]*float32(n.size.Y)/100))
	return m
}

// affineBounds conservatively bounds a rectangle, including nested rotations.
// Gio retains the exact transformed clip and hit region.
func affineBounds(m f32.Affine2D, r image.Rectangle) image.Rectangle {
	x0, y0 := float32(r.Min.X), float32(r.Min.Y)
	x1, y1 := float32(r.Max.X), float32(r.Max.Y)
	a, b, c, d := m.Transform(f32.Pt(x0, y0)), m.Transform(f32.Pt(x1, y0)), m.Transform(f32.Pt(x0, y1)), m.Transform(f32.Pt(x1, y1))
	return image.Rect(int(math.Floor(float64(min(a.X, b.X, c.X, d.X)))), int(math.Floor(float64(min(a.Y, b.Y, c.Y, d.Y)))), int(math.Ceil(float64(max(a.X, b.X, c.X, d.X)))), int(math.Ceil(float64(max(a.Y, b.Y, c.Y, d.Y)))))
}
