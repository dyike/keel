package el

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

func (e *engine) label(n *Node, text string) material.LabelStyle {
	ts := n.textStyle
	lb := material.Label(theme.Material, ts.size, text)
	lb.Color = *ts.color
	if ts.bold != nil && *ts.bold {
		lb.Font.Weight = font.Bold
	}
	lb.MaxLines = ts.lines
	return lb
}

func (e *engine) measureText(n *Node, maxW int) image.Point {
	gtx := e.measureGtx(layout.Constraints{Max: image.Pt(min(maxW, inf), inf)})
	return e.label(n, n.text).Layout(gtx).Size
}

// measureInput: inputs fill the width they are given; their height is one
// line, or about three for a TextArea.
func (e *engine) measureInput(n *Node, maxW int) image.Point {
	line := e.measureText(n, inf).Y
	if line == 0 {
		line = e.label(n, "国").Layout(e.measureGtx(layout.Constraints{Max: image.Pt(inf, inf)})).Size.Y
	}
	w := maxW
	if w >= inf {
		w = e.dp(200)
	}
	h := line
	if n.input.multiline {
		h = max(h, e.dp(72))
	}
	return image.Pt(w, h)
}

// paint draws n and its descendants; offsets are relative to the parent.
func (e *engine) paint(n *Node) {
	if n.style.hidden {
		return
	}
	gtx := e.gtx
	// Skip what cannot be seen, e.g. the part of a long chat scrolled away.
	// Absolute children may overflow their parent, so they are always painted.
	abs := image.Rectangle{Min: e.origin.Add(n.pos), Max: e.origin.Add(n.pos).Add(n.size)}
	if !n.style.absolute && !abs.Overlaps(e.visible) {
		return
	}
	saved := e.origin
	e.origin = abs.Min
	defer func() { e.origin = saved }()
	defer op.Offset(n.pos).Push(gtx.Ops).Pop()
	if n.decorate != nil {
		g := gtx
		g.Constraints = layout.Exact(n.size)
		n.decorate(g, func() { e.paintContent(n) })
	} else {
		e.paintContent(n)
	}
}

func (e *engine) paintContent(n *Node) {
	gtx := e.gtx
	st := n.style // a copy: hover and active variants change it for this frame only
	var state *elemState
	if n.interactive() || n.style.scrollY || n.input != nil {
		state = e.store.get(n.key)
	}
	if state != nil && n.interactive() {
		if n.hover != nil && state.click.Hovered() {
			n.hover(&st)
		}
		if n.active != nil && state.click.Pressed() {
			n.active(&st)
		}
	}
	if n.focusable && n.input == nil && gtx.Focused(state) {
		st.borderWidth, st.borderColor = 2, theme.Primary
		if n.focus != nil {
			n.focus(&st)
		}
	}
	if n.input != nil && gtx.Focused(&state.editor) {
		st.borderColor = theme.Primary
		if n.focus != nil {
			n.focus(&st)
		}
	}
	// Visual text-color variants inherit without changing measured text metrics.
	savedColor, savedText := e.paintTextColor, n.textStyle
	if st.text.color != nil {
		e.paintTextColor = st.text.color
	}
	if e.paintTextColor != nil {
		n.textStyle.color = e.paintTextColor
	}
	defer func() { e.paintTextColor = savedColor; n.textStyle = savedText }()

	rect := image.Rectangle{Max: n.size}
	radius := e.dp(st.radius)

	// Elements that take input or report semantics get their own clip area;
	// others do not, so their children can overflow them.
	sem := e.semantics(n)
	if len(sem) > 0 || state != nil {
		defer clip.UniformRRect(rect, radius).Push(gtx.Ops).Pop()
		for _, o := range sem {
			o.Add(gtx.Ops)
		}
		if state != nil && n.focusable && n.input == nil {
			state.keyFrame = e.store.frame
			// Register in paint order so Gio Tab order matches tree order.
			gtx.Event(key.FocusFilter{Target: state}, key.Filter{Focus: state, Optional: allKeyModifiers})
			event.Op(gtx.Ops, state)
		}
		if state != nil && n.interactive() {
			if state.fresh {
				// Gio drops areas whose handler asked for no events this
				// frame. dispatch asks from the next frame on; ask now so
				// the element is clickable, and visible to agents, at once.
				state.click.Update(gtx.Source)
				state.fresh = false
			}
			state.click.Add(gtx.Ops)
			state.onClick, state.onDoubleClick, state.clickable = n.onClick, n.onDoubleClick, true
			if st.cursor != pointer.CursorDefault {
				st.cursor.Add(gtx.Ops)
			}
		}
	}

	e.paintBox(st, rect, radius)
	pl, pt, pr, pb := e.edges(st.pad)
	bw := e.dp(n.style.borderWidth) // visual focus/hover borders never move content
	inner := image.Rectangle{Min: image.Pt(bw+pl, bw+pt), Max: image.Pt(n.size.X-bw-pr, n.size.Y-bw-pb)}

	switch {
	case n.isText:
		e.paintText(n, inner)
	case n.input != nil:
		state.keyFrame = e.store.frame
		e.paintInput(n, state, inner)
	case n.widget != nil:
		stk := op.Offset(inner.Min).Push(gtx.Ops)
		g := gtx
		g.Constraints = layout.Exact(inner.Size())
		n.widget.Layout(g)
		stk.Pop()
	case st.scrollY:
		e.paintScroll(n, state, inner)
	default:
		for _, c := range n.children {
			e.paint(c.node())
		}
	}
}

