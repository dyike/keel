package kit

import (
	"image"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"

	"github.com/dyike/keel/ui/core"
)

const codeHoverDelay = 500 * time.Millisecond

// update handles pointer, keyboard, input method and clipboard events.
func (v *CodeEditorView) update(gtx core.C) {
	if !gtx.Enabled() {
		return
	}
	v.pointer(gtx)
	v.keys(gtx)
	v.hover(gtx)
}

func (v *CodeEditorView) pointer(gtx core.C) {
	m := v.metrics
	// Wheel and trackpad scrolling.
	total := len(v.buf.lines)*m.lh - m.size.Y + m.lh
	if d := v.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, 1, pointer.ScrollRange{}, pointer.ScrollRange{Min: -int(v.scrollY), Max: max(0, total-int(v.scrollY))}); d != 0 {
		v.scrollY += float32(d)
		v.hoverText = ""
	}
	if d := v.scrollH.Update(gtx.Metric, gtx.Source, gtx.Now, 0, pointer.ScrollRange{Min: -int(v.scrollX), Max: 1 << 20}, pointer.ScrollRange{}); d != 0 {
		v.scrollX += float32(d)
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: v, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Move | pointer.Leave | pointer.Cancel})
		if !ok {
			break
		}
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Kind {
		case pointer.Press:
			if !e.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			gtx.Execute(key.FocusCmd{Tag: v})
			v.hoverText, v.tabExits = "", false
			if len(v.comp) > 0 {
				if r := v.completionRect(gtx); e.Position.Round().In(r) {
					i := v.completionTop() + (e.Position.Round().Y-r.Min.Y-gtx.Dp(4))/max(1, m.lh)
					if i >= 0 && i < len(v.comp) {
						v.compSel = i
						v.acceptCompletion(gtx)
					}
					continue
				}
				v.comp = nil
			}
			if e.Position.X < float32(m.gutter) {
				// The gutter selects whole lines.
				p := v.posAt(e.Position)
				v.anchor, v.caret = codePos{p.line, 0}, v.buf.clamp(codePos{p.line + 1, 0})
				if p.line == len(v.buf.lines)-1 {
					v.caret = codePos{p.line, len(v.buf.lines[p.line])}
				}
				v.dragging, v.blink = false, gtx.Now
				continue
			}
			if gtx.Now.Sub(v.lastPress) < 400*time.Millisecond && e.Position.Sub(v.pressAt).X < 4 && e.Position.Sub(v.pressAt).Y < 4 {
				v.pressCount++
			} else {
				v.pressCount = 1
			}
			v.lastPress, v.pressAt = gtx.Now, e.Position
			p := v.posAt(e.Position)
			switch {
			case v.pressCount == 2:
				v.anchor, v.caret = v.buf.wordAt(p)
			case v.pressCount >= 3:
				v.anchor, v.caret = codePos{p.line, 0}, codePos{p.line, len(v.buf.lines[p.line])}
			case e.Modifiers.Contain(key.ModShift):
				v.caret = p
			default:
				v.anchor, v.caret = p, p
			}
			v.dragging, v.goalX, v.blink = true, -1, gtx.Now
		case pointer.Drag:
			if v.dragging {
				// Past the top or bottom edge, scroll while selecting.
				if e.Position.Y < 0 {
					v.scrollY -= float32(m.lh)
				} else if e.Position.Y > float32(m.size.Y) {
					v.scrollY += float32(m.lh)
				}
				v.caret = v.posAt(f32.Pt(max(e.Position.X, float32(m.gutter)), e.Position.Y))
				v.reveal = true
			}
		case pointer.Release, pointer.Cancel:
			v.dragging = false
		case pointer.Move:
			if e.Position != v.hoverAt {
				v.hoverAt, v.hoverSince, v.hoverDone, v.hoverText = e.Position, gtx.Now, false, ""
			}
			v.hovering = true
		case pointer.Leave:
			v.hovering, v.hoverText = false, ""
		}
	}
}

