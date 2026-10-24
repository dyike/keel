package el

import (
	"image"
	"time"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
)

// Paint-only frames. A frame that only advances paint-time animation (see
// AmbientPhase) repaints the element tree and layout of the previous frame
// instead of rendering and laying them out again. Rendering, binding and
// layout are about a third of such a frame.
//
// Only a window's own root repaints (ui/window calls RetainFrames), and
// only at the tick AmbientPhase requested: frames anything else requests
// sooner render in full, and so does a test or embedding that drives a root
// directly, changing its views between frames. Within that, a frame
// repaints only when nothing it would render can differ:
//   - no input is waiting, and none was waiting when the previous frame
//     began (so a frame that dispatched input is always followed by a full
//     one, which renders the state that input left);
//   - nothing changed outside a frame since (core.Update, core.Call and
//     every other loop.InvalidateAll, and window state);
//   - the previous Render did not read the frame time or request animation
//     (Context.Now, Animating, AnimatingAmbient), and no timer is waiting;
//   - the root is drawn with the same constraints, metric and viewport.

// retainedFrame is what the previous full frame rendered and laid out.
type retainedFrame struct {
	tree       *Node
	cx         Context
	max        image.Point
	metric     unit.Metric
	visible    image.Rectangle
	origin     image.Point
	parents    []*elemState
	generation uint64
	// quiet is whether no input was waiting when that frame began.
	quiet bool
}

// RetainFrames lets this root repaint frames that only advance paint-time
// animation. ui/window calls it for a window's content; views of such a
// root change only through callbacks and core.Update, which invalidate it.
func (r *RootWidget) RetainFrames() { r.retainFrames = true }

// paintingRoot is the root whose tree is being painted, for AmbientPhase.
// Painting holds the frame lock.
var paintingRoot *RootWidget

// noteRenderTime records that the frame being rendered depends on time.
func (cx *Context) noteRenderTime() { cx.root.renderTime = true }

// AmbientPhase returns the phase, from 0 to 1, of a looping indicator with
// the given period at the frame time, and requests the next frame at the
// AmbientFrameInterval cadence, as Context.AnimatingAmbient does. Call it
// while painting, from a Widget's Layout, rather than in Render: the
// element tree then stays the same between ambient frames, which repaint
// without rendering or laying it out again. It returns 0 under reduced
// motion.
func AmbientPhase(gtx core.C, period time.Duration) float32 {
	if ReducedMotion() || period <= 0 {
		return 0
	}
	now := gtx.Now.UnixNano()
	if gtx.Enabled() {
		step := int64(AmbientFrameInterval)
		at := time.Unix(0, now-now%step+step)
		gtx.Execute(op.InvalidateCmd{At: at})
		if r := paintingRoot; r != nil && (r.ambientDue.IsZero() || at.Before(r.ambientDue)) {
			r.ambientDue = at
		}
	}
	return float32(now%int64(period)) / float32(period)
}

// canRepaint reports whether this frame may repaint the retained one.
func (r *RootWidget) canRepaint(gtx core.C, visible image.Rectangle, origin image.Point, parents []*elemState) bool {
	// The tick the previous frame's AmbientPhase calls requested.
	due := r.ambientDue
	r.ambientDue = time.Time{}
	p := &r.retained
	if !r.retainFrames || p.tree == nil || !gtx.Enabled() || due.IsZero() || gtx.Now.Before(due) {
		return false
	}
	if !p.quiet || gtx.Source.Pending() {
		return false
	}
	if p.generation != loop.Generation() || p.max != gtx.Constraints.Max || p.metric != gtx.Metric {
		return false
	}
	if p.visible != visible || p.origin != origin || len(p.parents) != len(parents) {
		return false
	}
	for i := range parents {
		if p.parents[i] != parents[i] {
			return false
		}
	}
	if r.requestedFocus != nil || r.focusPending {
		return false
	}
	for _, t := range r.timers {
		if !t.fired {
			return false
		}
	}
	return true
}

// retain keeps a full frame's tree and layout for paint-only frames, or
// drops them when that frame depended on time.
func (r *RootWidget) retain(gtx core.C, tree *Node, cx *Context, visible image.Rectangle, origin image.Point, parents []*elemState, quiet bool) {
	if r.renderTime || !gtx.Enabled() {
		r.retained = retainedFrame{}
		return
	}
	r.retained = retainedFrame{
		tree:       tree,
		cx:         *cx,
		max:        gtx.Constraints.Max,
		metric:     gtx.Metric,
		visible:    visible,
		origin:     origin,
		parents:    append(r.retained.parents[:0], parents...),
		generation: loop.Generation(),
		quiet:      quiet,
	}
}

// repaint paints the retained tree; the counterpart of layoutViewport's
// paint phase. Element state is neither advanced nor swept: the tree is the
// previous frame's.
func (r *RootWidget) repaint(gtx core.C, visible image.Rectangle) core.D {
	e := &r.e
	st := r.store
	st.beginStateStats()
	defer st.endStateStats()
	theme.BeginTextFrame()
	e.gtx, e.m, e.store = gtx, gtx.Metric, st
	r.source = gtx.Source
	p := &r.retained
	tree, cx := p.tree, &p.cx
	base := textStyle{size: theme.BodySize}
	max := gtx.Constraints.Max
	cx.resetInteractionQueries()
	e.origin, e.visible = image.Point{}, visible.Intersect(image.Rectangle{Max: max})
	e.prepareTextAtlas(tree)
	area := clip.Rect{Max: max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &r.bg)
	e.beginAnchors()
	e.blockInput, e.blockFocus = false, false
	for _, d := range cx.layers {
		e.blockInput = e.blockInput || (d.eligible && d.layer.modal)
		e.blockFocus = e.blockFocus || (d.eligible && d.layer.trap)
	}
	if e.blockInput {
		core.Role("el-inert").Add(gtx.Ops)
	}
	prior := paintingRoot
	paintingRoot = r
	defer func() { paintingRoot = prior }()
	e.paint(tree)
	e.textAtlasActive = false
	area.Pop()
	e.blockInput, e.blockFocus = false, false
	r.paintLayers(cx, base, r.focusedTag())
	r.repaints++
	if r.fill {
		return core.D{Size: max}
	}
	return core.D{Size: tree.size}
}