// semantics returns the ops that describe n to agents.
func (e *engine) semantics(n *Node) []interface{ Add(*op.Ops) } {
	var ops []interface{ Add(*op.Ops) }
	role := n.role
	if role == "" && (n.onClick != nil || n.onDoubleClick != nil) {
		role = "button"
	}
	switch role {
	case "":
	case "button":
		ops = append(ops, semantic.Button)
	case "link", "tab", "columnheader", "select":
		ops = append(ops, semantic.Button, core.Role(role, n.value))
	case "checkbox":
		ops = append(ops, semantic.CheckBox)
	case "radio":
		ops = append(ops, semantic.RadioButton)
	case "switch":
		ops = append(ops, semantic.Switch)
	default:
		ops = append(ops, core.Role(role, n.value))
	}
	name := n.name
	if n.isText && name == "" {
		name = n.text
	}
	if n.input != nil {
		if name == "" {
			name = n.input.placeholder
		}
		ops = append(ops, semantic.Editor)
	}
	if name != "" {
		ops = append(ops, semantic.LabelOp(name))
	}
	if n.selected != nil {
		ops = append(ops, semantic.SelectedOp(*n.selected))
	}
	if len(ops) > 0 {
		ops = append(ops, semantic.EnabledOp(true))
	}
	return ops
}

func (e *engine) paintBox(st Style, rect image.Rectangle, radius int) {
	ops := e.gtx.Ops
	if st.bg != nil {
		paint.FillShape(ops, *st.bg, clip.UniformRRect(rect, radius).Op(ops))
	}
	if bw := e.dp(st.borderWidth); bw > 0 {
		// Stroke centered on a rectangle inset by half the width, so the
		// border stays inside the element.
		h := bw / 2
		r := image.Rectangle{Min: rect.Min.Add(image.Pt(h, h)), Max: rect.Max.Sub(image.Pt(bw-h, bw-h))}
		path := clip.UniformRRect(r, max(radius-h, 0)).Path(ops)
		paint.FillShape(ops, st.borderColor, clip.Stroke{Path: path, Width: float32(bw)}.Op())
	}
}

func (e *engine) paintText(n *Node, inner image.Rectangle) {
	gtx := e.gtx
	defer op.Offset(inner.Min.Add(image.Pt(0, theme.Shift(e.m, n.textStyle.size)))).Push(gtx.Ops).Pop()
	g := gtx
	g.Constraints = layout.Constraints{Max: inner.Size()}
	e.label(n, n.text).Layout(g)
}

