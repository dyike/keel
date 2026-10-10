package el

import (
	"image"
	"image/color"
	"math"
	"strings"
	"time"
	"unicode"

	"gioui.org/f32"
	"gioui.org/io/key"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/inputcontent"
	"github.com/dyike/keel/ui/theme"
	"golang.org/x/image/math/fixed"
)

func (e *engine) label(n *Node, value string) material.LabelStyle {
	ts := n.textStyle
	lb := material.Label(theme.Material, ts.size, value)
	lb.Color = *ts.color
	lb.Font = textFont(ts)
	if ts.align != nil {
		switch *ts.align {
		case Center:
			lb.Alignment = text.Middle
		case End:
			lb.Alignment = text.End
		}
	}
	if ts.lineHeight > 0 {
		lb.LineHeightScale = ts.lineHeight
	}
	lb.MaxLines = ts.lines
	return lb
}

// textFont is the font a text style draws with.
func textFont(ts textStyle) font.Font {
	f := font.Font{Typeface: theme.Material.Face}
	if ts.mono != nil && *ts.mono {
		f.Typeface = theme.MonoFace
	}
	if ts.weight != nil {
		f.Weight = *ts.weight
	}
	if ts.italic != nil && *ts.italic {
		f.Style = font.Italic
	}
	return f
}

func (e *engine) measureText(n *Node, maxW int) image.Point {
	gtx := e.measureGtx(layout.Constraints{Max: image.Pt(min(maxW, inf), inf)})
	return e.measureLabel(gtx, e.label(n, n.text)).Size
}

