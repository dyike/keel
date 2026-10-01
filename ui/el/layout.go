package el

import (
	"image"
	"image/color"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
)

// inf stands for an unbounded length, e.g. the height inside a scroll container.
const inf = 1 << 24

// engine lays out and paints one element tree for one frame.
type engine struct {
	anchors                map[string]image.Rectangle
	blockInput, blockFocus bool
	paintTextColor         *color.NRGBA
	gtx                    core.C
	m                      unit.Metric
	store                  *store
	scrollParents          []*elemState
	scratch                op.Ops          // measuring passes record here and discard
	origin                 image.Point     // absolute position of the node being painted's parent
	visible                image.Rectangle // absolute area that can show anything
}

func (e *engine) dp(v float32) int { return e.m.Dp(unit.Dp(v)) }

func (e *engine) edges(ed Edges) (l, t, r, b int) {
	return e.dp(ed.Left), e.dp(ed.Top), e.dp(ed.Right), e.dp(ed.Bottom)
}

// layout sizes n, given its parent's content box (inf when unbounded), and
// sizes its descendants. Positions are set afterwards by place.
func (e *engine) layout(n *Node, availW, availH int, parent textStyle) {
	n.reused = false
	if n.cached {
		in := [4]int{availW, availH, n.forceW, n.forceH}
		if m := &n.memo; m.ok && m.in == in && m.parent == parent {
			n.size, n.reused = m.size, true
			return
		}
		defer func() { n.memo = layoutMemo{true, in, parent, n.size} }()
	}
	s := &n.style
	n.textStyle = s.text.inherit(parent)
	w, h := s.w.px(e.m, availW), s.h.px(e.m, availH)
	if n.forceW >= 0 {
		w = n.forceW
	}
	if n.forceH >= 0 {
		h = n.forceH
	}
	pl, pt, pr, pb := e.edges(s.pad)
	bw := e.dp(s.borderWidth)
	boxX, boxY := pl+pr+2*bw, pt+pb+2*bw
	w, h = e.clampW(n, w, availW), e.clampH(n, h, availH)

	innerW, innerH := -1, -1
	if w >= 0 {
		innerW = max(w-boxX, 0)
	}
	if h >= 0 {
		innerH = max(h-boxY, 0)
	}
	limW, limH := innerW, innerH
	if limW < 0 {
		limW = shrinkBy(availW, boxX)
	}
	if limH < 0 {
		limH = shrinkBy(availH, boxY)
	}

	var content image.Point
	switch {
	case n.isText:
		content = e.measureText(n, limW)
	case n.input != nil:
		content = e.measureInput(n, limW)
	case n.widget != nil:
		content = e.measureWidget(n, innerW, innerH, limW, limH)
	case s.scrollY:
		// Children get unbounded height; the element keeps its own.
		content = e.flex(n, innerW, -1, limW, inf)
		n.contentH = content.Y
	default:
		content = e.flex(n, innerW, innerH, limW, limH)
	}
	if w < 0 {
		w = e.clampW(n, content.X+boxX, availW)
	}
	if h < 0 {
		h = e.clampH(n, content.Y+boxY, availH)
	}
	n.size = image.Pt(w, h)
}

func shrinkBy(avail, by int) int {
	if avail >= inf {
		return inf
	}
	return max(avail-by, 0)
}

func (e *engine) clampW(n *Node, w, avail int) int {
	if w < 0 {
		return w
	}
	if mn := n.style.minW.px(e.m, avail); mn >= 0 {
		w = max(w, mn)
	}
	if mx := n.style.maxW.px(e.m, avail); mx >= 0 {
		w = min(w, mx)
	}
	return w
}

func (e *engine) clampH(n *Node, h, avail int) int {
	if h < 0 {
		return h
	}
	if mn := n.style.minH.px(e.m, avail); mn >= 0 {
		h = max(h, mn)
	}
	if mx := n.style.maxH.px(e.m, avail); mx >= 0 {
		h = min(h, mx)
	}
	return h
}

// inFlow lists the children that take part in flex layout.
func inFlow(n *Node) []*Node {
	var out []*Node
	for _, c := range n.children {
		cn := c.node()
		if !cn.style.hidden && !cn.style.absolute {
			out = append(out, cn)
		}
	}
	return out
}

// axis helpers: main/cross components of a point for a row or column.
func mainOf(p image.Point, row bool) int {
	if row {
		return p.X
	}
	return p.Y
}

func crossOf(p image.Point, row bool) int {
	if row {
		return p.Y
	}
	return p.X
}