func (e *engine) paintInput(n *Node, st *elemState, inner image.Rectangle) {
	gtx := e.gtx
	spec := n.input
	ed := &st.editor
	if !st.edInit {
		ed.SingleLine, ed.Submit = !spec.multiline, !spec.multiline
		st.edInit = true
	}
	if spec.password {
		ed.Mask = '•'
	} else {
		ed.Mask = 0
	}
	// User edits first, then program changes to the bound string.
	for {
		ev, ok := ed.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			text := ed.Text()
			if text == st.lastText {
				break // SetText below, not the user
			}
			st.lastText = text
			if spec.bind != nil {
				*spec.bind = text
			}
			if spec.onChange != nil {
				core.Call(gtx, func() { spec.onChange(text) })
			}
		case widget.SubmitEvent:
			if spec.onSubmit != nil {
				text := ed.Text()
				core.Call(gtx, func() { spec.onSubmit(text) })
			}
		}
	}
	if spec.bind != nil && *spec.bind != st.lastText {
		st.lastText = *spec.bind
		ed.SetText(*spec.bind)
	}
	defer op.Offset(inner.Min.Add(image.Pt(0, theme.Shift(e.m, n.textStyle.size)))).Push(gtx.Ops).Pop()
	g := gtx
	g.Constraints = layout.Exact(inner.Size())
	ts := n.textStyle
	me := material.Editor(theme.Material, ed, spec.placeholder)
	me.TextSize, me.Color, me.HintColor = ts.size, *ts.color, theme.Muted
	st.caret.Layout(g, me, theme.Material.Shaper)
	// Keep the value in the semantic tree for agents.
	value := ed.Text()
	if spec.password {
		value = strings.Repeat("•", ed.Len())
	}
	semantic.DescriptionOp(value).Add(gtx.Ops)
}

// paintScroll scrolls the padding box, like CSS: the padding scrolls with the
// content, so the last child can scroll clear of the bottom padding.
func (e *engine) paintScroll(n *Node, st *elemState, inner image.Rectangle) {
	gtx := e.gtx
	bw := e.dp(n.style.borderWidth)
	_, pt, _, pb := e.edges(n.style.pad)
	viewport := image.Rect(bw, bw, n.size.X-bw, n.size.Y-bw)
	total := n.contentH + pt + pb
	maxScroll := max(total-viewport.Dy(), 0)
	// Stick: if the view was at the end last frame, it stays at the end.
	atEnd := !st.scrolled || st.scrollY >= st.scrollMax-e.dp(2)
	if n.style.stickBottom && atEnd || st.scrolled && n.style.endVersion != st.version {
		st.scrollY = maxScroll
	}
	st.version = n.style.endVersion
	dist := st.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Vertical,
		pointer.ScrollRange{}, pointer.ScrollRange{Min: -st.scrollY, Max: maxScroll - st.scrollY})
	st.scrollY = min(max(st.scrollY+dist+st.scrollPending, 0), maxScroll)
	st.scrollPending = 0
	st.scrollMax, st.scrolled = maxScroll, true

	// The viewport is its own area with the scroll handler, so it is a node
	// in the semantic tree and agents only see what shows through it.
	stk := clip.Rect(viewport).Push(gtx.Ops)
	st.scroll.Add(gtx.Ops)
	off := op.Offset(image.Pt(0, -st.scrollY)).Push(gtx.Ops)
	savedVis, savedOrigin := e.visible, e.origin
	e.visible = e.visible.Intersect(viewport.Add(e.origin))
	e.origin = e.origin.Add(image.Pt(0, -st.scrollY))
	e.scrollParents = append(e.scrollParents, st)
	for _, c := range n.children {
		e.paint(c.node())
	}
	e.scrollParents = e.scrollParents[:len(e.scrollParents)-1]
	e.visible, e.origin = savedVis, savedOrigin
	off.Pop()
	stk.Pop()

	if maxScroll > 0 { // a thin indicator along the right edge
		view := viewport.Dy()
		thumb := max(view*view/total, e.dp(24))
		y := viewport.Min.Y + (view-thumb)*st.scrollY/maxScroll
		x := viewport.Max.X - e.dp(5)
		r := image.Rect(x, y, x+e.dp(3), y+thumb)
		paint.FillShape(gtx.Ops, color.NRGBA{A: 0x55}, clip.UniformRRect(r, e.dp(1.5)).Op(gtx.Ops))
	}
}