// hover shows diagnostics and the OnHover tip after the pointer rests.
func (v *CodeEditorView) hover(gtx core.C) {
	if !v.hovering || v.hoverDone || v.dragging || v.hoverAt.X < float32(v.metrics.gutter) {
		return
	}
	if wait := codeHoverDelay - gtx.Now.Sub(v.hoverSince); wait > 0 {
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(wait)})
		return
	}
	v.hoverDone = true
	p := v.posAt(v.hoverAt)
	if p.col >= len(v.buf.lines[p.line]) {
		return
	}
	var parts []string
	for _, d := range v.diagnostics {
		start, end := codePos{d.Line, d.Col}, codePos{d.EndLine, max(d.EndCol, d.Col+1)}
		if !p.less(start) && p.less(end) {
			parts = append(parts, d.Message)
		}
	}
	if v.onHover != nil {
		fn := v.onHover
		var tip string
		core.Call(gtx, func() { tip = fn(p.line, p.col) })
		if tip != "" {
			parts = append(parts, tip)
		}
	}
	v.hoverText = strings.Join(parts, " · ")
}

func (v *CodeEditorView) keys(gtx core.C) {
	filters := []event.Filter{
		key.FocusFilter{Target: v},
		transfer.TargetFilter{Target: v, Type: "application/text"},
		key.Filter{Focus: v, Name: key.NameReturn, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameEnter, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameEscape},
		key.Filter{Focus: v, Name: key.NameDeleteBackward, Optional: key.ModShortcutAlt | key.ModShift},
		key.Filter{Focus: v, Name: key.NameDeleteForward, Optional: key.ModShortcutAlt | key.ModShift},
		key.Filter{Focus: v, Name: key.NameHome, Optional: key.ModShortcut | key.ModShift},
		key.Filter{Focus: v, Name: key.NameEnd, Optional: key.ModShortcut | key.ModShift},
		key.Filter{Focus: v, Name: key.NamePageUp, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NamePageDown, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameLeftArrow, Optional: key.ModShortcutAlt | key.ModShift | key.ModShortcut},
		key.Filter{Focus: v, Name: key.NameRightArrow, Optional: key.ModShortcutAlt | key.ModShift | key.ModShortcut},
		key.Filter{Focus: v, Name: key.NameUpArrow, Optional: key.ModShift | key.ModShortcut},
		key.Filter{Focus: v, Name: key.NameDownArrow, Optional: key.ModShift | key.ModShortcut},
		key.Filter{Focus: v, Name: "A", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "C", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "X", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "V", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "Z", Required: key.ModShortcut, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameSpace, Required: key.ModCtrl},
	}
	if !v.tabExits || len(v.comp) > 0 {
		filters = append(filters, key.Filter{Focus: v, Name: key.NameTab, Optional: key.ModShift})
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		v.blink = gtx.Now
		switch e := ev.(type) {
		case key.FocusEvent:
			if !e.Focus {
				v.comp, v.tabExits = nil, false
			}
			v.imeText, v.imeLine = "", -1
		case key.Event:
			if e.State == key.Press {
				v.command(gtx, e)
			}
		case key.EditEvent:
			v.input(gtx, e)
		case key.SnippetEvent:
			v.imeText = "" // resend the current line
		case key.SelectionEvent:
			line := v.buf.lines[v.caret.line]
			v.anchor = codePos{v.caret.line, min(max(e.Start, 0), len(line))}
			v.caret = codePos{v.caret.line, min(max(e.End, 0), len(line))}
		case transfer.DataEvent:
			data, err := io.ReadAll(e.Open())
			if err == nil && !v.readOnly {
				v.replaceSelection(gtx, string(data), false)
			}
		}
	}
}