func (e *engine) margins(n *Node, row bool) (mainStart, mainEnd, crossStart, crossEnd int) {
	l, t, r, b := e.edges(n.style.margin)
	if row {
		return l, r, t, b
	}
	return t, b, l, r
}

// childAvail gives a child the container's content box, less its margins.
func (e *engine) childAvail(c *Node, limW, limH int) (int, int) {
	l, t, r, b := e.edges(c.style.margin)
	return shrinkBy(limW, l+r), shrinkBy(limH, t+b)
}

func (e *engine) layoutChild(c *Node, parent *Node, limW, limH int) {
	aw, ah := e.childAvail(c, limW, limH)
	e.layout(c, aw, ah, parent.textStyle)
}

func (n *Node) crossAuto(row bool) bool {
	if row {
		return n.style.h.kind == autoLen
	}
	return n.style.w.kind == autoLen
}

func (n *Node) setForce(main, cross int, row bool) {
	if row {
		n.forceW, n.forceH = main, cross
	} else {
		n.forceW, n.forceH = cross, main
	}
}

func (n *Node) force(row bool) (main, cross int) {
	if row {
		return n.forceW, n.forceH
	}
	return n.forceH, n.forceW
}

// flex sizes the children of a container and returns its content size.
// innerW/innerH are -1 when the container's own size is automatic.
func (e *engine) flex(n *Node, innerW, innerH, limW, limH int) image.Point {
	s := &n.style
	row := s.row
	kids := inFlow(n)
	gap := e.dp(s.gap)
	align := s.align
	if !s.alignSet {
		align = Stretch
	}
	mainDef, crossDef := innerH, innerW
	if row {
		mainDef, crossDef = innerW, innerH
	}

	outer := func(c *Node) (int, int) {
		ms, me, cs, ce := e.margins(c, row)
		return mainOf(c.size, row) + ms + me, crossOf(c.size, row) + cs + ce
	}
	// Stretch up front when the cross size is known, so a child is laid out
	// once, not measured and then laid out again at the stretched size.
	preCross := func(c *Node) int {
		if align == Stretch && crossDef >= 0 && c.crossAuto(row) {
			_, _, cs, ce := e.margins(c, row)
			return max(crossDef-cs-ce, 0)
		}
		return -1
	}
	// Grow means flex: 1 — growing children start from zero and share the free
	// space, so each is laid out once, at its final size. That needs a known
	// main size; without one, growing is meaningless and all are natural.
	growing := func(c *Node) bool { return mainDef >= 0 && c.style.grow > 0 }

	total := 0
	var grow float32
	for i, c := range kids {
		if i > 0 {
			total += gap
		}
		if growing(c) {
			grow += c.style.grow
			ms, me, _, _ := e.margins(c, row)
			total += ms + me
			continue
		}
		c.setForce(-1, preCross(c), row)
		e.layoutChild(c, n, limW, limH)
		m, _ := outer(c)
		total += m
	}
	if mainDef >= 0 {
		free := mainDef - total
		if grow > 0 {
			left := max(free, 0)
			for _, c := range kids {
				if !growing(c) {
					continue
				}
				share := int(float32(max(free, 0)) * c.style.grow / grow)
				left -= share
				c.setForce(share, preCross(c), row)
				e.layoutChild(c, n, limW, limH)
			}
			total = mainDef - left
		} else if free < 0 {
			var weight float32
			for _, c := range kids {
				if c.style.shrink >= 0 {
					weight += float32(mainOf(c.size, row))
				}
			}
			if weight > 0 {
				for _, c := range kids {
					if c.style.shrink < 0 {
						continue
					}
					cut := int(float32(-free) * float32(mainOf(c.size, row)) / weight)
					_, cross := c.force(row)
					c.setForce(max(mainOf(c.size, row)-cut, 0), cross, row)
					e.layoutChild(c, n, limW, limH)
				}
				total = mainDef
			}
		}
	}

	// Cross size, then stretch children that have no cross size of their own.
	crossSize := 0
	for _, c := range kids {
		_, cr := outer(c)
		crossSize = max(crossSize, cr)
	}
	if crossDef >= 0 {
		crossSize = crossDef
	}
	if align == Stretch {
		for _, c := range kids {
			if !c.crossAuto(row) {
				continue
			}
			_, _, cs, ce := e.margins(c, row)
			want := max(crossSize-cs-ce, 0)
			if crossOf(c.size, row) != want {
				main, _ := c.force(row)
				c.setForce(main, want, row)
				e.layoutChild(c, n, limW, limH)
			}
		}
	}
	if mainDef >= 0 {
		total = max(total, 0)
	}
	if row {
		return image.Pt(total, crossSize)
	}
	return image.Pt(crossSize, total)
}

