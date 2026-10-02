package el

import (
	"gioui.org/io/key"
	"image"
	"strings"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
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
	n.input.line = line
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
	pos := n.pos
	if view := e.horizontalView; view != nil {
		if n.style.pinX < 0 {
			pos.X = view.Min.X + e.dp(n.style.pinOffset) - e.origin.X
		}
		if n.style.pinX > 0 {
			pos.X = view.Max.X - e.dp(n.style.pinOffset) - n.size.X - e.origin.X
		}
	}
	abs := image.Rectangle{Min: e.origin.Add(pos), Max: e.origin.Add(pos).Add(n.size)}
	if !n.style.absolute && !abs.Overlaps(e.visible) {
		return
	}
	if n.id != "" && e.anchors != nil {
		// A zero-size element anchors at its position: a notification
		// corner, the point a context menu opens at.
		if !abs.Intersect(e.visible).Empty() || abs.Empty() && abs.Min.In(e.visible) {
			if _, found := e.anchors[n.id]; !found {
				e.anchors[n.id] = abs
			}
		}
	}
	saved := e.origin
	e.origin = abs.Min
	defer func() { e.origin = saved }()
	defer op.Offset(pos).Push(gtx.Ops).Pop()
	if n.decorate != nil {
		g := gtx
		g.Constraints = layout.Exact(n.size)
		n.decorate(g, func() { e.paintContent(n) })
	} else {
		e.paintContent(n)
	}
}