// input applies typed text. The input method works in columns of the caret's
// line, as reportIME describes it; a selection across lines is replaced.
func (v *CodeEditorView) input(gtx core.C, e key.EditEvent) {
	if v.readOnly {
		return
	}
	from, to := ordered(v.anchor, v.caret)
	if from.line == to.line {
		n := len(v.buf.lines[v.caret.line])
		a, b := min(max(e.Range.Start, 0), n), min(max(e.Range.End, 0), n)
		if a > b {
			a, b = b, a
		}
		from, to = codePos{v.caret.line, a}, codePos{v.caret.line, b}
	}
	v.anchor, v.caret = from, to
	v.replaceSelection(gtx, e.Text, true)
	if len(e.Text) > 0 && v.onComplete != nil {
		if r, _ := utf8.DecodeLastRuneInString(e.Text); isIdent(r) || r == '.' {
			v.complete(gtx, false)
		} else if len(v.comp) > 0 {
			v.comp = nil
		}
	}
}

// replaceSelection puts text in place of the selection and tells OnChange.
func (v *CodeEditorView) replaceSelection(gtx core.C, text string, typing bool) {
	from, to := ordered(v.anchor, v.caret)
	end := v.buf.edit(from, to, text, [2]codePos{v.anchor, v.caret}, typing)
	v.anchor, v.caret = end, end
	v.changed(gtx)
}

func (v *CodeEditorView) changed(gtx core.C) {
	v.reveal, v.goalX, v.tabExits, v.hoverText = true, -1, false, ""
	if v.onChange != nil {
		fn, text := v.onChange, v.buf.text()
		core.Call(gtx, func() { fn(text) })
	}
}

