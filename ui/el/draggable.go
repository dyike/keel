package el

import (
	"math"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/internal/animation"
)

// DragPoint is a translation in dp relative to the element's layout position.
type DragPoint struct{ X, Y float32 }

// DragBounds limits translations, not the element's layout rectangle.
type DragBounds struct{ MinX, MinY, MaxX, MaxY float32 }

// DragAxis constrains movement along the parent's coordinate axes.
type DragAxis uint8

const (
	DragBoth DragAxis = iota
	DragHorizontal
	DragVertical
)

// DragOptions configures a root-owned draggable box. Bounds and Snap slices
// are borrowed for this Render; keep them immutable while in use. Elastic is
// 0..1: zero clamps; positive values allow bounded rubber-band overshoot.
// Momentum projects 200ms of release velocity before choosing a bound/snap.
// Spring settling preserves velocity until it hits the elastic limit
// (100*Elastic dp beyond Bounds); a tween follows its curve.
// Reset changes jump to Initial and cancel an active drag. OnMove reports
// direct manipulation; OnRelease reports the chosen settling target.
type DragOptions struct {
	Initial   DragPoint
	Reset     uint64
	Axis      DragAxis
	Bounds    *DragBounds
	Snap      []DragPoint
	Elastic   float32
	Momentum  bool
	Animation Animation
	OnMove    func(DragPoint)
	OnRelease func(DragPoint)
}

// Draggable moves a Div without changing layout. Its transformed hit region,
// caret and anchors follow the box. Interactive descendants keep their input.
// It composes with AnimatePose and AnimateLayout, but owns OnDrag/DragAccept.
// Stable IDs preserve the position; omitting Draggable drops its state.
func (d *DivEl) Draggable(options DragOptions) *DivEl {
	if options.Axis > DragVertical || options.Elastic < 0 || options.Elastic > 1 || !finiteDrag(options.Elastic) || !finiteDrag(options.Initial.X) || !finiteDrag(options.Initial.Y) {
		panic("el: invalid drag options")
	}
	if b := options.Bounds; b != nil {
		if !finiteDrag(b.MinX) || !finiteDrag(b.MinY) || !finiteDrag(b.MaxX) || !finiteDrag(b.MaxY) || b.MinX > b.MaxX || b.MinY > b.MaxY {
			panic("el: invalid drag bounds")
		}
	}
	for _, p := range options.Snap {
		if !finiteDrag(p.X) || !finiteDrag(p.Y) {
			panic("el: invalid drag snap point")
		}
	}
	if options.Animation.Kind != animation.KindSpring && options.Animation.Kind != animation.KindTween || options.Animation.Repeat != 0 {
		panic("el: drag settling requires a finite spring or tween")
	}
	options.Animation = defaultAnimation(options.Animation)
	d.node().dragMotion = &options
	return d
}