func (e *engine) paintContent(n *Node) {
	savedGtx := e.gtx
	if n.effectiveDisabled || e.blockInput {
		e.gtx = e.gtx.Disabled()
	}
	defer func() { e.gtx = savedGtx }()
	gtx := e.gtx
	if n.style.revealSet {
		savedVisible := e.visible
		e.visible = e.visible.Intersect(image.Rectangle{Min: e.origin, Max: e.origin.Add(n.size)})
		defer func() { e.visible = savedVisible }()
		defer clip.Rect(image.Rectangle{Max: n.size}).Push(gtx.Ops).Pop()
	}
	st := n.style // a copy: hover and active variants change it for this frame only
	var state *elemState
	if n.id != "" || n.interactive() || (n.style.scrollY || n.style.scrollX) || n.input != nil {
		state = e.store.get(n.key)
	}
	if state != nil && gtx.Enabled() {
		state.enabledFrame = e.store.frame
	}
	if state != nil && n.interactive() && !n.effectiveDisabled && !e.blockInput {
		if n.hover != nil && state.click.Hovered() {
			n.hover(&st)
		}
		if n.active != nil && state.click.Pressed() {
			n.active(&st)
		}
	}
	if n.isFocusable() && n.input == nil && gtx.Focused(state) && !state.pointerFocus {
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
	if n.effectiveDisabled {
		// The disabled root picks the muted color, or its DisabledStyle
		// does; descendants keep that choice instead of resetting it.
		if n.disabledRoot || e.paintTextColor == nil {
			st.text.color = &theme.Muted
		} else {
			st.text.color = e.paintTextColor
		}
		if n.disabledStyle != nil {
			n.disabledStyle(&st)
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
	radius := min(e.dp(st.radius), min(n.size.X, n.size.Y)/2)

	// Elements that take input or report semantics get their own clip area;
	// others do not, so their children can overflow them.
	sem := e.semantics(n)
	if state != nil && n.id != "" && !n.effectiveDisabled && !e.blockInput {
		// Register above children so interactive descendants cannot hide hover.
		defer func() {
			hoverArea := clip.Rect(rect).Push(gtx.Ops)
			gtx.Event(pointer.Filter{Target: &state.hoverTag, Kinds: pointer.Enter | pointer.Leave})
			pass := pointer.PassOp{}.Push(gtx.Ops)
			event.Op(gtx.Ops, &state.hoverTag)
			pass.Pop()
			hoverArea.Pop()
		}()
	}
	if len(sem) > 0 || state != nil && (n.interactive() || (n.style.scrollY || n.style.scrollX) || n.input != nil) {
		defer clip.UniformRRect(rect, radius).Push(gtx.Ops).Pop()

		for _, o := range sem {
			o.Add(gtx.Ops)
		}
		if state != nil && n.isFocusable() && n.input == nil && !e.blockFocus && !e.blockInput {
			if gtx.Enabled() {
				state.keyFrame = e.store.frame
			}
			// Register in paint order so Gio Tab order matches tree order.
			gtx.Event(focusFilters(state)...)
			event.Op(gtx.Ops, state)
		}
		if state != nil && n.interactive() && !n.effectiveDisabled && !e.blockInput {
			if state.fresh && gtx.Enabled() {
				// Gio drops areas whose handler asked for no events this
				// frame. dispatch asks from the next frame on; ask now so
				// the element is clickable, and visible to agents, at once.
				state.click.Update(gtx.Source)
				if n.onDrag != nil {
					state.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
				}
				state.fresh = false
			}
			if n.onContextMenu != nil {
				// Register above interactive descendants, but pass primary events through.
				defer func() {
					area := clip.Rect(rect).Push(gtx.Ops)
					gtx.Event(pointer.Filter{Target: &state.contextTag, Kinds: pointer.Press | pointer.Release | pointer.Cancel})
					pass := pointer.PassOp{}.Push(gtx.Ops)
					event.Op(gtx.Ops, &state.contextTag)
					pass.Pop()
					area.Pop()
				}()
			}
			state.click.Add(gtx.Ops)
			if n.onDrag != nil {
				state.drag.Add(gtx.Ops)
			}
			if gtx.Enabled() {
				state.onClick, state.onDoubleClick, state.clickable = n.onClick, n.onDoubleClick, true
				state.onDrag, state.size = n.onDrag, n.size
			}
			if st.cursor != pointer.CursorDefault {
				st.cursor.Add(gtx.Ops)
			}
		}
	}

	// A press on the box that no child takes focuses its text: an input's own
	// padding, or a field frame's blank space (FocusOnPress). Registered
	// before the children, so the editor and any buttons stay on top.
	if state != nil && (n.focusOnPress != "" || n.input != nil) && !n.effectiveDisabled && !e.blockInput {
		area := clip.Rect(rect).Push(gtx.Ops)
		event.Op(gtx.Ops, &state.pressTag)
		pointer.CursorText.Add(gtx.Ops)
		area.Pop()
		if gtx.Enabled() {
			// Ask now as well: Gio drops an area nobody asked events for.
			gtx.Event(pointer.Filter{Target: &state.pressTag, Kinds: pointer.Press})
			state.pressFocus, state.pressEditor, state.pressable = n.focusOnPress, n.input != nil, true
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
		saved := e.gtx
		if e.blockFocus || n.effectiveDisabled {
			e.gtx = e.gtx.Disabled()
		}
		if gtx.Enabled() {
			state.keyFrame = e.store.frame
		}
		e.paintInput(n, state, inner)
		e.gtx = saved
	case n.widget != nil:
		stk := op.Offset(inner.Min).Push(gtx.Ops)
		g := gtx
		if e.blockFocus || n.effectiveDisabled {
			g = g.Disabled()
		}
		g.Constraints = layout.Exact(inner.Size())
		n.widget.Layout(g)
		stk.Pop()
	case st.scrollY || st.scrollX:
		e.paintScroll(n, state, inner)
	default:
		e.paintChildren(n)
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
		if n.value != "" {
			ops = append(ops, core.Role("button", n.value))
		}
	case "link", "tab", "columnheader", "select":
		ops = append(ops, semantic.Button, core.Role(role, n.value))
	case "checkbox":
		ops = append(ops, semantic.CheckBox)
		if n.value != "" {
			ops = append(ops, core.Role("checkbox", n.value)) // e.g. mixed
		}
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
		ops = append(ops, semantic.EnabledOp(!n.effectiveDisabled))
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
	ed.MaxLen, ed.Filter, ed.ReadOnly = spec.maxLen, spec.filter, spec.readOnly
	// A single-line box has no use for ↑ ↓ PageUp PageDown beyond jumping to
	// its ends, so OnKey takes them before the editor sees them.
	if n.onKey != nil && !spec.multiline && gtx.Enabled() {
		for {
			ev, ok := gtx.Event(
				key.Filter{Focus: ed, Name: key.NameUpArrow}, key.Filter{Focus: ed, Name: key.NameDownArrow},
				key.Filter{Focus: ed, Name: key.NamePageUp}, key.Filter{Focus: ed, Name: key.NamePageDown},
			)
			if !ok {
				break
			}
			ke, ok := ev.(key.Event)
			if !ok {
				continue
			}
			state := KeyPress
			if ke.State == key.Release {
				state = KeyRelease
			}
			fn, e := n.onKey, KeyEvent{Name: string(ke.Name), Modifiers: ke.Modifiers, State: state}
			core.Call(gtx, func() { fn(e) })
		}
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
	// A single-line box taller than its line (theme.ControlHeight) centers
	// the line, like a native field; the editor keeps the line's height.
	if !spec.multiline && spec.line > 0 && inner.Dy() > spec.line {
		inner.Min.Y += (inner.Dy() - spec.line) / 2
		inner.Max.Y = inner.Min.Y + spec.line
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
	if !gtx.Enabled() {
		copy := *st
		st = &copy
	}
	previousX, previousY := st.scrollX, st.scrollY
	bw := e.dp(n.style.borderWidth)
	pl, pt, pr, pb := e.edges(n.style.pad)
	viewport := image.Rect(bw, bw, n.size.X-bw, n.size.Y-bw)
	totalX := n.contentW + pl + pr
	maxX := max(totalX-viewport.Dx(), 0)
	if n.style.scrollX {
		dx := st.scrollHorizontal.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Horizontal,
			pointer.ScrollRange{Min: -st.scrollX, Max: maxX - st.scrollX}, pointer.ScrollRange{})
		st.scrollX = min(max(st.scrollX+dx+st.scrollPendingX, 0), maxX)
		st.scrollPendingX = 0
		st.scrollMaxX, st.scrollViewX, st.scrollContentX, st.scrolledX = maxX, viewport.Dx(), totalX, true
	} else {
		st.scrollX = 0
	}
	total := n.contentH + pt + pb
	maxScroll := max(total-viewport.Dy(), 0)
	if n.style.scrollY {
		// Stick: if the view was at the end last frame, it stays at the end.
		atEnd := !st.scrolled || st.scrollY >= st.scrollMax-e.dp(2)
		if n.style.stickBottom && atEnd || st.scrolled && n.style.endVersion != st.version {
			st.scrollY = maxScroll
		}
		st.version = n.style.endVersion
		if st.scrolled && n.style.keepVersion != st.keepVersion {
			st.scrollY = max(maxScroll-(st.scrollMax-st.scrollY), 0) // same distance from the bottom as last frame
		}
		st.keepVersion = n.style.keepVersion
		dist := st.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Vertical,
			pointer.ScrollRange{}, pointer.ScrollRange{Min: -st.scrollY, Max: maxScroll - st.scrollY})
		st.scrollY = min(max(st.scrollY+dist+st.scrollPending, 0), maxScroll)
		st.scrollPending = 0
		st.scrollMax, st.scrolled = maxScroll, true
		st.scrollView, st.scrollContent = viewport.Dy(), total
	} else {
		st.scrollY = 0
	}

	barWidth := e.dp(10)
	xTrack := image.Rect(viewport.Min.X, max(viewport.Min.Y, viewport.Max.Y-barWidth), viewport.Max.X, viewport.Max.Y)
	yTrack := image.Rect(max(viewport.Min.X, viewport.Max.X-barWidth), viewport.Min.Y, viewport.Max.X, viewport.Max.Y)
	if n.style.scrollX && maxX > 0 && n.style.scrollY && maxScroll > 0 {
		xTrack.Max.X = max(xTrack.Min.X, xTrack.Max.X-barWidth)
		yTrack.Max.Y = max(yTrack.Min.Y, yTrack.Max.Y-barWidth)
	}
	if n.style.scrollX {
		st.scrollX = st.scrollbarX.update(gtx, xTrack, true, st.scrollX, viewport.Dx(), totalX, e.dp(24), st)
	}
	if n.style.scrollY {
		st.scrollY = st.scrollbarY.update(gtx, yTrack, false, st.scrollY, viewport.Dy(), total, e.dp(24), st)
	}

	if gtx.Enabled() && (st.scrollX != previousX || st.scrollY != previousY) {
		// Virtual content is built before paint; rebuild at the new offset.
		gtx.Execute(op.InvalidateCmd{})
	}

	// The viewport is its own area with the scroll handler, so it is a node
	// in the semantic tree and agents only see what shows through it.
	stk := clip.Rect(viewport).Push(gtx.Ops)
	if n.style.scrollY {
		st.scroll.Add(gtx.Ops)
	}
	if n.style.scrollX {
		st.scrollHorizontal.Add(gtx.Ops)
	}
	off := op.Offset(image.Pt(-st.scrollX, -st.scrollY)).Push(gtx.Ops)
	savedVis, savedOrigin := e.visible, e.origin
	e.visible = e.visible.Intersect(viewport.Add(e.origin))
	e.origin = e.origin.Add(image.Pt(-st.scrollX, -st.scrollY))
	if n.style.scrollY {
		e.scrollParents = append(e.scrollParents, st)
	}
	savedHorizontal := e.horizontalView
	if n.style.scrollX {
		view := viewport.Add(savedOrigin)
		e.horizontalView = &view
	}
	e.paintChildren(n)
	e.horizontalView = savedHorizontal
	if n.style.scrollY {
		e.scrollParents = e.scrollParents[:len(e.scrollParents)-1]
	}
	e.visible, e.origin = savedVis, savedOrigin
	off.Pop()

	if n.style.scrollY {
		st.scrollbarY.paint(gtx, yTrack, false, st.scrollY, viewport.Dy(), total, e.dp(24))
	}
	if n.style.scrollX {
		st.scrollbarX.paint(gtx, xTrack, true, st.scrollX, viewport.Dx(), totalX, e.dp(24))
	}
	stk.Pop()
}

// Paint pinned siblings last and clip ordinary siblings to the remaining lane.
// Clipping affects both pixels and pointer/semantic operations.
func (e *engine) paintChildren(n *Node) {
	if e.horizontalView == nil {
		for _, c := range n.children {
			e.paint(c.node())
		}
		return
	}
	view := *e.horizontalView
	left, right := view.Min.X, view.Max.X
	pinned := false
	for _, c := range n.children {
		child := c.node()
		if child.style.hidden {
			continue
		}
		if child.style.pinX < 0 {
			left = max(left, view.Min.X+e.dp(child.style.pinOffset)+child.size.X)
			pinned = true
		}
		if child.style.pinX > 0 {
			right = min(right, view.Max.X-e.dp(child.style.pinOffset)-child.size.X)
			pinned = true
		}
	}
	if !pinned {
		for _, c := range n.children {
			e.paint(c.node())
		}
		return
	}
	draw := func(child *Node, minX, maxX int) {
		visible := e.visible
		area := image.Rect(minX, visible.Min.Y, max(minX, maxX), visible.Max.Y).Intersect(visible)
		if area.Empty() {
			return
		}
		e.visible = area
		stack := clip.Rect(area.Sub(e.origin)).Push(e.gtx.Ops)
		e.paint(child)
		stack.Pop()
		e.visible = visible
	}
	for _, c := range n.children {
		if c.node().style.pinX == 0 {
			draw(c.node(), left, right)
		}
	}
	for _, c := range n.children {
		child := c.node()
		if child.style.pinX < 0 {
			draw(child, view.Min.X, min(left, view.Max.X))
		}
		if child.style.pinX > 0 {
			draw(child, max(left, right), view.Max.X)
		}
	}
}
