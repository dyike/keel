package el

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/event"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/unit"
	"image"
)

// DragAccept decides once, after the initial movement exceeds 3dp, whether
// this element captures a drag. dx/dy are pointer displacement in dp (not
// scroll deltas). Rejection emits a canceled DragEnd and leaves enclosing
// handlers free to capture. Nil restores ordinary OnDrag behavior.
// Presses on interactive descendants are left to those descendants.
// The predicate must not mutate UI state. Use with OnDrag.
func (s *Styled[T]) DragAccept(fn func(dx, dy float32) bool) *T {
	s.n.dragAccept = fn
	return s.self
}

type conditionalDrag struct {
	active, decided bool
	blocked         []image.Rectangle
	id              pointer.ID
	start           f32.Point
}

func (d *conditionalDrag) add(ops *op.Ops) { event.Op(ops, d) }
func (d *conditionalDrag) update(m unit.Metric, source input.Source, accept func(float32, float32) bool) (pointer.Event, bool) {
	for {
		ev, ok := source.Event(pointer.Filter{Target: d, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			return pointer.Event{}, false
		}
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Kind {
		case pointer.Press:
			blocked := false
			for _, rect := range d.blocked {
				if image.Pt(int(e.Position.X), int(e.Position.Y)).In(rect) {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}
			if d.active || e.Source != pointer.Touch && e.Buttons != pointer.ButtonPrimary {
				continue
			}
			d.active, d.decided, d.id, d.start = true, false, e.PointerID, e.Position
		case pointer.Drag:
			if !d.active || e.PointerID != d.id {
				continue
			}
			if !d.decided {
				diff := e.Position.Sub(d.start)
				slop := float32(m.Dp(3))
				if diff.X*diff.X+diff.Y*diff.Y >= slop*slop {
					d.decided = true
					scale := m.PxPerDp
					if scale <= 0 {
						scale = 1
					}
					if !accept(diff.X/scale, diff.Y/scale) {
						d.active = false
						e.Kind = pointer.Cancel
						return e, true
					}
				}
			}
			if d.decided && e.Priority < pointer.Grabbed {
				source.Execute(pointer.GrabCmd{Tag: d, ID: e.PointerID})
			}
		case pointer.Release, pointer.Cancel:
			if !d.active || e.Kind != pointer.Cancel && e.PointerID != d.id {
				continue
			}
			d.active = false
		}
		return e, true
	}
}

// Descendant interactive areas are recorded in the conditional parent's local
// coordinates. This prevents competing captures regardless of dispatch order.
type dragScope struct {
	drag   *conditionalDrag
	origin image.Point
}
