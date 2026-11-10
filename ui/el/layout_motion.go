package el

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/internal/animation"
)

// AnimateLayout animates changes to this element's position and nonzero size
// within its parent. Layout immediately uses the destination geometry; painting
// maps it from the previous visible box to that destination. Text and children
// scale with the box during a size change. Pointer/caret/anchor geometry follows
// the same transform. Use stable IDs for reordered siblings. Scrolling does not
// start a layout transition. Changing parents creates a new identity.
//
// Only finite springs and tweens are accepted. It composes with AnimatePose;
// omit the declaration to drop its state. The first layout and density changes
// snap to their destination. A zero-size axis snaps because it has no drawable
// content to stretch.
func (s *Styled[T]) AnimateLayout(a Animation) *T {
	if a.Kind != animation.KindSpring && a.Kind != animation.KindTween || a.Repeat != 0 {
		panic("el: layout animation requires a non-repeating spring or tween")
	}
	a = defaultAnimation(a)
	s.node().layoutMotion = &a
	return s.self
}

type layoutMotionState struct {
	values [4]animation.Value
	metric unit.Metric
}

func (e *engine) sampleLayoutMotion(n *Node) (f32.Affine2D, bool) {
	st := e.store.get(n.key)
	targets := [4]float32{float32(n.pos.X), float32(n.pos.Y), float32(n.size.X), float32(n.size.Y)}
	if st.layoutMotion == nil || st.layoutMotion.metric != e.m {
		st.layoutMotion = &layoutMotionState{metric: e.m}
		for i, target := range targets {
			st.layoutMotion.values[i].Set(target)
		}
	}
	values, active := targets, false
	for i, target := range targets {
		v := &st.layoutMotion.values[i]
		if i >= 2 && (target == 0 || v.Get() == 0) {
			v.Set(target)
		}
		if !v.Animating() && v.Get() == target {
			continue
		}
		var moving bool
		values[i], moving = v.Sample(e.gtx.Now, target, *n.layoutMotion, ReducedMotion() || n.layoutMotion.Duration < 0)
		active = active || moving
	}
	sx, sy := float32(1), float32(1)
	if n.size.X > 0 {
		sx = max(values[2]/float32(n.size.X), .0001)
	}
	if n.size.Y > 0 {
		sy = max(values[3]/float32(n.size.Y), .0001)
	}
	m := f32.AffineId().Scale(f32.Point{}, f32.Pt(sx, sy)).Offset(f32.Pt(values[0]-targets[0], values[1]-targets[1]))
	return m, active
}

func (e *engine) sampleVisualMotion(n *Node) (f32.Affine2D, float32, bool) {
	m, fade, active := f32.AffineId(), float32(0), false
	if n.pose != nil {
		m, fade, active = e.samplePose(n)
	}
	if n.layoutMotion != nil {
		layout, moving := e.sampleLayoutMotion(n)
		m, active = layout.Mul(m), active || moving
	}
	if n.dragMotion != nil {
		st := e.store.get(n.key)
		if st.dragMotion == nil {
			st.prepareDrag(*n.dragMotion)
		}
		drag, moving := st.dragMotion.sample(e.gtx.Now, ReducedMotion(), e.dragScale())
		m, active = drag.Mul(m), active || moving
	}
	return m, fade, active
}