func (v *CodeEditorView) command(gtx core.C, e key.Event) {
	shift := e.Modifiers.Contain(key.ModShift)
	word := e.Modifiers.Contain(key.ModShortcutAlt)
	shortcut := e.Modifiers.Contain(key.ModShortcut)
	move := func(p codePos) {
		v.caret = v.buf.clamp(p)
		if !shift {
			v.anchor = v.caret
		}
		v.reveal = true
	}
	if len(v.comp) > 0 {
		switch e.Name {
		case key.NameUpArrow:
			v.compSel = (v.compSel + len(v.comp) - 1) % len(v.comp)
			return
		case key.NameDownArrow:
			v.compSel = (v.compSel + 1) % len(v.comp)
			return
		case key.NameReturn, key.NameEnter, key.NameTab:
			v.acceptCompletion(gtx)
			return
		case key.NameEscape:
			v.comp = nil
			return
		}
		v.comp = nil
	}
	from, to := ordered(v.anchor, v.caret)
	switch e.Name {
	case key.NameEscape:
		v.anchor, v.tabExits = v.caret, true
	case key.NameLeftArrow:
		switch {
		case shortcut:
			move(codePos{v.caret.line, 0})
		case !shift && from != to:
			move(from)
		case word:
			move(v.wordStep(v.caret, -1))
		default:
			move(v.step(v.caret, -1))
		}
		v.goalX = -1
	case key.NameRightArrow:
		switch {
		case shortcut:
			move(codePos{v.caret.line, len(v.buf.lines[v.caret.line])})
		case !shift && from != to:
			move(to)
		case word:
			move(v.wordStep(v.caret, 1))
		default:
			move(v.step(v.caret, 1))
		}
		v.goalX = -1
	case key.NameUpArrow, key.NameDownArrow, key.NamePageUp, key.NamePageDown:
		lines := map[key.Name]int{key.NameUpArrow: -1, key.NameDownArrow: 1, key.NamePageUp: -v.visible, key.NamePageDown: v.visible}[e.Name]
		if shortcut {
			if lines < 0 {
				move(codePos{})
			} else {
				last := len(v.buf.lines) - 1
				move(codePos{last, len(v.buf.lines[last])})
			}
			return
		}
		if v.goalX < 0 {
			v.goalX = v.colX(v.caret.line)[v.caret.col]
		}
		line := min(max(v.caret.line+lines, 0), len(v.buf.lines)-1)
		xs := v.colX(line)
		col := 0
		for col < len(xs)-1 && v.goalX > (xs[col]+xs[col+1])/2 {
			col++
		}
		goal := v.goalX
		move(codePos{line, col})
		v.goalX = goal
	case key.NameHome:
		if shortcut {
			move(codePos{})
		} else {
			// Toggle between the first non-blank column and column 0.
			ind := len([]rune(v.buf.indent(v.caret.line)))
			if v.caret.col == ind {
				ind = 0
			}
			move(codePos{v.caret.line, ind})
		}
		v.goalX = -1
	case key.NameEnd:
		if shortcut {
			last := len(v.buf.lines) - 1
			move(codePos{last, len(v.buf.lines[last])})
		} else {
			move(codePos{v.caret.line, len(v.buf.lines[v.caret.line])})
		}
		v.goalX = -1
	case key.NameDeleteBackward, key.NameDeleteForward:
		if v.readOnly {
			return
		}
		if from == to {
			dir := 1
			if e.Name == key.NameDeleteBackward {
				dir = -1
			}
			switch {
			case shortcut && dir < 0:
				v.anchor = codePos{v.caret.line, 0}
			case word:
				v.anchor = v.wordStep(v.caret, dir)
			default:
				v.anchor = v.step(v.caret, dir)
			}
		}
		v.replaceSelection(gtx, "", false)
	case key.NameReturn, key.NameEnter:
		if v.readOnly {
			return
		}
		indent := v.buf.indent(v.caret.line)
		l := v.buf.lines[from.line]
		if from.col > 0 && strings.ContainsRune("{([:", l[from.col-1]) {
			indent += v.indentUnit()
		}
		v.replaceSelection(gtx, "\n"+indent, false)
	case key.NameTab:
		if v.readOnly {
			return
		}
		if from.line != to.line || shift {
			last := to.line
			if to.col == 0 && to.line > from.line {
				last-- // a selection ending at a line's start leaves that line alone
			}
			v.shiftLines(gtx, from.line, last, !shift)
			return
		}
		v.replaceSelection(gtx, v.indentUnit(), true)
	case key.NameSpace: // Ctrl+Space
		v.complete(gtx, true)
	case "A":
		last := len(v.buf.lines) - 1
		v.anchor, v.caret = codePos{}, codePos{last, len(v.buf.lines[last])}
	case "C", "X":
		if s := v.buf.slice(from, to); s != "" {
			gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(s))})
			if e.Name == "X" && !v.readOnly {
				v.replaceSelection(gtx, "", false)
			}
		}
	case "V":
		if !v.readOnly {
			gtx.Execute(clipboard.ReadCmd{Tag: v})
		}
	case "Z":
		if v.readOnly {
			return
		}
		var sel [2]codePos
		var ok bool
		if shift {
			sel, ok = v.buf.redoOne()
		} else {
			sel, ok = v.buf.undoOne()
		}
		if ok {
			v.anchor, v.caret = v.buf.clamp(sel[0]), v.buf.clamp(sel[1])
			v.changed(gtx)
		}
	}
}

// step moves one rune, across line ends.
func (v *CodeEditorView) step(p codePos, dir int) codePos {
	p.col += dir
	if p.col < 0 && p.line > 0 {
		return codePos{p.line - 1, len(v.buf.lines[p.line-1])}
	}
	if p.col > len(v.buf.lines[p.line]) && p.line < len(v.buf.lines)-1 {
		return codePos{p.line + 1, 0}
	}
	return v.buf.clamp(p)
}

// wordStep moves to the next word boundary in a direction.
func (v *CodeEditorView) wordStep(p codePos, dir int) codePos {
	l := v.buf.lines[p.line]
	if dir < 0 {
		if p.col == 0 {
			return v.step(p, -1)
		}
		c := p.col
		for c > 0 && !isIdent(l[c-1]) {
			c--
		}
		for c > 0 && isIdent(l[c-1]) {
			c--
		}
		return codePos{p.line, c}
	}
	if p.col == len(l) {
		return v.step(p, 1)
	}
	c := p.col
	for c < len(l) && !isIdent(l[c]) {
		c++
	}
	for c < len(l) && isIdent(l[c]) {
		c++
	}
	return codePos{p.line, c}
}

