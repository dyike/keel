package motion

import (
	"image"
	"math"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

type layoutState struct {
	origin image.Point
	x, y   Value
	init   bool
}

// LayoutGroup animates identified boxes when their laid-out position changes.
// It currently supports Div elements and position only. Do not combine it with
// another Decorate callback or move it between scroll coordinate spaces.
type LayoutGroup struct {
	states map[string]*layoutState
}

// Animate applies position-only FLIP animation to e. The first frame at a new
// position is compensated during paint; subsequent frames animate to zero.
func (g *LayoutGroup) Animate(cx *el.Context, id string, e *el.DivEl, tr Transition) *el.DivEl {
	if e == nil || id == "" {
		return e
	}
	if g.states == nil {
		g.states = make(map[string]*layoutState)
	}
	state := g.states[id]
	if state == nil {
		state = new(layoutState)
		state.x.Set(0)
		state.y.Set(0)
		g.states[id] = state
	}
	now, reduced := cx.Now(), el.ReducedMotion()
	x, animateX := state.x.toAt(now, 0, tr, reduced)
	y, animateY := state.y.toAt(now, 0, tr, reduced)
	if animateX || animateY {
		cx.Animating()
	}
	e.Translate(x, y)
	e.Decorate(func(gtx core.C, draw func()) {
		origin, _ := cx.PaintGeometry()
		scale := cx.PixelScale()
		applied := image.Pt(int(math.Round(float64(x*scale))), int(math.Round(float64(y*scale))))
		natural := origin.Sub(applied)
		if !state.init {
			state.origin, state.init = natural, true
			draw()
			return
		}
		delta := natural.Sub(state.origin)
		state.origin = natural
		if delta == (image.Point{}) {
			draw()
			return
		}
		state.x.Set(x - float32(delta.X)/scale)
		state.y.Set(y - float32(delta.Y)/scale)
		if !reduced {
			cx.Animating()
			defer op.Offset(delta.Mul(-1)).Push(gtx.Ops).Pop()
		}
		draw()
	})
	return e
}

// Forget drops geometry retained for id, for long-lived groups with transient
// identifiers. Presence-managed lists can call it after an exit completes.
func (g *LayoutGroup) Forget(id string) { delete(g.states, id) }

// Reset drops every retained position.
func (g *LayoutGroup) Reset() { clear(g.states) }