func finiteDrag(v float32) bool        { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func acceptDrag(float32, float32) bool { return true }

type dragSample struct {
	at    time.Time
	point DragPoint
}
type dragMotionState struct {
	options                DragOptions
	values                 [2]animation.Value
	target, start, pressed DragPoint
	toParent               f32.Affine2D
	samples                [8]dragSample
	count                  int
	dragging, eventClock   bool
}

func (d *dragMotionState) constrain(p DragPoint) DragPoint {
	if b := d.options.Bounds; b != nil {
		p.X = min(max(p.X, b.MinX), b.MaxX)
		p.Y = min(max(p.Y, b.MinY), b.MaxY)
	}
	return p
}

func (d *dragMotionState) sample(now time.Time, reduced bool, scale float32) (f32.Affine2D, bool) {
	values := [2]float32{d.target.X, d.target.Y}
	active := false
	for i, target := range values {
		v := &d.values[i]
		if !v.Animating() && v.Get() == target && v.Velocity() == 0 {
			continue
		}
		var moving bool
		values[i], moving = v.Sample(now, target, d.options.Animation, reduced || d.options.Animation.Duration < 0)
		if b := d.options.Bounds; b != nil && !d.dragging {
			lo, hi := b.MinX, b.MaxX
			if i == 1 {
				lo, hi = b.MinY, b.MaxY
			}
			// Direct manipulation asymptotically approaches this same elastic
			// limit. Fast release events must not launch the box beyond it.
			margin := 100 * d.options.Elastic
			bounded := min(max(values[i], lo-margin), hi+margin)
			if bounded != values[i] {
				values[i] = bounded
				v.Set(bounded) // A collision absorbs outward velocity, including on re-grab.
				moving = bounded != target
			}
		}
		active = active || moving
	}
	return f32.AffineId().Offset(f32.Pt(values[0]*scale, values[1]*scale)), active
}

func (d *dragMotionState) cancel() {
	d.dragging, d.count = false, 0
	d.target = d.constrain(DragPoint{d.values[0].Get(), d.values[1].Get()})
	d.values[0].Set(d.target.X)
	d.values[1].Set(d.target.Y)
}

func (d *dragMotionState) track(now time.Time, p DragPoint) {
	if d.count > 0 && d.samples[d.count-1].at.Equal(now) {
		d.samples[d.count-1].point = p
		return
	}
	if d.count == len(d.samples) {
		copy(d.samples[:], d.samples[1:])
		d.count--
	}
	d.samples[d.count] = dragSample{now, p}
	d.count++
}
func (d *dragMotionState) velocity(now time.Time) DragPoint {
	if d.count < 2 {
		return DragPoint{}
	}
	last := d.samples[d.count-1]
	if now.Sub(last.at) > 100*time.Millisecond {
		return DragPoint{}
	}
	first := last
	for i := d.count - 2; i >= 0; i-- {
		if last.at.Sub(d.samples[i].at) > 100*time.Millisecond {
			break
		}
		first = d.samples[i]
	}
	dt := float32(last.at.Sub(first.at).Seconds())
	if dt <= 0 {
		return DragPoint{}
	}
	return DragPoint{(last.point.X - first.point.X) / dt, (last.point.Y - first.point.Y) / dt}
}

type dragNotifications struct {
	moved, released  bool
	position, target DragPoint
}

func (d *dragMotionState) handle(ev pointer.Event, now time.Time, scale float32, accepted bool) (notify dragNotifications) {
	if scale <= 0 {
		scale = 1
	}
	p := d.toParent.Transform(ev.Position)
	point := DragPoint{p.X / scale, p.Y / scale}
	current := DragPoint{d.values[0].Get(), d.values[1].Get()}
	// Native timestamps preserve velocity when several moves arrive in one frame.
	// Synthetic events may omit them; never mix the two time domains.
	at := now
	if timed := ev.Time > 0; timed {
		at = time.Unix(0, int64(ev.Time))
		if !d.eventClock {
			d.count = 0
		}
		d.eventClock = true
	} else if d.eventClock {
		d.count, d.eventClock = 0, false
	}
	switch ev.Kind {
	case pointer.Press:
		d.dragging, d.count = true, 0
		d.start, d.pressed, d.target = current, point, current
		d.values[0].Set(current.X)
		d.values[1].Set(current.Y)
		d.track(at, current)
	case pointer.Drag:
		if !d.dragging || !accepted {
			return
		}
		notify.moved = d.move(point, current, at)
		notify.position = d.target
	case pointer.Release, pointer.Cancel:
		if !d.dragging {
			return
		}
		d.dragging = false
		if !accepted {
			return
		}
		if ev.Kind == pointer.Release {
			notify.moved = d.move(point, current, at)
			notify.position = d.target
			current = DragPoint{d.values[0].Get(), d.values[1].Get()}
		}
		velocity := DragPoint{}
		if d.options.Momentum && ev.Kind != pointer.Cancel {
			velocity = d.velocity(at)
		}
		target := DragPoint{current.X + velocity.X*.2, current.Y + velocity.Y*.2}
		if ev.Kind == pointer.Cancel {
			target = d.start
			velocity = DragPoint{}
		} else if len(d.options.Snap) > 0 {
			best := math.Inf(1)
			var snap DragPoint
			for _, candidate := range d.options.Snap {
				p := d.constrain(candidate)
				if d.options.Axis == DragHorizontal {
					p.Y = d.start.Y
				}
				if d.options.Axis == DragVertical {
					p.X = d.start.X
				}
				dx, dy := float64(target.X)-float64(p.X), float64(target.Y)-float64(p.Y)
				if distance := dx*dx + dy*dy; distance < best {
					best, snap = distance, p
				}
			}
			target = snap
		}
		if d.options.Axis == DragHorizontal {
			target.Y = d.start.Y
			velocity.Y = 0
		}
		if d.options.Axis == DragVertical {
			target.X = d.start.X
			velocity.X = 0
		}
		d.target = d.constrain(target)
		d.values[0].SetVelocity(current.X, velocity.X)
		d.values[1].SetVelocity(current.Y, velocity.Y)
		notify.released, notify.target = true, d.target
	}
	return
}

func (e *engine) dragScale() float32 {
	if e.m.PxPerDp > 0 {
		return e.m.PxPerDp
	}
	return 1
}

func (st *elemState) prepareDrag(options DragOptions) {
	if st.dragMotion == nil || st.dragMotion.options.Reset != options.Reset {
		st.dragMotion = &dragMotionState{options: options}
		d := st.dragMotion
		d.target = d.constrain(options.Initial)
		d.values[0].Set(d.target.X)
		d.values[1].Set(d.target.Y)
		st.conditionalDrag = conditionalDrag{}
		st.fresh = true
	} else {
		d := st.dragMotion
		d.options = options
		if !d.dragging {
			d.target = d.constrain(d.target)
		}
	}
}

func (n *Node) dragPredicate() func(float32, float32) bool {
	if n.dragMotion != nil {
		return acceptDrag
	}
	return n.dragAccept
}

func (d *dragMotionState) move(point, current DragPoint, at time.Time) bool {
	next := DragPoint{d.start.X + point.X - d.pressed.X, d.start.Y + point.Y - d.pressed.Y}
	if d.options.Axis == DragHorizontal {
		next.Y = d.start.Y
	}
	if d.options.Axis == DragVertical {
		next.X = d.start.X
	}
	bounded := d.constrain(next)
	rubber := func(x, limit float32) float32 {
		delta := x - limit
		return limit + delta*d.options.Elastic/(1+float32(math.Abs(float64(delta)))/100)
	}
	next = DragPoint{rubber(next.X, bounded.X), rubber(next.Y, bounded.Y)}
	d.target = next
	d.values[0].Set(next.X)
	d.values[1].Set(next.Y)
	d.track(at, next)
	return next != current
}