// indentUnit is a tab in files indented with tabs, else four spaces.
func (v *CodeEditorView) indentUnit() string {
	for i := max(0, v.caret.line-50); i < min(len(v.buf.lines), v.caret.line+50); i++ {
		if l := v.buf.lines[i]; len(l) > 0 && l[0] == '\t' {
			return "\t"
		}
	}
	return "    "
}

// shiftLines indents or outdents lines as one undo step.
func (v *CodeEditorView) shiftLines(gtx core.C, first, last int, in bool) {
	unit := v.indentUnit()
	start, end := codePos{first, 0}, codePos{last, len(v.buf.lines[last])}
	lines := strings.Split(v.buf.slice(start, end), "\n")
	for i, l := range lines {
		switch {
		case in:
			lines[i] = unit + l
		case strings.HasPrefix(l, "\t"):
			lines[i] = l[1:]
		default:
			lines[i] = strings.TrimPrefix(l, strings.Repeat(" ", min(4, len(l)-len(strings.TrimLeft(l, " ")))))
		}
	}
	v.buf.edit(start, end, strings.Join(lines, "\n"), [2]codePos{v.anchor, v.caret}, false)
	v.anchor = codePos{first, 0}
	v.caret = codePos{last, len(v.buf.lines[last])}
	v.changed(gtx)
}

// complete asks OnComplete for the word before the caret.
func (v *CodeEditorView) complete(gtx core.C, explicit bool) {
	if v.onComplete == nil {
		return
	}
	start, _ := v.buf.wordAt(v.caret)
	start.col = min(start.col, v.caret.col)
	prefix := string(v.buf.lines[v.caret.line][start.col:v.caret.col])
	if prefix == "" && !explicit {
		v.comp = nil
		return
	}
	fn := v.onComplete
	var items []CodeCompletion
	core.Call(gtx, func() { items = fn(v.caret.line, v.caret.col, prefix) })
	v.comp, v.compSel, v.compFrom = items, 0, codePos{v.caret.line, start.col}
}

func (v *CodeEditorView) acceptCompletion(gtx core.C) {
	c := v.comp[v.compSel]
	v.comp = nil
	insert := c.Insert
	if insert == "" {
		insert = c.Label
	}
	v.anchor = v.compFrom
	v.replaceSelection(gtx, insert, false)
}

// reportIME tells the platform input method the caret's line, the selection
// in it and where the caret is, so candidates appear beside it.
func (v *CodeEditorView) reportIME(gtx core.C) {
	if !v.focused {
		return
	}
	line := string(v.buf.lines[v.caret.line])
	if v.imeLine != v.caret.line || v.imeText != line {
		v.imeLine, v.imeText = v.caret.line, line
		gtx.Execute(key.SnippetCmd{Tag: v, Snippet: key.Snippet{Range: key.Range{Start: 0, End: len(v.buf.lines[v.caret.line])}, Text: line}})
	}
	rng := key.Range{Start: v.caret.col, End: v.caret.col}
	if v.anchor.line == v.caret.line {
		rng = key.Range{Start: v.anchor.col, End: v.caret.col}
	}
	p := v.caretPoint(v.caret).Add(image.Pt(0, v.metrics.baseline))
	if rng != v.imeRange || p != v.imeCaret {
		v.imeRange, v.imeCaret = rng, p
		gtx.Execute(key.SelectionCmd{Tag: v, Range: rng, Caret: key.Caret{Pos: layout.FPt(p), Ascent: float32(v.metrics.ascent), Descent: float32(v.metrics.descent)}})
	}
}
