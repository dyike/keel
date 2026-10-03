package kit

import (
	"image"
	"io"
	"slices"
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

// codeReplace is one replacement of a command: from..to becomes text, and
// the caret lands caret runes into it (-1: at its end). With keep set, the
// inserted text minus keep runes at each end is selected instead (wrapping a
// selection in brackets keeps it selected).
type codeReplace struct {
	from, to codePos
	text     string
	caret    int
	keep     int
	keepEnd  int
}

// applyReplaces makes replacements as one undo step, top to bottom, moving
// later positions for earlier edits, and leaves a caret per replacement.
func (v *CodeEditorView) applyReplaces(reps []codeReplace, typing bool) {
	slices.SortFunc(reps, func(a, b codeReplace) int {
		if a.from.less(b.from) {
			return -1
		}
		if b.from.less(a.from) {
			return 1
		}
		return 0
	})
	v.buf.begin(v.sels)
	sels := make([]codeSel, 0, len(reps))
	var lastTo, lastEnd codePos
	moved := false
	primAt := v.primary().caret
	newPrim := 0
	for i, r := range reps {
		if moved {
			r.from, r.to = shiftPos(r.from, lastTo, lastEnd), shiftPos(r.to, lastTo, lastEnd)
		}
		end := v.buf.edit(r.from, r.to, r.text)
		lastTo, lastEnd, moved = r.to, end, true
		var s codeSel
		switch {
		case r.keep > 0:
			n := utf8.RuneCountInString(r.text)
			suffix := r.keepEnd
			if suffix == 0 {
				suffix = r.keep
			}
			s = codeSel{endOf(r.from, string([]rune(r.text)[:r.keep])), endOf(r.from, string([]rune(r.text)[:n-suffix]))}
		case r.caret >= 0:
			p := endOf(r.from, string([]rune(r.text)[:r.caret]))
			s = codeSel{p, p}
		default:
			s = codeSel{end, end}
		}
		sels = append(sels, s)
		if !primAt.less(reps[i].from) {
			newPrim = i
		}
	}
	v.sels, v.prim = sels, newPrim
	v.normalize()
	v.buf.commit(v.sels, typing && len(reps) == 1)
	first, last := reps[0].from.line, lastEnd.line
	v.buf.highlightNear(v.lang, codeStyleName(), first, last)
}

// normalize sorts the selections and merges overlapping ones, keeping the
// primary.
func (v *CodeEditorView) normalize() {
	if len(v.sels) == 0 {
		v.sels, v.prim = []codeSel{{}}, 0
		return
	}
	for i := range v.sels {
		v.sels[i].anchor, v.sels[i].caret = v.buf.clamp(v.sels[i].anchor), v.buf.clamp(v.sels[i].caret)
	}
	prim := v.sels[min(v.prim, len(v.sels)-1)]
	slices.SortStableFunc(v.sels, func(a, b codeSel) int {
		af, _ := a.span()
		bf, _ := b.span()
		if af.less(bf) {
			return -1
		}
		if bf.less(af) {
			return 1
		}
		return 0
	})
	out := v.sels[:1]
	for _, s := range v.sels[1:] {
		last := &out[len(out)-1]
		lf, lt := last.span()
		sf, st := s.span()
		if sf.less(lt) || sf == lt && (s.empty() || last.empty()) {
			if lt.less(st) {
				lt = st
			}
			if last.caret.less(last.anchor) {
				*last = codeSel{lt, lf}
			} else {
				*last = codeSel{lf, lt}
			}
			if s == prim {
				prim = *last
			}
			continue
		}
		out = append(out, s)
	}
	v.sels, v.prim = out, 0
	for i, s := range out {
		if s == prim || s.caret == prim.caret {
			v.prim = i
		}
	}
}

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
	total := v.rowCount()*m.lh - m.size.Y + m.lh
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
		alt := e.Modifiers.Contain(key.ModAlt)
		link := e.Modifiers.Contain(key.ModShortcut) && !alt
		switch e.Kind {
		case pointer.Press:
			if !e.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			gtx.Execute(key.FocusCmd{Tag: v})
			v.hoverText, v.tabExits, v.blink = "", false, gtx.Now
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
				line := v.lineOf(int((e.Position.Y + v.scrollY) / float32(max(1, m.lh))))
				if e.Position.X < float32(m.foldW) {
					v.toggleFold(line)
					continue
				}
				// The line numbers select whole lines.
				end := v.foldEnd(line)
				if !v.folds[line] {
					end = line
				}
				to := v.buf.clamp(codePos{end + 1, 0})
				if end == v.buf.count()-1 {
					to = codePos{end, len(v.buf.line(end))}
				}
				v.sels, v.prim = []codeSel{{codePos{line, 0}, to}}, 0
				continue
			}
			if gtx.Now.Sub(v.lastPress) < 400*time.Millisecond && abs32(e.Position.X-v.pressAt.X) < 4 && abs32(e.Position.Y-v.pressAt.Y) < 4 {
				v.pressCount++
			} else {
				v.pressCount = 1
			}
			v.lastPress, v.pressAt = gtx.Now, e.Position
			p := v.posAt(e.Position)
			switch {
			case link && v.onDefinition != nil:
				fn := v.onDefinition
				w, _ := v.buf.wordAt(p)
				core.Call(gtx, func() { fn(w.line, w.col) })
				continue
			case alt:
				// Alt+click adds a caret; dragging builds a block from here.
				v.sels = append(v.sels, codeSel{p, p})
				v.prim = len(v.sels) - 1
				v.column, v.columnFrom = true, e.Position
				v.normalize()
			case v.pressCount == 2:
				a, b := v.buf.wordAt(p)
				v.sels, v.prim = []codeSel{{a, b}}, 0
			case v.pressCount >= 3:
				v.sels, v.prim = []codeSel{{codePos{p.line, 0}, codePos{p.line, len(v.buf.line(p.line))}}}, 0
			case e.Modifiers.Contain(key.ModShift):
				s := v.primary()
				v.sels, v.prim = []codeSel{{s.anchor, p}}, 0
			default:
				v.sels, v.prim = []codeSel{{p, p}}, 0
				v.column = false
			}
			v.dragging, v.goalX = true, -1
		case pointer.Drag:
			if !v.dragging {
				continue
			}
			if e.Position.Y < 0 {
				v.scrollY -= float32(m.lh)
			} else if e.Position.Y > float32(m.size.Y) {
				v.scrollY += float32(m.lh)
			}
			at := f32.Pt(max(e.Position.X, float32(m.gutter)), e.Position.Y)
			if v.column {
				v.columnSelect(v.columnFrom, at)
			} else {
				s := &v.sels[v.prim]
				s.caret = v.posAt(at)
			}
			v.reveal = true
		case pointer.Release, pointer.Cancel:
			v.dragging, v.column = false, false
		case pointer.Move:
			if e.Position != v.hoverAt {
				v.hoverAt, v.hoverSince, v.hoverDone, v.hoverText = e.Position, gtx.Now, false, ""
			}
			v.hovering = true
			v.linkFrom, v.linkTo = codePos{}, codePos{}
			if link && v.onDefinition != nil && e.Position.X >= float32(m.gutter) {
				v.linkFrom, v.linkTo = v.buf.wordAt(v.posAt(e.Position))
			}
		case pointer.Leave:
			v.hovering, v.hoverText = false, ""
			v.linkFrom, v.linkTo = codePos{}, codePos{}
		}
	}
}