// place positions the children of n inside its final size, recursively.
func (e *engine) place(n *Node) {
	if n.reused {
		return // children keep the positions of the frame that laid them out
	}
	s := &n.style
	row := s.row
	pl, pt, pr, pb := e.edges(s.pad)
	bw := e.dp(s.borderWidth)
	origin := image.Pt(bw+pl, bw+pt)
	inner := image.Pt(max(n.size.X-pl-pr-2*bw, 0), max(n.size.Y-pt-pb-2*bw, 0))
	if s.scrollY {
		inner.Y = n.contentH
	}
	kids := inFlow(n)
	gap := e.dp(s.gap)
	align := s.align
	if !s.alignSet {
		align = Stretch
	}

	total := 0
	for i, c := range kids {
		ms, me, _, _ := e.margins(c, row)
		total += mainOf(c.size, row) + ms + me
		if i > 0 {
			total += gap
		}
	}
	free := max(mainOf(inner, row)-total, 0)
	cursor, extra := 0, 0
	switch s.justify {
	case Center:
		cursor = free / 2
	case End:
		cursor = free
	case SpaceBetween:
		if len(kids) > 1 {
			extra = free / (len(kids) - 1)
		}
	case SpaceAround:
		if len(kids) > 0 {
			extra = free / len(kids)
			cursor = extra / 2
		}
	}
	crossInner := crossOf(inner, row)
	for _, c := range kids {
		ms, me, cs, ce := e.margins(c, row)
		cross := cs
		switch align {
		case Center:
			cross = cs + (crossInner-crossOf(c.size, row)-cs-ce)/2
		case End:
			cross = crossInner - crossOf(c.size, row) - ce
		}
		main := cursor + ms
		if row {
			c.pos = origin.Add(image.Pt(main, cross))
		} else {
			c.pos = origin.Add(image.Pt(cross, main))
		}
		cursor = main + mainOf(c.size, row) + me + gap + extra
	}

	// Absolute children: sized against the padding box and pinned to its edges.
	padBox := image.Pt(n.size.X-2*bw, n.size.Y-2*bw)
	for _, cn := range n.children {
		c := cn.node()
		if c.style.hidden || !c.style.absolute {
			continue
		}
		c.forceW, c.forceH = -1, -1
		// Pinned to opposite edges with no size of its own: stretch between them.
		if c.style.left != nil && c.style.right != nil && c.style.w.kind == autoLen {
			c.forceW = max(padBox.X-e.dp(*c.style.left)-e.dp(*c.style.right), 0)
		}
		if c.style.top != nil && c.style.bottom != nil && c.style.h.kind == autoLen {
			c.forceH = max(padBox.Y-e.dp(*c.style.top)-e.dp(*c.style.bottom), 0)
		}
		e.layout(c, padBox.X, padBox.Y, n.textStyle)
		x, y := bw, bw
		switch {
		case c.style.left != nil:
			x += e.dp(*c.style.left)
		case c.style.right != nil:
			x = n.size.X - bw - e.dp(*c.style.right) - c.size.X
		}
		switch {
		case c.style.top != nil:
			y += e.dp(*c.style.top)
		case c.style.bottom != nil:
			y = n.size.Y - bw - e.dp(*c.style.bottom) - c.size.Y
		}
		c.pos = image.Pt(x, y)
	}
	for _, c := range n.children {
		if cn := c.node(); !cn.style.hidden {
			e.place(cn)
		}
	}
}

// measureGtx is a context for measuring: its ops are thrown away and its input
// source is disabled, so widgets neither paint nor consume events.
func (e *engine) measureGtx(cs layout.Constraints) core.C {
	e.scratch.Reset()
	g := e.gtx
	g.Ops = &e.scratch
	g.Source = input.Source{}
	g.Constraints = cs
	return g
}

func (e *engine) measureWidget(n *Node, innerW, innerH, limW, limH int) image.Point {
	cs := layout.Constraints{Max: image.Pt(min(limW, inf), min(limH, inf))}
	if innerW >= 0 {
		cs.Min.X, cs.Max.X = innerW, innerW
	}
	if innerH >= 0 {
		cs.Min.Y, cs.Max.Y = innerH, innerH
	}
	return n.widget.Layout(e.measureGtx(cs)).Size
}
