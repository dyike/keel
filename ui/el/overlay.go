package el

import (
	"image"
	"math"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// Side selects the edge of an anchor at which a layer is placed.
type Side uint8

const (
	Bottom Side = iota
	Top
	Left
	Right
)

// Layer describes an overlay for one Render. Declare it again while open.
type Layer struct {
	anchor                       string
	owner                        string
	content                      Element
	side                         Side
	align                        Align
	offset                       float32
	topInset                     float32
	arrow                        bool
	matchWidth                   bool
	modal, centered, trap, scrim bool
	edge                         bool // a Modal placed against a root edge
	keepOnOutside                bool
	keepOnEscape                 bool
	dismiss                      func()
	beforeDismiss                func() bool
	onEscape                     func() bool
	presenceFade                 float32
	exiting                      bool
}

func Anchored(anchorID string, content Element) *Layer {
	return &Layer{anchor: anchorID, content: content, offset: 4}
}
func Modal(content Element) *Layer {
	return &Layer{content: content, modal: true, centered: true, trap: true, scrim: true}
}

// Placement places an Anchored layer at side of its anchor. On a Modal it
// places the content against that edge of the root instead of centering it,
// e.g. a sheet: Modal(panel).Placement(Right, Start).
func (l *Layer) Placement(side Side, align Align) *Layer {
	l.side, l.align, l.edge = side, align, l.centered
	return l
}

// TopInset reserves space above an edge-positioned modal, in dp. The inset
// is capped at the root height. Other layer placements ignore it. The scrim
// still covers the root; painting and input stay below the inset, even in animation.
func (l *Layer) TopInset(dp float32) *Layer {
	if dp >= 0 && !math.IsNaN(float64(dp)) && !math.IsInf(float64(dp), 0) {
		l.topInset = dp
	}
	return l
}

// Owner ties a layer's lifetime to an enabled element in the current tree.
// Use this for modals declared by a component nested in a disabled or hidden
// container. The owner controls eligibility, not position or modality.
func (l *Layer) Owner(id string) *Layer { l.owner = id; return l }

func (l *Layer) Offset(dp float32) *Layer { l.offset = dp; return l }

// Arrow adds a 6dp pointer to an anchored layer. Offset measures to its tip.
// It follows the actual placement, including flips, and uses the panel background.
func (l *Layer) Arrow(on bool) *Layer { l.arrow = on; return l }

// MatchAnchorWidth sizes the layer to its anchor's width, as a dropdown
// matches its trigger.
func (l *Layer) MatchAnchorWidth() *Layer   { l.matchWidth = true; return l }
func (l *Layer) Modal() *Layer              { l.modal = true; return l }
func (l *Layer) TrapFocus() *Layer          { l.trap = true; return l }
func (l *Layer) OnDismiss(fn func()) *Layer { l.dismiss = fn; return l }

// BeforeDismiss may reject Esc/outside dismissal. Owner removal bypasses it.
func (l *Layer) BeforeDismiss(fn func() bool) *Layer { l.beforeDismiss = fn; return l }

// OnEscape handles an Escape press before normal dismissal. Returning true
// consumes the press and keeps the layer open; false permits normal dismissal.
// KeepOnEscape takes precedence. Outside presses do not call this callback.
func (l *Layer) OnEscape(fn func() bool) *Layer { l.onEscape = fn; return l }

// KeepOnOutsidePress stops presses outside the layer from calling OnDismiss;
// Esc still does. A destructive confirmation uses it so a stray click on the
// scrim cannot cancel it. A modal layer still blocks the press.
func (l *Layer) KeepOnOutsidePress() *Layer { l.keepOnOutside = true; return l }

// KeepOnEscape consumes Esc without dismissing this layer or layers below it.
func (l *Layer) KeepOnEscape() *Layer { l.keepOnEscape = true; return l }
func (l *Layer) Scrim(on bool) *Layer { l.scrim = on; return l }

// Presence applies an entry/exit sample to the content and scrim. An exiting
// layer immediately releases modality and returns focus. Its visual exit
// cannot handle dismissal or pointer input and is hidden from accessibility.
func (l *Layer) Presence(p PresenceFrame) *Layer {
	n := l.content.node()
	n.pose = &poseSpec{p.Pose, Instant}
	n.inert = p.Exiting
	l.presenceFade = p.Pose.Fade
	if !p.Visible {
		n.style.hidden = true
	}
	if p.Exiting {
		l.exiting, l.modal, l.trap = true, false, false
		l.keepOnEscape, l.keepOnOutside = true, true
		l.dismiss, l.beforeDismiss, l.onEscape = nil, nil, nil
	}
	return l
}

type overlayDecl struct {
	eligible bool
	key      any
	layer    *Layer
	state    *layerState
}
type layerState struct {
	seed              stateKey
	outside, surface  int
	layer             *Layer
	tree              *Node
	bounds, anchor    image.Rectangle
	arrowBounds       image.Rectangle
	returnFocus       event.Tag
	active, dismissed bool
	trapped           bool // focus ownership from the preceding painted frame
	flightAreas       []flightHit
	inputTransform    f32.Affine2D // visual root coordinates to the layout box
	clipBounds        image.Rectangle
}

// Overlay declares a layer in paint order. key must be comparable and unique
// within this root. Omission closes the layer; OnDismiss asks the owner to omit it.
// Declare outside Cache builders. Root supports full-window layers; Embed uses
// its maximum constraints and deferred painting as a best-effort fallback.
func (cx *Context) Overlay(key any, layer *Layer) {
	if layer == nil || layer.content == nil {
		return
	}
	for _, d := range cx.layers {
		if d.key == key {
			panic("el: duplicate overlay key")
		}
	}
	cx.layers = append(cx.layers, overlayDecl{key: key, layer: layer})
}

// Hovered reports the last processed pointer position for a visible ID.
// Disabled elements and elements behind a modal layer are never hovered.
func (cx *Context) Hovered(id string) bool {
	if id == "" {
		return false
	}
	cx.prepareInteractionQueries()
	for _, st := range cx.queryHover[id] {
		if st.hovered && !st.disabled && !st.blocked {
			return true
		}
	}
	return false
}

func (r *RootWidget) prepareLayers(cx *Context) {
	if len(cx.layers) == 0 {
		// Closing the last modal must still unblock its background controls.
		r.blockTree(r.mainTree, false)
		return
	}
	if r.layers == nil {
		r.layers = map[any]*layerState{}
	}
	ids := map[string]bool{}
	var collect func(*Node)
	collect = func(n *Node) {
		if n.style.hidden || n.effectiveDisabled || n.effectiveInert || n.disabled || n.inert {
			return
		}
		if n.id != "" {
			ids[n.id] = true
		}
		for _, c := range n.children {
			collect(c.node())
		}
	}
	collect(r.mainTree)
	serial := r.layerSerial
	for i := range cx.layers {
		d := &cx.layers[i]
		d.eligible = !d.layer.content.node().style.hidden && (d.layer.centered || ids[d.layer.anchor]) && (d.layer.owner == "" || ids[d.layer.owner])
		if d.eligible {
			collect(d.layer.content.node())
		}
		st := r.layers[d.key]
		if st == nil {
			serial++
			st = &layerState{seed: childKey(childKey(0, "overlay", 0), "", int(serial))}
			if r.e.gtx.Enabled() {
				r.layers[d.key] = st
				r.layerSerial = serial
			}
		}
		// Read-only painting may recompute geometry, but must not overwrite
		// the live layer's focus history or hit areas. Element state stays shared.
		if !r.e.gtx.Enabled() {
			copy := *st
			st = &copy
		}
		d.state = st
		r.store.assignKeys(d.layer.content.node(), st.seed)
		r.prepareKeys(d.layer.content.node(), nil, false)
	}
	// Event dispatch uses current declarations, not callbacks captured last frame.
	modal := -1
	for i, d := range cx.layers {
		if d.eligible && d.layer.modal {
			modal = i
		}
	}
	r.blockTree(r.mainTree, modal >= 0)
	for i, d := range cx.layers {
		r.blockTree(d.layer.content.node(), !d.eligible || i < modal)
	}
}
func (r *RootWidget) blockTree(n *Node, blocked bool) {
	blocked = blocked || n.effectiveInert
	if !r.e.gtx.Enabled() {
		return
	}
	if s := r.store.of(n); s != nil {
		s.blocked = blocked
		if blocked {
			// Blocked handlers ask for no events, so Gio drops their areas;
			// re-register them at once when unblocked, as for new elements,
			// or the first click after a modal closes is lost.
			s.hovered, s.fresh = false, true
		}
	}
	for _, c := range n.children {
		r.blockTree(c.node(), blocked)
	}
}
func (r *RootWidget) dismissLayer(st *layerState, l *Layer, forced bool) {
	if !r.e.gtx.Enabled() || st.dismissed {
		return
	}
	if !forced && l.beforeDismiss != nil {
		allowed := false
		core.Call(r.e.gtx, func() { r.callbacks = true; allowed = l.beforeDismiss() })
		if !allowed {
			return
		}
	}
	st.dismissed = true
	if l.dismiss != nil {
		core.Call(r.e.gtx, func() { r.callbacks = true; l.dismiss() })
	}
}
func (r *RootWidget) dispatchLayers(cx *Context) {
	if !r.e.gtx.Enabled() {
		return
	}
	// Escape goes to the topmost painted layer that handles dismissal, so a
	// notification stack above a dialog does not swallow it.
	for i := len(cx.layers) - 1; i >= 0; i-- {
		d := cx.layers[i]
		if !d.state.active || d.layer.exiting || (d.layer.dismiss == nil && d.layer.onEscape == nil && !d.layer.keepOnEscape) {
			continue
		}
		for {
			ev, ok := r.e.gtx.Event(key.Filter{Name: key.NameEscape})
			if !ok {
				break
			}
			if k, ok := ev.(key.Event); ok && k.State == key.Press {
				r.keyboardModality = true
			}
			if k, ok := ev.(key.Event); ok && k.State == key.Press && !d.layer.keepOnEscape {
				handled := false
				if d.layer.onEscape != nil {
					core.Call(r.e.gtx, func() { r.callbacks = true; handled = d.layer.onEscape() })
				}
				if !handled {
					r.dismissLayer(d.state, d.layer, false)
				}
			}
		}
		break
	}
	for i := len(cx.layers) - 1; i >= 0; i-- {
		d := cx.layers[i]
		st := d.state
		if !st.active || d.layer.exiting {
			continue
		}
		for {
			ev, ok := r.e.gtx.Event(pointer.Filter{Target: &st.outside, Kinds: pointer.Press})
			if !ok {
				break
			}
			r.keyboardModality = false
			p := ev.(pointer.Event).Position
			local := st.inputTransform.Transform(p)
			pos := image.Pt(int(math.Floor(float64(local.X))), int(math.Floor(float64(local.Y))))
			anchorPos := image.Pt(int(math.Floor(float64(p.X))), int(math.Floor(float64(p.Y))))
			outside := !anchorPos.In(st.clipBounds) || !pos.In(st.bounds) && !pos.In(st.arrowBounds)
			if anchorPos.In(st.clipBounds) {
				for _, area := range st.flightAreas {
					if area.contains(p) {
						outside = false
						break
					}
				}
			}
			if !d.layer.keepOnOutside && outside && (d.layer.modal || !anchorPos.In(st.anchor)) {
				r.dismissLayer(st, d.layer, false)
			}
		}
	}
}

// layerPosition flips only when the opposite side fits, then clamps both axes.
func layerPosition(anchor image.Rectangle, size, root image.Point, side Side, align Align, gap int) image.Point {
	pos, _ := layerPlacement(anchor, size, root, side, align, gap)
	return pos
}

func layerPlacement(anchor image.Rectangle, size, root image.Point, side Side, align Align, gap int) (image.Point, Side) {
	cross := func(lo, hi, length int) int {
		switch align {
		case Center:
			return lo + (hi-lo-length)/2
		case End:
			return hi - length
		default:
			return lo
		}
	}
	candidate := func(s Side) image.Point {
		switch s {
		case Top:
			return image.Pt(cross(anchor.Min.X, anchor.Max.X, size.X), anchor.Min.Y-gap-size.Y)
		case Left:
			return image.Pt(anchor.Min.X-gap-size.X, cross(anchor.Min.Y, anchor.Max.Y, size.Y))
		case Right:
			return image.Pt(anchor.Max.X+gap, cross(anchor.Min.Y, anchor.Max.Y, size.Y))
		default:
			return image.Pt(cross(anchor.Min.X, anchor.Max.X, size.X), anchor.Max.Y+gap)
		}
	}
	fits := func(p image.Point) bool {
		if side == Left || side == Right {
			return p.X >= 0 && p.X+size.X <= root.X
		}
		return p.Y >= 0 && p.Y+size.Y <= root.Y
	}
	p := candidate(side)
	opposite := map[Side]Side{Bottom: Top, Top: Bottom, Left: Right, Right: Left}[side]
	if other := candidate(opposite); !fits(p) && fits(other) {
		p = other
		side = opposite
	}
	return image.Pt(max(0, min(p.X, root.X-size.X)), max(0, min(p.Y, root.Y-size.Y))), side
}

func (r *RootWidget) paintLayers(cx *Context, base textStyle, priorFocus event.Tag) {
	gtx := r.e.gtx
	maxSize := gtx.Constraints.Max
	old := r.activeLayers
	var active []*layerState
	modal, trap := -1, -1
	for i, d := range cx.layers {
		if d.eligible && d.layer.modal {
			modal = i
		}
		if d.eligible && d.layer.trap {
			trap = i
		}
	}
	for i, d := range cx.layers {
		st, l := d.state, d.layer
		n := l.content.node()
		anchor, found := r.e.anchors[l.anchor]
		if !d.eligible || (!l.centered && !found) || n.style.hidden {
			r.dismissLayer(st, l, true)
			continue
		}
		if l.matchWidth && !l.centered {
			scale := r.e.m.PxPerDp
			if scale == 0 {
				scale = 1
			}
			// Exactly the anchor's width: content that stretches (a list, a
			// search field) would otherwise take the whole root.
			n.style.w = Dp(float32(anchor.Dx()) / scale)
		}
		available := maxSize
		inset := 0
		if l.edge {
			scale := r.e.m.PxPerDp
			if scale <= 0 {
				scale = 1
			}
			inset = min(maxSize.Y, r.e.dp(min(l.topInset, float32(maxSize.Y)/scale)))
			available.Y -= inset
		}
		r.e.layout(n, available.X, available.Y, base)
		r.e.place(n)
		actualSide := l.side
		if l.edge {
			n.pos = layerPosition(image.Rectangle{Max: available}, n.size, available, l.side, l.align, 0)
			// Inside the root rectangle: Bottom/Right mean against that edge.
			switch l.side {
			case Bottom:
				n.pos.Y = available.Y - n.size.Y
			case Top:
				n.pos.Y = 0
			case Left:
				n.pos.X = 0
			case Right:
				n.pos.X = maxSize.X - n.size.X
			}
			n.pos = image.Pt(max(0, n.pos.X), max(0, n.pos.Y)+inset)
		} else if l.centered {
			n.pos = image.Pt(max(0, (maxSize.X-n.size.X)/2), max(0, (maxSize.Y-n.size.Y)/2))
		} else {
			gap := r.e.dp(l.offset)
			if l.arrow {
				gap += r.e.dp(6)
			}
			n.pos, actualSide = layerPlacement(anchor, n.size, maxSize, l.side, l.align, gap)
		}
		reentering := st.layer != nil && st.layer.exiting && !l.exiting
		st.layer, st.tree, st.anchor = l, n, anchor
		st.bounds = image.Rectangle{Min: n.pos, Max: n.pos.Add(n.size)}
		st.arrowBounds = image.Rectangle{}
		var arrow *layerArrow
		if l.arrow && !l.centered {
			arrow = newLayerArrow(st.bounds, anchor, actualSide, l.align, r.e.dp(6), r.e.dp(n.style.radius))
			st.arrowBounds = arrow.bounds
		}
		visual, layerFade := f32.AffineId(), float32(0)
		if n.hasVisualMotion() {
			m, fade, _ := r.e.sampleVisualMotion(n)
			if n.sharedLayout != nil {
				m, _ = r.e.sampleSharedLayout(n, m, n.pos)
			}
			visual = f32.AffineId().Offset(f32.Pt(float32(n.pos.X), float32(n.pos.Y))).Mul(m).Mul(f32.AffineId().Offset(f32.Pt(-float32(n.pos.X), -float32(n.pos.Y))))
			layerFade = fade
		}
		st.inputTransform = visual.Invert()
		st.clipBounds = image.Rectangle{Min: image.Pt(0, inset), Max: maxSize}
		if !st.active || reentering || !st.trapped && l.trap {
			st.returnFocus = priorFocus
			st.dismissed = false
		}
		active = append(active, st)
		r.e.blockInput, r.e.blockFocus = i < modal, i < trap
		macro := op.Record(gtx.Ops)
		area := clip.Rect{Max: maxSize}.Push(gtx.Ops)
		if l.centered && l.scrim {
			scrim := theme.Scrim
			scrim.A = uint8(float32(scrim.A) * (1 - l.presenceFade))
			paint.Fill(gtx.Ops, scrim)
		}
		if !r.e.blockInput && !l.exiting {
			// Pass-through observes outside presses without consuming the target click.
			var pass pointer.PassStack
			if !l.modal {
				pass = pointer.PassOp{}.Push(gtx.Ops)
			}
			gtx.Event(pointer.Filter{Target: &st.outside, Kinds: pointer.Press})
			event.Op(gtx.Ops, &st.outside)
			if !l.modal {
				pass.Pop()
			}
			// Content is an opaque hit area, even between its interactive children.
			surfaceViewport := clip.Rect(st.clipBounds).Push(gtx.Ops)
			transformed := op.Affine(visual).Push(gtx.Ops)
			surface := clip.Rect(st.bounds.Union(st.arrowBounds)).Push(gtx.Ops)
			gtx.Event(pointer.Filter{Target: &st.surface, Kinds: pointer.Press | pointer.Release})
			event.Op(gtx.Ops, &st.surface)
			surface.Pop()
			transformed.Pop()
			surfaceViewport.Pop()
		}
		if i < modal {
			core.Role("el-inert").Add(gtx.Ops)
		}
		contentClip := clip.Rect{Min: image.Pt(0, inset), Max: maxSize}.Push(gtx.Ops)
		r.e.paint(n)
		st.flightAreas = st.flightAreas[:0]
		r.e.paintFlights(st.clipBounds, &st.flightAreas)
		contentClip.Pop()
		if arrow != nil {
			transformed := op.Affine(visual).Push(gtx.Ops)
			if layerFade > 0 {
				opacity := paint.PushOpacity(gtx.Ops, 1-layerFade)
				arrow.paint(gtx.Ops, n.style)
				opacity.Pop()
			} else {
				arrow.paint(gtx.Ops, n.style)
			}
			transformed.Pop()
		}
		area.Pop()
		call := macro.Stop()
		if r.fill {
			call.Add(gtx.Ops)
		} else {
			op.Defer(gtx.Ops, call)
		}
	}
	r.e.blockInput, r.e.blockFocus = false, false
	if !gtx.Enabled() {
		return
	}
	// Restore from innermost to outermost so closing an entire stack returns to
	// the original trigger, rather than to an element in another removed layer.
	var restore event.Tag
	closed := false
	for i := len(old) - 1; i >= 0; i-- {
		st := old[i]
		present := false
		for _, a := range active {
			if a == st {
				present = true
				break
			}
		}
		if st.trapped && (!present || !st.layer.trap) {
			restore, closed = st.returnFocus, true
		}
	}
	if closed {
		r.focusTag(restore)
	}
	for _, st := range active {
		if !st.trapped && st.layer.trap {
			r.focusFirst(st.tree)
		}
	}
	for _, st := range old {
		st.active = false
		st.trapped = false
	}
	for _, st := range active {
		st.active = true
		st.trapped = st.layer.trap
	}
	r.activeLayers = active
	used := map[any]bool{}
	for _, d := range cx.layers {
		used[d.key] = true
	}
	for k := range r.layers {
		if !used[k] {
			delete(r.layers, k)
		}
	}
	if len(active) > 0 {
		gtx.Event(key.Filter{Name: key.NameEscape})
	}
	if r.focusPending {
		// Restrict explicit focus requests to the innermost trapped layer.
		for i := len(active) - 1; i >= 0; i-- {
			if active[i].layer.trap {
				r.applyFocus(gtx, active[i].tree)
				return
			}
		}
		for i := len(active) - 1; i >= 0; i-- {
			if r.hasFocusID(active[i].tree, r.focusID) {
				r.applyFocus(gtx, active[i].tree)
				return
			}
		}
	}
	r.applyFocus(gtx, r.mainTree)
}
func (r *RootWidget) hasFocusID(n *Node, id string) bool {
	if n.id == id {
		return true
	}
	for _, c := range n.children {
		if r.hasFocusID(c.node(), id) {
			return true
		}
	}
	return false
}
func (r *RootWidget) focusedTag() event.Tag {
	for _, s := range r.store.states {
		if r.e.gtx.Focused(s) {
			return s
		}
		if s.editor != nil && r.e.gtx.Focused(s.editor) {
			return s.editor
		}
	}
	return nil
}
func (r *RootWidget) focusTag(tag event.Tag) {
	for _, s := range r.store.states {
		if s.keyFrame != r.store.frame || s.disabled || s.blocked {
			continue
		}
		if tag == s || tag == s.editor {
			// After a pointer close the trigger takes focus quietly, as if pressed.
			s.focusFrom(!r.keyboardModality)
			r.e.gtx.Execute(key.FocusCmd{Tag: tag})
			return
		}
	}
	r.e.gtx.Execute(key.FocusCmd{})
}
func (r *RootWidget) focusFirst(n *Node) bool {
	if targets, configured := r.collectTabTargets([]*Node{n}); configured {
		if len(targets) == 0 {
			return false
		}
		targets[0].state.focusFrom(false)
		r.e.gtx.Execute(key.FocusCmd{Tag: targets[0].tag})
		return true
	}
	return r.focusFirstDefault(n)
}

func (r *RootWidget) focusFirstDefault(n *Node) bool {
	if s := r.store.of(n); s != nil && s.keyFrame == r.store.frame && !s.disabled && !s.blocked {
		if n.input != nil {
			r.e.gtx.Execute(key.FocusCmd{Tag: s.editor})
			return true
		}
		if n.isFocusable() {
			r.e.gtx.Execute(key.FocusCmd{Tag: s})
			return true
		}
	}
	for _, c := range n.children {
		if r.focusFirstDefault(c.node()) {
			return true
		}
	}
	return false
}

func (r *RootWidget) dispatchHover(gtx core.C) {
	for _, st := range r.store.states {
		if st.disabled || st.blocked {
			st.hovered = false
			continue
		}
		for {
			kinds := pointer.Enter | pointer.Leave
			if st.onPointerMove != nil {
				kinds |= pointer.Move
			}
			ev, ok := gtx.Event(pointer.Filter{Target: &st.hoverTag, Kinds: kinds})
			if !ok {
				break
			}
			pe := ev.(pointer.Event)
			if st.onPointerMove != nil && (pe.Kind == pointer.Enter || pe.Kind == pointer.Move) {
				scale := gtx.Metric.PxPerDp
				if scale <= 0 {
					scale = 1
				}
				core.Call(gtx, func() { r.callbacks = true; st.onPointerMove(pe.Position.X/scale, pe.Position.Y/scale) })
			}
			hovered := pe.Kind != pointer.Leave
			if st.hovered != hovered {
				st.hovered = hovered
				r.callbacks = true
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
}