func abs32(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

// columnSelect makes one selection per row between two points, spanning the
// columns between their x positions; short rows get what they have.
func (v *CodeEditorView) columnSelect(a, b f32.Point) {
	m := v.metrics
	ra := int((a.Y + v.scrollY) / float32(max(1, m.lh)))
	rb := int((b.Y + v.scrollY) / float32(max(1, m.lh)))
	ra, rb = min(max(ra, 0), v.rowCount()-1), min(max(rb, 0), v.rowCount()-1)
	step := 1
	if rb < ra {
		step = -1
	}
	var sels []codeSel
	for r := ra; ; r += step {
		line := v.lineOf(r)
		pa := v.posAt(f32.Pt(a.X, float32(r*m.lh)-v.scrollY))
		pb := v.posAt(f32.Pt(b.X, float32(r*m.lh)-v.scrollY))
		sels = append(sels, codeSel{codePos{line, pa.col}, codePos{line, pb.col}})
		if r == rb {
			break
		}
	}
	v.sels, v.prim = sels, len(sels)-1
	v.normalize()
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
	if p.col >= len(v.buf.line(p.line)) {
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
	arrows := key.ModShortcutAlt | key.ModShift | key.ModShortcut | key.ModAlt | key.ModCtrl
	filters := []event.Filter{
		key.FocusFilter{Target: v},
		transfer.TargetFilter{Target: v, Type: "application/text"},
		key.Filter{Focus: v, Name: key.NameReturn, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameEnter, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameEscape},
		key.Filter{Focus: v, Name: key.NameDeleteBackward, Optional: key.ModShortcutAlt | key.ModShift | key.ModShortcut},
		key.Filter{Focus: v, Name: key.NameDeleteForward, Optional: key.ModShortcutAlt | key.ModShift},
		key.Filter{Focus: v, Name: key.NameHome, Optional: key.ModShortcut | key.ModShift},
		key.Filter{Focus: v, Name: key.NameEnd, Optional: key.ModShortcut | key.ModShift},
		key.Filter{Focus: v, Name: key.NamePageUp, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NamePageDown, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameLeftArrow, Optional: arrows},
		key.Filter{Focus: v, Name: key.NameRightArrow, Optional: arrows},
		key.Filter{Focus: v, Name: key.NameUpArrow, Optional: arrows},
		key.Filter{Focus: v, Name: key.NameDownArrow, Optional: arrows},
		key.Filter{Focus: v, Name: key.NameF3, Optional: key.ModShift},
		key.Filter{Focus: v, Name: key.NameF12},
		key.Filter{Focus: v, Name: "A", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "C", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "D", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "F", Required: key.ModShortcut, Optional: key.ModAlt},
		key.Filter{Focus: v, Name: "G", Required: key.ModShortcut, Optional: key.ModShift},
		key.Filter{Focus: v, Name: "H", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "X", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "V", Required: key.ModShortcut},
		key.Filter{Focus: v, Name: "Z", Required: key.ModShortcut, Optional: key.ModShift},
		key.Filter{Focus: v, Name: "[", Required: key.ModShortcut | key.ModAlt},
		key.Filter{Focus: v, Name: "]", Required: key.ModShortcut | key.ModAlt},
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
			v.imeText = "" // resend the caret's line
		case key.SelectionEvent:
			c := v.primary().caret
			n := len(v.buf.line(c.line))
			v.sels, v.prim = []codeSel{{codePos{c.line, min(max(e.Start, 0), n)}, codePos{c.line, min(max(e.End, 0), n)}}}, 0
		case transfer.DataEvent:
			data, err := io.ReadAll(e.Open())
			if err == nil && !v.readOnly {
				v.paste(gtx, string(data))
			}
		}
	}
}

// input applies typed text. The input method works in columns of the
// primary caret's line, as reportIME describes it; with several carets or a
// selection across lines, the text goes in at every selection.
func (v *CodeEditorView) input(gtx core.C, e key.EditEvent) {
	if v.readOnly {
		return
	}
	if p := v.primary(); len(v.sels) == 1 && p.anchor.line == p.caret.line {
		n := len(v.buf.line(p.caret.line))
		a, b := min(max(e.Range.Start, 0), n), min(max(e.Range.End, 0), n)
		if a > b {
			a, b = b, a
		}
		v.sels[0] = codeSel{codePos{p.caret.line, a}, codePos{p.caret.line, b}}
	}
	reps := make([]codeReplace, 0, len(v.sels))
	r, size := utf8.DecodeRuneInString(e.Text)
	single := size == len(e.Text) && size > 0
	for _, s := range v.sels {
		reps = append(reps, v.typed(s, e.Text, r, single))
	}
	v.applyReplaces(reps, true)
	v.changed(gtx)
	if last, _ := utf8.DecodeLastRuneInString(e.Text); v.onComplete != nil && len(e.Text) > 0 && len(v.sels) == 1 {
		if isIdent(last) || last == '.' {
			v.complete(gtx, false)
		} else {
			v.comp = nil
		}
	}
}

// replaceEach replaces every selection with text.
func (v *CodeEditorView) replaceEach(gtx core.C, text string) {
	reps := make([]codeReplace, len(v.sels))
	for i, s := range v.sels {
		from, to := s.span()
		reps[i] = codeReplace{from: from, to: to, text: text, caret: -1}
	}
	v.applyReplaces(reps, false)
	v.changed(gtx)
}

// paste puts clipboard text at every selection; with as many lines as
// carets, one line each.
func (v *CodeEditorView) paste(gtx core.C, text string) {
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(v.sels) > 1 && len(lines) == len(v.sels) {
		reps := make([]codeReplace, len(v.sels))
		for i, s := range v.sels {
			from, to := s.span()
			reps[i] = codeReplace{from: from, to: to, text: lines[i], caret: -1}
		}
		v.applyReplaces(reps, false)
		v.changed(gtx)
		return
	}
	v.replaceEach(gtx, text)
}

// changed runs after an edit from the widget: scroll to the caret and tell
// OnChange inside core.Call, which redraws.
func (v *CodeEditorView) changed(gtx core.C) {
	v.afterEdit()
	if v.onChange != nil {
		fn, text := v.onChange, v.buf.text()
		core.Call(gtx, func() { fn(text) })
	}
}

// changedNoCall is changed for edits made from an element callback, which
// already redraws.
func (v *CodeEditorView) changedNoCall() {
	v.afterEdit()
	if v.onChange != nil {
		v.onChange(v.buf.text())
	}
}

func (v *CodeEditorView) afterEdit() {
	v.reveal, v.goalX, v.tabExits, v.hoverText = true, -1, false, ""
	v.keepCaretsVisible()
}

// moveAll moves every caret; with extend, anchors stay.
func (v *CodeEditorView) moveAll(extend bool, fn func(s codeSel) codePos) {
	for i, s := range v.sels {
		c := v.buf.clamp(fn(s))
		v.sels[i].caret = c
		if !extend {
			v.sels[i].anchor = c
		}
	}
	v.normalize()
	v.reveal = true
}

func (v *CodeEditorView) command(gtx core.C, e key.Event) {
	shift := e.Modifiers.Contain(key.ModShift)
	word := e.Modifiers.Contain(key.ModShortcutAlt)
	shortcut := e.Modifiers.Contain(key.ModShortcut)
	alt := e.Modifiers.Contain(key.ModAlt)
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
	switch e.Name {
	case key.NameEscape:
		switch {
		case len(v.sels) > 1:
			p := v.primary()
			v.sels, v.prim = []codeSel{{p.caret, p.caret}}, 0
		case v.search.open || v.search.active:
			v.CloseSearch()
		default:
			p := v.primary()
			v.sels[v.prim].anchor = p.caret
			v.tabExits = true
		}
	case key.NameLeftArrow, key.NameRightArrow:
		dir := 1
		if e.Name == key.NameLeftArrow {
			dir = -1
		}
		v.moveAll(shift, func(s codeSel) codePos {
			from, to := s.span()
			switch {
			case shortcut && !alt:
				if dir < 0 {
					return codePos{s.caret.line, 0}
				}
				return codePos{s.caret.line, len(v.buf.line(s.caret.line))}
			case !shift && !s.empty():
				if dir < 0 {
					return from
				}
				return to
			case word:
				return v.wordStep(s.caret, dir)
			}
			return v.step(s.caret, dir)
		})
		v.goalX = -1
	case key.NameUpArrow, key.NameDownArrow, key.NamePageUp, key.NamePageDown:
		rows := map[key.Name]int{key.NameUpArrow: -1, key.NameDownArrow: 1, key.NamePageUp: -v.visible, key.NamePageDown: v.visible}[e.Name]
		switch {
		case (shortcut && alt || alt && shift) && (e.Name == key.NameUpArrow || e.Name == key.NameDownArrow):
			v.addCaretVertical(rows)
			return
		case shortcut:
			if rows < 0 {
				v.moveAll(shift, func(codeSel) codePos { return codePos{} })
			} else {
				last := v.buf.count() - 1
				v.moveAll(shift, func(codeSel) codePos { return codePos{last, len(v.buf.line(last))} })
			}
			return
		}
		goal := v.goalX
		if goal < 0 {
			c := v.primary().caret
			goal = v.colX(c.line)[c.col]
		}
		prim := v.prim
		for i, s := range v.sels {
			x := v.colX(s.caret.line)[s.caret.col]
			if i == prim {
				x = goal
			}
			c := v.vertical(s.caret, rows, x)
			v.sels[i].caret = c
			if !shift {
				v.sels[i].anchor = c
			}
		}
		v.normalize()
		v.reveal, v.goalX = true, goal
	case key.NameHome:
		v.moveAll(shift, func(s codeSel) codePos {
			if shortcut {
				return codePos{}
			}
			ind := len([]rune(v.buf.indent(s.caret.line)))
			if s.caret.col == ind {
				ind = 0
			}
			return codePos{s.caret.line, ind}
		})
		v.goalX = -1
	case key.NameEnd:
		v.moveAll(shift, func(s codeSel) codePos {
			if shortcut {
				last := v.buf.count() - 1
				return codePos{last, len(v.buf.line(last))}
			}
			return codePos{s.caret.line, len(v.buf.line(s.caret.line))}
		})
		v.goalX = -1
	case key.NameDeleteBackward, key.NameDeleteForward:
		if v.readOnly {
			return
		}
		dir := 1
		if e.Name == key.NameDeleteBackward {
			dir = -1
		}
		reps := make([]codeReplace, 0, len(v.sels))
		for _, s := range v.sels {
			from, to := s.span()
			if s.empty() {
				switch {
				case shortcut && dir < 0:
					from = codePos{s.caret.line, 0}
				case word:
					if p := v.wordStep(s.caret, dir); dir < 0 {
						from = p
					} else {
						to = p
					}
				case dir < 0:
					from = v.step(s.caret, -1)
					// An empty pair goes as one.
					if a, b, ok := v.emptyPair(s.caret); ok {
						from, to = a, b
					}
				default:
					to = v.step(s.caret, 1)
				}
			}
			reps = append(reps, codeReplace{from: from, to: to, caret: -1})
		}
		v.applyReplaces(reps, false)
		v.changed(gtx)
	case key.NameReturn, key.NameEnter:
		if v.readOnly {
			return
		}
		reps := make([]codeReplace, 0, len(v.sels))
		for _, s := range v.sels {
			reps = append(reps, v.newline(s))
		}
		v.applyReplaces(reps, false)
		v.changed(gtx)
	case key.NameTab:
		if v.readOnly {
			return
		}
		multi := shift
		for _, s := range v.sels {
			if from, to := s.span(); from.line != to.line {
				multi = true
			}
		}
		if multi {
			v.shiftLines(gtx, !shift)
			return
		}
		v.replaceEach(gtx, v.indentUnit())
	case key.NameSpace: // Ctrl+Space
		v.complete(gtx, true)
	case key.NameF12:
		if v.onDefinition != nil {
			fn := v.onDefinition
			w, _ := v.buf.wordAt(v.primary().caret)
			core.Call(gtx, func() { fn(w.line, w.col) })
		}
	case key.NameF3, "G":
		dir := 1
		if shift {
			dir = -1
		}
		if !v.search.open || v.search.query == "" {
			v.OpenSearch(false)
		}
		v.findNext(dir)
	case "F":
		v.OpenSearch(alt) // Cmd+Alt+F on macOS opens replace
	case "H":
		v.OpenSearch(true)
	case "A":
		last := v.buf.count() - 1
		v.sels, v.prim = []codeSel{{codePos{}, codePos{last, len(v.buf.line(last))}}}, 0
	case "D":
		v.selectNextOccurrence()
	case "[", "]":
		line := v.primary().caret.line
		if e.Name == "[" {
			if !v.folds[line] {
				v.toggleFold(line)
			}
		} else {
			for start := range v.folds {
				if line >= start && line <= v.foldEnd(start) {
					delete(v.folds, start)
				}
			}
			v.invalidateRows()
		}
	case "C", "X":
		var parts []string
		var whole []int
		for _, s := range v.sels {
			from, to := s.span()
			if s.empty() {
				// Without a selection, copy and cut take the line.
				parts = append(parts, string(v.buf.line(from.line))+"\n")
				whole = append(whole, from.line)
				continue
			}
			parts = append(parts, v.buf.slice(from, to))
		}
		text := strings.Join(parts, "\n")
		if len(whole) == len(v.sels) {
			text = strings.Join(parts, "")
		}
		gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
		if e.Name == "X" && !v.readOnly {
			reps := make([]codeReplace, 0, len(v.sels))
			for _, s := range v.sels {
				from, to := s.span()
				if s.empty() {
					from, to = codePos{from.line, 0}, v.buf.clamp(codePos{from.line + 1, 0})
					if from.line == v.buf.count()-1 {
						to = codePos{from.line, len(v.buf.line(from.line))}
					}
				}
				reps = append(reps, codeReplace{from: from, to: to, caret: -1})
			}
			v.applyReplaces(reps, false)
			v.changed(gtx)
		}
	case "V":
		if !v.readOnly {
			gtx.Execute(clipboard.ReadCmd{Tag: v})
		}
	case "Z":
		if v.readOnly {
			return
		}
		var sels []codeSel
		var ok bool
		if shift {
			sels, ok = v.buf.redoStep()
		} else {
			sels, ok = v.buf.undoStep()
		}
		if ok {
			v.sels, v.prim = append([]codeSel(nil), sels...), max(0, len(sels)-1)
			v.normalize()
			v.changed(gtx)
		}
	}
}

// vertical moves a caret by display rows, to the column nearest x.
func (v *CodeEditorView) vertical(c codePos, rows, x int) codePos {
	row := min(max(v.rowOf(c.line)+rows, 0), v.rowCount()-1)
	line := v.lineOf(row)
	xs := v.colX(line)
	col := 0
	for col < len(xs)-1 && x > (xs[col]+xs[col+1])/2 {
		col++
	}
	return codePos{line, col}
}

// addCaretVertical adds a caret above or below the outermost caret in that
// direction, at the primary caret's x.
func (v *CodeEditorView) addCaretVertical(dir int) {
	x := v.goalX
	if x < 0 {
		c := v.primary().caret
		x = v.colX(c.line)[c.col]
	}
	edge := v.sels[0].caret
	if dir > 0 {
		edge = v.sels[len(v.sels)-1].caret
	}
	p := v.vertical(edge, dir, x)
	if p.line == edge.line {
		return
	}
	v.sels = append(v.sels, codeSel{p, p})
	v.prim = len(v.sels) - 1
	v.normalize()
	v.goalX, v.reveal = x, true
}

// selectNextOccurrence selects the word at the caret, or adds a selection
// at the next occurrence of the primary selection's text.
func (v *CodeEditorView) selectNextOccurrence() {
	p := v.primary()
	if p.empty() {
		a, b := v.buf.wordAt(p.caret)
		if a != b {
			v.sels[v.prim] = codeSel{a, b}
		}
		return
	}
	from, to := p.span()
	if from.line != to.line {
		return
	}
	needle := v.buf.line(from.line)[from.col:to.col]
	last := v.sels[len(v.sels)-1].caret
	_, last = ordered(v.sels[len(v.sels)-1].anchor, last)
	found := codeRange{codePos{-1, 0}, codePos{}}
	search := func(startLine, startCol, endLine int) bool {
		hit := false
		v.buf.lines.each(startLine, func(i int, l *codeLine) bool {
			c0 := 0
			if i == startLine {
				c0 = startCol
			}
			for c := c0; c+len(needle) <= len(l.text); c++ {
				if slices.Equal(l.text[c:c+len(needle)], needle) {
					r := codeRange{codePos{i, c}, codePos{i, c + len(needle)}}
					taken := false
					for _, s := range v.sels {
						if f, _ := s.span(); f == r.from {
							taken = true
						}
					}
					if !taken {
						found, hit = r, true
						return false
					}
				}
			}
			return i < endLine
		})
		return hit
	}
	if !search(last.line, last.col, v.buf.count()-1) {
		search(0, 0, last.line)
	}
	if found.from.line < 0 {
		return
	}
	v.sels = append(v.sels, codeSel{found.from, found.to})
	v.prim = len(v.sels) - 1
	v.normalize()
	v.reveal = true
	v.keepCaretsVisible()
}

// step moves one rune, across line ends.
func (v *CodeEditorView) step(p codePos, dir int) codePos {
	p.col += dir
	if p.col < 0 && p.line > 0 {
		return codePos{p.line - 1, len(v.buf.line(p.line - 1))}
	}
	if p.col > len(v.buf.line(p.line)) && p.line < v.buf.count()-1 {
		return codePos{p.line + 1, 0}
	}
	return v.buf.clamp(p)
}

// wordStep moves to the next word boundary in a direction.
func (v *CodeEditorView) wordStep(p codePos, dir int) codePos {
	l := v.buf.line(p.line)
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

// indentUnit is what Tab inserts: TabSize's choice, or a tab in files
// indented with tabs and spaces otherwise.
func (v *CodeEditorView) indentUnit() string {
	switch v.hardTabs {
	case 1:
		return "\t"
	case 2:
		return strings.Repeat(" ", v.tabSize)
	}
	c := v.primary().caret.line
	hard := false
	v.buf.lines.each(max(0, c-50), func(i int, l *codeLine) bool {
		if len(l.text) > 0 && l.text[0] == '\t' {
			hard = true
			return false
		}
		return i < c+50
	})
	if hard {
		return "\t"
	}
	return strings.Repeat(" ", v.tabSize)
}

// shiftLines indents or outdents every line touched by a selection, as one
// undo step. A selection ending at a line's start leaves that line alone.
func (v *CodeEditorView) shiftLines(gtx core.C, in bool) {
	unit := v.indentUnit()
	seen := map[int]bool{}
	var lines []int
	for _, s := range v.sels {
		from, to := s.span()
		last := to.line
		if to.col == 0 && to.line > from.line {
			last--
		}
		for l := from.line; l <= last; l++ {
			if !seen[l] {
				seen[l] = true
				lines = append(lines, l)
			}
		}
	}
	slices.Sort(lines)
	before := append([]codeSel(nil), v.sels...)
	v.buf.begin(before)
	for _, l := range lines {
		text := v.buf.line(l)
		switch {
		case in:
			if len(text) > 0 {
				v.buf.edit(codePos{l, 0}, codePos{l, 0}, unit)
			}
		case len(text) > 0 && text[0] == '\t':
			v.buf.edit(codePos{l, 0}, codePos{l, 1}, "")
		default:
			n := 0
			for n < len(text) && n < v.tabSize && text[n] == ' ' {
				n++
			}
			if n > 0 {
				v.buf.edit(codePos{l, 0}, codePos{l, n}, "")
			}
		}
	}
	// Selections keep covering whole lines.
	for i, s := range v.sels {
		from, to := s.span()
		last := to
		if to.col != 0 || to.line == from.line {
			last = codePos{to.line, len(v.buf.line(to.line))}
		}
		v.sels[i] = codeSel{codePos{from.line, 0}, last}
	}
	v.buf.commit(v.sels, false)
	if len(lines) > 0 {
		v.buf.highlightNear(v.lang, codeStyleName(), lines[0], lines[len(lines)-1])
	}
	v.changed(gtx)
}

// complete asks OnComplete for the word before the primary caret.
func (v *CodeEditorView) complete(gtx core.C, explicit bool) {
	if v.onComplete == nil {
		return
	}
	c := v.primary().caret
	start, _ := v.buf.wordAt(c)
	start.col = min(start.col, c.col)
	prefix := string(v.buf.line(c.line)[start.col:c.col])
	if prefix == "" && !explicit {
		v.comp = nil
		return
	}
	fn := v.onComplete
	var items []CodeCompletion
	core.Call(gtx, func() { items = fn(c.line, c.col, prefix) })
	v.comp, v.compSel, v.compFrom = items, 0, codePos{c.line, start.col}
}

func (v *CodeEditorView) acceptCompletion(gtx core.C) {
	c := v.comp[v.compSel]
	v.comp = nil
	insert := c.Insert
	if insert == "" {
		insert = c.Label
	}
	v.sels, v.prim = []codeSel{{v.compFrom, v.primary().caret}}, 0
	v.replaceEach(gtx, insert)
}

// reportIME tells the platform input method the primary caret's line, the
// selection in it and where the caret is, so candidates appear beside it.
func (v *CodeEditorView) reportIME(gtx core.C) {
	if !v.focused {
		return
	}
	p := v.primary()
	line := string(v.buf.line(p.caret.line))
	if v.imeLine != p.caret.line || v.imeText != line {
		v.imeLine, v.imeText = p.caret.line, line
		gtx.Execute(key.SnippetCmd{Tag: v, Snippet: key.Snippet{Range: key.Range{Start: 0, End: len(v.buf.line(p.caret.line))}, Text: line}})
	}
	rng := key.Range{Start: p.caret.col, End: p.caret.col}
	if p.anchor.line == p.caret.line {
		rng = key.Range{Start: p.anchor.col, End: p.caret.col}
	}
	pt := v.caretPoint(p.caret).Add(image.Pt(0, v.metrics.baseline))
	if rng != v.imeRange || pt != v.imeCaret {
		v.imeRange, v.imeCaret = rng, pt
		gtx.Execute(key.SelectionCmd{Tag: v, Range: rng, Caret: key.Caret{Pos: layout.FPt(pt), Ascent: float32(v.metrics.ascent), Descent: float32(v.metrics.descent)}})
	}
}
