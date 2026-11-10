package el

import (
	"image"
	"image/color"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// Fly lifts a SharedLayout handoff above the other content of its root or
// overlay, escaping ancestor clips and opacity while preserving the element's
// own opacity and theme. Later overlays still cover it. At rest it returns to
// normal painting in its destination parent. It requires SharedLayout on the
// same element and does not retain the departing tree.
func (s *Styled[T]) Fly() *T {
	s.node().sharedFly = true
	return s.self
}

// A flight lives only for one paint pass. Keeping the node in the queue across
// frames would keep stale arena pointers and callbacks alive.
type sharedFlight struct {
	n                      *Node
	origin                 image.Point
	world                  f32.Affine2D
	transform              f32.Affine2D
	fade, exitFade         float32
	palette                theme.Palette
	textColor              optional[color.NRGBA]
	gtx                    core.C
	blockInput, blockFocus bool
	scrollParents          []*elemState
	dragScopes             []dragScope
}

func (e *engine) queueFlight(n *Node, origin image.Point, fade float32) {
	exitFade := e.flightExitFade
	if n.inert {
		opacity := e.sharedOpacity(n)
		exitFade = 1 - (1-exitFade)*(1-fade)*opacity
	}
	transform := e.poseWorld.Mul(f32.AffineId().Offset(f32.Pt(float32(origin.X), float32(origin.Y))))
	if s := e.sharedLayouts[n.sharedLayout.identity]; s != nil {
		// Use the unit-box transform directly. Conjugating through the destination
		// origin loses integer translation bits and can change native glyph sampling.
		transform = sharedBoxTransform(s.shown, n.size)
	}
	e.flights = append(e.flights, sharedFlight{
		n: n, origin: origin, world: e.poseWorld, transform: transform, fade: 1 - (1-fade)*(1-e.flightExitFade),
		exitFade: exitFade, palette: theme.Current(), textColor: e.paintTextColor, gtx: e.gtx,
		blockInput: e.blockInput || n.effectiveInert, blockFocus: e.blockFocus,
		scrollParents: append([]*elemState(nil), e.scrollParents...),
		dragScopes:    append([]dragScope(nil), e.dragScopes...),
	})
	e.markSharedLayout(n)
	e.frameSource.Execute(op.InvalidateCmd{})
	if r := paintingRoot; r != nil {
		r.motionDue = true
	}
}

func (e *engine) paintFlights(viewport image.Rectangle, areas *[]flightHit) {
	// Painting a lifted ancestor can discover more flights inside it. Process
	// those in this scope too, before a later overlay is allowed to draw.
	defer func() {
		clear(e.flights)
		e.flights = e.flights[:0]
	}()
	for i := 0; i < len(e.flights); i++ {
		e.paintFlight(e.flights[i], viewport, areas)
	}
}

func (e *engine) paintFlight(f sharedFlight, viewport image.Rectangle, areas *[]flightHit) {
	origin, world, visible := e.origin, e.poseWorld, e.visible
	exitFade := e.flightExitFade
	gtx, textColor, atlas := e.gtx, e.paintTextColor, e.textAtlasActive
	input, focus, horizontal := e.blockInput, e.blockFocus, e.horizontalView
	parents, scopes := e.scrollParents, e.dragScopes
	defer func() {
		e.origin, e.poseWorld, e.visible = origin, world, visible
		e.flightExitFade = exitFade
		e.gtx, e.paintTextColor, e.textAtlasActive = gtx, textColor, atlas
		e.blockInput, e.blockFocus, e.horizontalView = input, focus, horizontal
		e.scrollParents, e.dragScopes = parents, scopes
	}()
	e.origin, e.poseWorld = f.origin, f.world
	e.flightExitFade = f.exitFade
	e.visible = affineBounds(f.world.Invert(), viewport)
	e.gtx, e.paintTextColor, e.textAtlasActive = f.gtx, f.textColor, false
	e.blockInput, e.blockFocus, e.horizontalView = f.blockInput, f.blockFocus, nil
	e.scrollParents, e.dragScopes = f.scrollParents, f.dragScopes
	defer theme.Scope(f.palette)()
	n := f.n
	transform := f.transform
	if areas != nil && !f.blockInput && !n.effectiveDisabled {
		*areas = append(*areas, flightHit{transform.Invert(), n.size, e.cornersOf(e.sharedPaintStyle(n, n.style), n.size)})
	}
	defer clip.Rect(viewport).Push(e.gtx.Ops).Pop()
	if f.blockInput {
		core.Role("el-inert").Add(e.gtx.Ops)
	}
	defer op.Affine(transform).Push(e.gtx.Ops).Pop()
	if n.id != "" && e.anchors != nil {
		if _, found := e.anchors[n.id]; !found {
			e.anchors[n.id] = affineBounds(transform, image.Rectangle{Max: n.size})
		}
	}
	if f.fade > 0 {
		defer paint.PushOpacity(e.gtx.Ops, 1-f.fade).Pop()
	}
	if a := e.sharedOpacity(n); a < 1 {
		defer paint.PushOpacity(e.gtx.Ops, a).Pop()
	}
	if n.decorate != nil {
		g := e.gtx
		g.Constraints = layout.Exact(n.size)
		n.decorate(g, func() { e.paintContent(n) })
	} else {
		e.paintContent(n)
	}
}

// A flying child remains part of its overlay even outside the panel's box.
// Test in its local space, including rounded corners, rather than accepting
// the empty corners of a rotated axis-aligned bounding box.
type flightHit struct {
	inverse f32.Affine2D
	size    image.Point
	corners corners
}

func (a flightHit) contains(p f32.Point) bool {
	p = a.inverse.Transform(p)
	w, h := float32(a.size.X), float32(a.size.Y)
	if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
		return false
	}
	for i, radius := range a.corners {
		r := float32(radius)
		x, y := p.X, p.Y
		if i == 1 || i == 2 {
			x = w - x
		}
		if i >= 2 {
			y = h - y
		}
		if x < r && y < r && (x-r)*(x-r)+(y-r)*(y-r) > r*r {
			return false
		}
	}
	return true
}