// measureInput: inputs fill the width they are given; their height is one
// line, or about three for a TextArea unless AutoGrow is configured.
func (e *engine) measureInput(n *Node, maxW int) image.Point {
	// Empty text has a shorter line box and clips Latin descenders. Use the
	// same stable mixed-script metrics as textShift, independent of input contents.
	line := e.measureLabel(e.measureGtx(layout.Constraints{Max: image.Pt(inf, inf)}), e.label(n, "国Ag")).Size.Y
	w := maxW
	if w >= inf {
		w = e.dp(200)
	}
	var objectState *elemState
	if n.input.document != nil {
		objectState = e.store.get(n.key)
		e.prepareInputObjects(n, objectState, w)
		for _, dims := range objectState.inputObjects.sizes {
			line = max(line, dims.Size.Y)
		}
	}
	n.input.line = line
	h := line
	if n.input.multiline {
		h = max(h, e.dp(72))
		if spec := n.input; spec.minRows > 0 {
			value := ""
			if spec.document != nil {
				value = spec.document.presentation().Text
			} else if spec.bind != nil {
				value = *spec.bind
			} else if st := e.store.states[n.key]; st != nil {
				value = st.lastText
			}
			if spec.password {
				value = strings.Map(func(r rune) rune {
					if r == '\n' {
						return r
					}
					return '•'
				}, value)
			}
			// Match the font and line height passed to material.Editor at paint time.
			lb := e.label(n, value)
			if objectState != nil && objectState.inputObjects.shaper != nil {
				th := *theme.Material
				th.Shaper = objectState.inputObjects.shaper
				lb = material.Label(&th, n.textStyle.size, value)
				lb.Font = objectState.inputObjects.font
				lb.LineHeightScale = n.textStyle.lineHeight
			}
			// One extra row: overflowing text then clamps to exactly maxRows
			// rows, whatever the truncator's own line box would add.
			lb.MaxLines = spec.maxRows + 1
			measured := e.measureLabel(e.measureGtx(layout.Constraints{Max: image.Pt(w, inf)}), lb).Size.Y
			// Measure baseline spacing rather than multiplying glyph bounds:
			// the first line and subsequent line advances need not match.
			one := e.measureLabel(e.measureGtx(layout.Constraints{Max: image.Pt(inf, inf)}), e.label(n, "M")).Size.Y
			two := e.measureLabel(e.measureGtx(layout.Constraints{Max: image.Pt(inf, inf)}), e.label(n, "M\nM")).Size.Y
			advance := max(two-one, 1)
			rowHeight := func(rows int) int { return one + min(rows-1, (inf-one)/advance)*advance }
			minH, maxH := rowHeight(spec.minRows), rowHeight(spec.maxRows)
			h = min(max(measured, minH), maxH)
		}
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
	pos = pos.Add(image.Pt(e.dp(n.style.translate[0]), e.dp(n.style.translate[1])))
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
	if a := n.style.opacity; n.style.opacitySet && a < 1 {
		defer paint.PushOpacity(gtx.Ops, a).Pop()
	}
	if n.decorate != nil {
		g := gtx
		g.Constraints = layout.Exact(n.size)
		n.decorate(g, func() { e.paintContent(n) })
	} else {
		e.paintContent(n)
	}
}

func (e *engine) paintContent(n *Node) {
	if n.palette != nil {
		defer theme.Scope(*n.palette)()
	}
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
	if n.id != "" || n.interactive() || n.onScroll != nil || (n.style.scrollY || n.style.scrollX) || n.input != nil {
		state = e.store.get(n.key)
	}
	if gtx.Enabled() {
		if n.onClick != nil || n.onDoubleClick != nil || n.onDrag != nil || n.input != nil || n.widget != nil || n.style.scrollX || n.style.scrollY {
			bounds := image.Rectangle{Min: e.origin, Max: e.origin.Add(n.size)}.Intersect(e.visible)
			for _, scope := range e.dragScopes {
				scope.drag.blocked = append(scope.drag.blocked, bounds.Sub(scope.origin))
			}
		}
		if state != nil && n.onDrag != nil && n.dragAccept != nil {
			state.conditionalDrag.blocked = state.conditionalDrag.blocked[:0]
			e.dragScopes = append(e.dragScopes, dragScope{&state.conditionalDrag, e.origin})
			defer func() { e.dragScopes = e.dragScopes[:len(e.dragScopes)-1] }()
		}
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
		if (n.hover != nil || n.active != nil) && gtx.Enabled() {
			e.easeBackground(&st, state)
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
	if theme.Frameless {
		st.borderWidth = 0
		if !n.effectiveDisabled && ((n.isFocusable() && n.input == nil && gtx.Focused(state) && !state.pointerFocus) || (n.input != nil && gtx.Focused(&state.editor))) {
			bg := theme.Highlight
			if st.bg != nil && st.bg.A != 0 {
				bg = *st.bg
				bg.R = uint8((4*uint16(bg.R) + uint16(theme.Muted.R)) / 5)
				bg.G = uint8((4*uint16(bg.G) + uint16(theme.Muted.G)) / 5)
				bg.B = uint8((4*uint16(bg.B) + uint16(theme.Muted.B)) / 5)
			}
			st.Bg(bg)
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
	radius := e.cornersOf(st, n.size)

	// The shadow lies outside the element, so it goes down before the
	// element's own clip.
	if n.style.shadow != nil {
		e.paintShadow(*n.style.shadow, rect, radius)
	}
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
		defer radius.rrect(rect).Push(gtx.Ops).Pop()

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
					if n.dragAccept != nil {
						state.conditionalDrag.update(gtx.Metric, gtx.Source, n.dragAccept)
					} else {
						state.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
					}
				}
				state.fresh = false
			}
			if n.onContextMenu != nil {
				// Observe above interactive descendants, passing events through.
				defer func() {
					area := clip.Rect(rect).Push(gtx.Ops)
					gtx.Event(pointer.Filter{Target: &state.contextTag, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
					pass := pointer.PassOp{}.Push(gtx.Ops)
					event.Op(gtx.Ops, &state.contextTag)
					pass.Pop()
					area.Pop()
				}()
			}
			state.click.Add(gtx.Ops)
			if n.onDrag != nil {
				if n.dragAccept != nil {
					state.conditionalDrag.add(gtx.Ops)
				} else {
					state.drag.Add(gtx.Ops)
				}
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

	if state != nil && n.onScroll != nil && !n.effectiveDisabled && !e.blockInput {
		area := clip.Rect(rect).Push(gtx.Ops)
		gtx.Event(n.onScroll.filter(&state.scrollTag, gtx.Metric))
		event.Op(gtx.Ops, &state.scrollTag)
		defer area.Pop()
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
	if n.isText && n.role == "" && n.name == "" && n.value == "" && n.selected == nil {
		// Plain text: the Gio label reports itself, in an area padded to
		// cover glyphs that reach past the line box. A clip of our own would
		// cut those (the tails of g and y in tight fonts such as YaHei).
		return nil
	}
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

// corners are the radii in px of a box's corners: top left, top right,
// bottom right, bottom left.
type corners [4]int

// cornersOf resolves a style's radius, or its separate corners, for a box of
// size, never more than half the shorter side.
func (e *engine) cornersOf(st Style, size image.Point) corners {
	limit := min(size.X, size.Y) / 2
	var c corners
	for i := range c {
		r := st.radius
		if st.corners != nil {
			r = st.corners[i]
		}
		c[i] = min(e.dp(r), limit)
	}
	return c
}

func (c corners) rrect(rect image.Rectangle) clip.RRect {
	return clip.RRect{Rect: rect, NW: c[0], NE: c[1], SE: c[2], SW: c[3]}
}

// inset is c for a box shrunk by d on every side.
func (c corners) inset(d int) corners {
	for i := range c {
		c[i] = max(c[i]-d, 0)
	}
	return c
}

func (c corners) max() int { return max(c[0], c[1], c[2], c[3]) }

func (e *engine) paintBox(st Style, rect image.Rectangle, radius corners) {
	ops := e.gtx.Ops
	if g := st.gradient; g != nil {
		paintGradient(ops, *g, rect, radius)
	} else if st.bg != nil {
		paint.FillShape(ops, *st.bg, radius.rrect(rect).Op(ops))
	}
	if bw := e.dp(st.borderWidth); bw > 0 {
		// Stroke centered on a rectangle inset by half the width, so the
		// border stays inside the element.
		h := bw / 2
		r := image.Rectangle{Min: rect.Min.Add(image.Pt(h, h)), Max: rect.Max.Sub(image.Pt(bw-h, bw-h))}
		path := radius.inset(h).rrect(r).Path(ops)
		if st.borderDashed {
			path = dashedBorderPath(ops, r, float32(radius.inset(h).max()), float32(e.dp(4)), float32(e.dp(3)))
		}
		paint.FillShape(ops, st.borderColor, clip.Stroke{Path: path, Width: float32(bw)}.Op())
	}
}

// paintShadow approximates a blurred shadow with rounded rectangles that grow
// and fade: their translucent layers add up to theme.Shadow under the element
// and thin out to nothing Blur dp beyond its edge.
func (e *engine) paintShadow(sh theme.Elevation, rect image.Rectangle, radius corners) {
	ops := e.gtx.Ops
	blur, offset := e.dp(sh.Blur), e.dp(sh.Offset)
	if blur <= 0 || theme.Shadow.A == 0 {
		return
	}
	layers := min(max(blur/2, 3), 16)
	c := theme.Shadow
	c.A = uint8(max(1, int(theme.Shadow.A)/layers))
	base := rect.Add(image.Pt(0, offset))
	for i := 1; i <= layers; i++ {
		grow := blur * i / layers
		r := image.Rectangle{Min: base.Min.Sub(image.Pt(grow, grow)), Max: base.Max.Add(image.Pt(grow, grow))}
		paint.FillShape(ops, c, radius.inset(-grow).rrect(r).Op(ops))
	}
}

func (e *engine) paintText(n *Node, inner image.Rectangle) {
	gtx := e.gtx
	defer op.Offset(inner.Min.Add(image.Pt(0, e.textShift(n, scriptOf(n.text))))).Push(gtx.Ops).Pop()
	g := gtx
	g.Constraints = layout.Constraints{Max: inner.Size()}
	if n.shimmer != nil {
		e.paintShimmerText(n, g, inner.Size())
	} else if len(n.textRanges) > 0 {
		e.paintRangeText(n, g, inner.Size())
	} else {
		lb := e.label(n, n.text)
		if !e.paintAtlasLabel(n, g, lb) {
			e.paintLabel(g, lb)
		}
	}
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
	// its ends, so OnKey takes them before the editor sees them. Escape also
	// reaches the handler for inline controls; modal layers handle it earlier.
	// A multiline box keeps those keys for editing; it gives OnKey only the
	// keys it captures, e.g. a plain Enter to send while Shift+Enter breaks
	// the line. An input method composing text consumes Enter first.
	if n.onKey != nil && (!spec.multiline || len(spec.captureKeys) > 0) && gtx.Enabled() {
		var filters []event.Filter
		if !spec.multiline {
			filters = append(filters,
				key.Filter{Focus: ed, Name: key.NameUpArrow}, key.Filter{Focus: ed, Name: key.NameDownArrow},
				key.Filter{Focus: ed, Name: key.NamePageUp}, key.Filter{Focus: ed, Name: key.NamePageDown},
				key.Filter{Focus: ed, Name: key.NameEscape})
		}
		for _, name := range spec.captureKeys {
			if name != "" {
				filters = append(filters, key.Filter{Focus: ed, Name: key.Name(name)})
			}
		}
		for {
			ev, ok := gtx.Event(filters...)
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
	e.inputMenuActions(n, st)
	if spec.document != nil {
		e.inputDocumentEvents(n, st)
	} else {
		st.inputDocument = nil
		e.inputPasteKeys(n, st)
		e.inputUndoKeys(n, st)
		// User edits first, then program changes to the bound string.
		for {
			beforeStart, beforeEnd := ed.Selection()
			before := InputEdit{Text: st.lastText, Start: beforeStart, End: beforeEnd}
			ev, ok := ed.Update(gtx)
			if !ok && len(st.inputActions) > 0 {
				action := st.inputActions[0]
				st.inputActions = st.inputActions[1:]
				ev, ok = e.inputAction(n, st, action)
			}
			if !ok {
				ev, ok = e.inputPasteEvent(n, st)
			}
			if !ok {
				break
			}
			switch ev.(type) {
			case widget.ChangeEvent:
				text := ed.Text()
				if text == st.lastText {
					break // SetText below, not the user
				}
				if spec.transform != nil || spec.transformEdit != nil {
					start, end := ed.Selection()
					normalized := InputEdit{Text: text, Start: start, End: end}
					if spec.transformEdit != nil {
						normalized = spec.transformEdit(before, normalized)
					} else {
						normalized = spec.transform(normalized)
					}
					if normalized.Text != text {
						ed.SetText(normalized.Text)
					}
					if normalized.Text != text || normalized.Start != start || normalized.End != end {
						ed.SetCaret(normalized.Start, normalized.End)
					}
					text = ed.Text()
					if text == st.lastText {
						break
					}
				}
				st.inputUndo = append(st.inputUndo, before)
				if len(st.inputUndo) > 100 {
					st.inputUndo = st.inputUndo[1:]
				}
				st.inputRedo = nil
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
			st.inputUndo, st.inputRedo = nil, nil
			st.lastText = *spec.bind
			ed.SetText(*spec.bind)
		}
	}
	focused := gtx.Enabled() && gtx.Focused(ed)
	if focused && !st.inputFocused && spec.selectOnFocus {
		ed.SetCaret(0, ed.Len())
	}
	st.inputFocused = focused
	if selection := st.inputSelection; selection != nil {
		if spec.document != nil {
			d := spec.document
			if r, err := inputcontent.ByteRange(d.Content().Text(), InputRange{Start: selection[0], End: selection[1]}); err == nil {
				_ = d.session.SelectSource(r)
				documentSyncEditor(st, d)
			}
		} else {
			ed.SetCaret(selection[0], selection[1])
		}
		st.inputSelection = nil
	}
	// A single-line box taller than its line (theme.ControlHeight) centers
	// the line, like a native field; the editor keeps the line's height.
	if !spec.multiline && spec.line > 0 && inner.Dy() > spec.line {
		inner.Min.Y += (inner.Dy() - spec.line) / 2
		inner.Max.Y = inner.Min.Y + spec.line
	}
	defer op.Offset(inner.Min.Add(image.Pt(0, e.textShift(n, scriptMixed)))).Push(gtx.Ops).Pop()
	g := gtx
	g.Constraints = layout.Exact(inner.Size())
	ts := n.textStyle
	th := theme.Material
	shaper := theme.Material.Shaper
	if spec.document != nil {
		e.prepareInputObjects(n, st, inner.Dx())
		documentSyncEditor(st, spec.document)
		if st.inputObjects.shaper != nil {
			copy := *th
			copy.Shaper = st.inputObjects.shaper
			th = &copy
			shaper = copy.Shaper
		}
	}
	me := material.Editor(th, ed, spec.placeholder)
	me.TextSize, me.Color, me.HintColor = ts.size, *ts.color, theme.Muted
	me.Font = textFont(ts)
	if spec.document != nil && st.inputObjects.shaper != nil {
		me.Font = st.inputObjects.font
	}
	me.LineHeightScale = ts.lineHeight
	if spec.document != nil {
		st.caret.LayoutDecorated(g, me, shaper, func(g layout.Context) { e.paintInputTokens(n, st, g) })
	} else {
		st.caret.Layout(g, me, theme.Material.Shaper)
	}
	if spec.document != nil {
		e.inputDocumentIME(n, st, g)
	}
	// Keep the value in the semantic tree for agents.
	value := ed.Text()
	if spec.document != nil {
		value = spec.document.Content().Text()
	}
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
	viewportHovered := st.scrollHover.Update(gtx.Source)
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
	xTrack, yTrack := scrollbarTracks(viewport, e.cornersOf(n.style, n.size).inset(bw), barWidth,
		n.style.scrollX && maxX > 0 && n.style.scrollY && maxScroll > 0)
	if n.style.scrollX && n.style.controlledScroll == nil {
		st.scrollX = st.scrollbarX.update(gtx, xTrack, true, st.scrollX, viewport.Dx(), totalX, e.dp(24), st)
	}
	if n.style.scrollY && n.style.controlledScroll == nil {
		st.scrollY = st.scrollbarY.update(gtx, yTrack, false, st.scrollY, viewport.Dy(), total, e.dp(24), st)
	}

	if offset := n.style.controlledScroll; offset != nil {
		if n.style.scrollX {
			st.scrollX = min(max(e.dp(offset[0]), 0), maxX)
		}
		if n.style.scrollY {
			st.scrollY = min(max(e.dp(offset[1]), 0), maxScroll)
		}
	}

	if gtx.Enabled() && (st.scrollX != previousX || st.scrollY != previousY) {
		// Virtual content is built before paint; rebuild at the new offset.
		gtx.Execute(op.InvalidateCmd{})
	}

	if st.scrollX != previousX || st.scrollY != previousY || st.scrollbarX.active || st.scrollbarY.active {
		st.scrollVisibleUntil = gtx.Now.Add(ScrollbarLinger)
	}
	mode := resolveScrollbars(n.style.scrollbarMode, n.style.scrollbarModeSet)
	showBars := true
	switch mode {
	case ScrollbarHover:
		showBars = viewportHovered || st.scrollbarX.active || st.scrollbarY.active
	case ScrollbarScrolling:
		showBars = gtx.Now.Before(st.scrollVisibleUntil) || st.scrollbarX.active || st.scrollbarY.active
		if showBars && gtx.Enabled() && n.style.controlledScroll == nil {
			gtx.Execute(op.InvalidateCmd{At: st.scrollVisibleUntil})
		}
	}
	// Fade between shown and hidden; wanted bars take input at any opacity.
	// Always mode does not fade, and switching away from it hides at once.
	alpha := float32(1)
	if mode != ScrollbarAlways {
		if showBars {
			if st.scrollAlpha == 0 {
				st.scrollShownAt = gtx.Now
			}
			st.scrollWantedAt = gtx.Now
		}
		alpha = scrollbarAlpha(showBars, gtx.Now, st.scrollShownAt, st.scrollWantedAt)
		if alpha > 0 && alpha < 1 && gtx.Enabled() {
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	st.scrollAlpha = alpha
	if mode == ScrollbarAlways {
		st.scrollWantedAt = time.Time{}
	}

	// The viewport is its own area with the scroll handler, so it is a node
	// in the semantic tree and agents only see what shows through it.
	stk := clip.Rect(viewport).Push(gtx.Ops)
	if n.style.scrollY && n.style.controlledScroll == nil {
		st.scroll.Add(gtx.Ops)
	}
	if n.style.scrollX && n.style.controlledScroll == nil {
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

	if mode == ScrollbarHover && n.style.controlledScroll == nil {
		pass := pointer.PassOp{}.Push(gtx.Ops)
		st.scrollHover.Add(gtx.Ops)
		pass.Pop()
	}
	if (alpha > 0 || showBars) && n.style.scrollY && n.style.controlledScroll == nil {
		st.scrollbarY.paint(gtx, yTrack, false, st.scrollY, viewport.Dy(), total, e.dp(24), alpha, showBars)
	}
	if (alpha > 0 || showBars) && n.style.scrollX && n.style.controlledScroll == nil {
		st.scrollbarX.paint(gtx, xTrack, true, st.scrollX, viewport.Dx(), totalX, e.dp(24), alpha, showBars)
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

// shiftKey identifies a text style whose ink has been measured.
type shiftKey struct {
	face       font.Typeface
	px         int
	weight     font.Weight
	lineHeight float32
	script     script
}

// shifts caches textShift per text style. Touched only under the frame lock.
var shifts = map[shiftKey]int{}

// script says which faces set a text, and so which line box it gets: East
// Asian text usually comes from a fallback face (PingFang behind SF Pro) whose
// ascent, descent and body differ from the Latin face's. Mixed text takes the
// taller extent of both.
type script uint8

const (
	scriptLatin script = iota
	scriptWide
	scriptMixed
)

func scriptOf(s string) script {
	var latin, wide bool
	for _, r := range s {
		switch {
		case unicode.IsSpace(r):
			// Spaces are set by the run around them.
		case r >= 0x1100 && (unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) ||
			r >= 0x3000 && r <= 0x303f || r >= 0xff00 && r <= 0xffef):
			wide = true
		default:
			latin = true
		}
	}
	switch {
	case wide && latin:
		return scriptMixed
	case wide:
		return scriptWide
	}
	return scriptLatin
}

// textShift moves glyphs so their ink, not the font's line box, is centered
// in the box, matching icons beside the text. Fonts differ: macOS PingFang
// leaves a deep empty descent, so CJK text sits high; Windows YaHei fits its
// ink tightly. The shift is measured in the line box of the faces that set
// the text: Latin centers its cap height, East Asian text centers 国, and mixed
// text centers both. One mixed measurement for all text set Latin labels low
// and Chinese labels high beside their icons. The shift never pushes the
// lowest or highest ink further out of the line box.
func (e *engine) textShift(n *Node, sc script) int {
	ts := n.textStyle
	face := textFont(ts)
	px := e.m.Sp(ts.size)
	key := shiftKey{face.Typeface, px, face.Weight, ts.lineHeight, sc}
	if s, ok := shifts[key]; ok {
		return s
	}
	// The line box is what a Label of this style measures: Gio sizes a text
	// box that way, and fallback fonts (国 in a Latin face) would otherwise
	// skew metrics read from single glyphs.
	// Bodies: A, 国, or both; g gives the lowest ink.
	sample, body := "Ag", 1
	switch sc {
	case scriptWide:
		sample = "国"
	case scriptMixed:
		sample, body = "国Ag", 2
	}
	lb := material.Label(theme.Material, ts.size, sample)
	lb.Font = face
	if ts.lineHeight > 0 {
		lb.LineHeightScale = ts.lineHeight
	}
	dims := e.measureLabel(e.measureGtx(layout.Constraints{Max: image.Pt(inf, inf)}), lb)
	boxDescent := dims.Baseline // below the baseline
	boxAscent := dims.Size.Y - boxDescent
	shaper := theme.Material.Shaper
	shaper.LayoutString(text.Parameters{Font: face, PxPerEm: fixed.I(px), MaxWidth: 1 << 20}, sample)
	bodyTop, bodyBottom, highest, lowest := 0, 0, 0, 0
	found := false
	for i := 0; ; i++ {
		g, ok := shaper.NextGlyph()
		if !ok {
			break
		}
		if g.Bounds.Min.Y == g.Bounds.Max.Y {
			continue
		}
		top, bottom := g.Bounds.Min.Y.Floor(), g.Bounds.Max.Y.Ceil()
		highest, lowest = min(highest, top), max(lowest, bottom)
		if i < body {
			if !found {
				bodyTop, bodyBottom, found = top, bottom, true
			}
			bodyTop, bodyBottom = min(bodyTop, top), max(bodyBottom, bottom)
		}
	}
	shift := 0
	if found {
		// The line box runs from -boxAscent to +boxDescent around the baseline.
		// Round halves down (positive): truncating, or rounding -0.5 away
		// from zero, left CJK a visible half-pixel step high at 2x.
		shift = int(math.Floor(float64((boxDescent-boxAscent)-(bodyTop+bodyBottom))/2 + 0.5))
		shift = max(min(shift, max(boxDescent-lowest, 0)), min(-boxAscent-highest, 0))
	}
	shifts[key] = shift
	return shift
}

// paintGradient fills a rounded rectangle with g. The stops sit where the
// gradient line, through the center at g.Angle, leaves the rectangle, so the
// corners get the end colors whatever the angle.
func paintGradient(ops *op.Ops, g theme.Gradient, rect image.Rectangle, radius corners) {
	a := float64(g.Angle) * math.Pi / 180
	dx, dy := math.Cos(a), math.Sin(a)
	w, h := float64(rect.Dx()), float64(rect.Dy())
	half := (math.Abs(w*dx) + math.Abs(h*dy)) / 2
	cx, cy := float64(rect.Min.X)+w/2, float64(rect.Min.Y)+h/2
	defer radius.rrect(rect).Push(ops).Pop()
	paint.LinearGradientOp{
		Stop1: f32.Pt(float32(cx-dx*half), float32(cy-dy*half)), Color1: g.From,
		Stop2: f32.Pt(float32(cx+dx*half), float32(cy+dy*half)), Color2: g.To,
	}.Add(ops)
	paint.PaintOp{}.Add(ops)
}

// bgEase is how long a hover or press background takes to change.
const bgEase = 120 * time.Millisecond

// easeBackground fades an element's background between its resting, hover
// and pressed colors instead of switching at once. Reduced motion switches.
func (e *engine) easeBackground(st *Style, state *elemState) {
	var target color.NRGBA
	if st.bg != nil {
		target = *st.bg
	}
	now := e.gtx.Now
	if !state.bgInit || theme.ReducedMotion {
		state.bgInit, state.bgShown, state.bgTo = true, target, target
		return
	}
	if target != state.bgTo {
		state.bgFrom, state.bgTo, state.bgStart = state.bgShown, target, now
	}
	t := float32(now.Sub(state.bgStart)) / float32(bgEase)
	shown := target
	if t < 1 {
		shown = lerpColor(state.bgFrom, target, t)
		e.gtx.Execute(op.InvalidateCmd{})
	}
	state.bgShown = shown
	if shown.A != 0 || st.bg != nil {
		st.bg = &shown
	}
}

// lerpColor blends two colors; a transparent end keeps the other's hue.
func lerpColor(a, b color.NRGBA, t float32) color.NRGBA {
	if a.A == 0 {
		a.R, a.G, a.B = b.R, b.G, b.B
	}
	if b.A == 0 {
		b.R, b.G, b.B = a.R, a.G, a.B
	}
	mix := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t + 0.5) }
	return color.NRGBA{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: mix(a.A, b.A)}
}
