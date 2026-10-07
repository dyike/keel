package kit

import (
	"image"
	"image/color"
	"sort"
	"strconv"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	gtext "gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/image/math/fixed"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// codeMetrics is the editor's geometry for one frame, in px.
type codeMetrics struct {
	px         int // text size
	lh         int // row height
	baseline   int // from a row's top
	ascent     int
	descent    int
	foldW      int // width of the fold arrows column, at the gutter's left
	gutter     int // width of fold arrows and line numbers
	pad        int // space between gutter and text
	tab        int // width of a tab stop
	space      int // width of a space
	size       image.Point
	textOrigin int // x of column 0 before scrolling
}

const codeHighlightDelay = 150 * time.Millisecond

func (v *CodeEditorView) face() font.Font { return font.Font{Typeface: theme.MonoFace} }

// shape lays out one line of text without wrapping.
func (v *CodeEditorView) shape(px int, s string) []gtext.Glyph {
	return v.shapeFace(px, v.face(), s)
}

func (v *CodeEditorView) shapeFace(px int, face font.Font, s string) []gtext.Glyph {
	sh := theme.Material.Shaper
	sh.LayoutString(gtext.Parameters{Font: face, PxPerEm: fixed.I(px), MaxWidth: 1 << 24}, s)
	var out []gtext.Glyph
	for {
		g, ok := sh.NextGlyph()
		if !ok {
			return out
		}
		if g.Flags&gtext.FlagParagraphBreak != 0 && g.Runes == 0 {
			continue
		}
		out = append(out, g)
	}
}

func (v *CodeEditorView) textWidth(s string) int {
	w := 0
	for _, g := range v.shape(v.metrics.px, s) {
		w += g.Advance.Round()
	}
	return w
}

func (v *CodeEditorView) measure(gtx core.C) {
	m := &v.metrics
	px := gtx.Sp(theme.TextMd)
	if m.px != px {
		clear(v.xs)
	}
	m.px = px
	m.lh = px * 3 / 2
	m.ascent, m.descent = px, px/4
	for _, g := range v.shape(px, "Mg") {
		m.ascent, m.descent = g.Ascent.Ceil(), g.Descent.Ceil()
		break
	}
	m.baseline = (m.lh-(m.ascent+m.descent))/2 + m.ascent
	m.space = max(1, v.textWidth(" "))
	if tab := v.tabSize * m.space; tab != m.tab {
		m.tab = tab
		clear(v.xs)
	}
	digit := v.textWidth("0")
	m.pad = gtx.Dp(12)
	m.foldW = gtx.Dp(16)
	m.gutter = m.foldW + len(strconv.Itoa(v.buf.count()))*digit + gtx.Dp(14)
	m.size = gtx.Constraints.Max
	m.textOrigin = m.gutter + m.pad
	v.visible = max(1, m.size.Y/max(1, m.lh))
}

// colX returns the x of every column boundary of a line (len+1 values),
// expanding tabs to stops. Cached by line text.
func (v *CodeEditorView) colX(line int) []int {
	l := v.buf.line(line)
	key := string(l)
	if xs, ok := v.xs[key]; ok {
		return xs
	}
	if len(v.xs) > 4096 {
		clear(v.xs)
	}
	xs := make([]int, len(l)+1)
	x := fixed.I(0)
	i := 0
	for i < len(l) {
		if l[i] == '\t' {
			stop := (x.Round()/max(1, v.metrics.tab) + 1) * v.metrics.tab
			x = fixed.I(stop)
			i++
			xs[i] = x.Round()
			continue
		}
		j := i
		for j < len(l) && l[j] != '\t' {
			j++
		}
		k := i
		for _, g := range v.shape(v.metrics.px, string(l[i:j])) {
			n := max(int(g.Runes), 1)
			for r := 1; r <= n && k < j; r++ {
				k++
				xs[k] = (x + g.Advance*fixed.Int26_6(r)/fixed.Int26_6(n)).Round()
			}
			x += g.Advance
		}
		for k < j { // runes the shaper merged away
			k++
			xs[k] = x.Round()
		}
		i = j
	}
	v.xs[key] = xs
	return xs
}

// posAt maps a point in the editor to the nearest text position.
func (v *CodeEditorView) posAt(p f32.Point) codePos {
	m := v.metrics
	vr := v.vrow(int((p.Y + v.scrollY) / float32(max(1, m.lh))))
	return codePos{vr.line, v.colIn(vr, int(p.X+v.scrollX)-m.textOrigin)}
}

// caretPoint is a position's top-left in editor coordinates.
func (v *CodeEditorView) caretPoint(p codePos) image.Point {
	m := v.metrics
	x := m.textOrigin + v.rowX(p) - int(v.scrollX)
	return image.Pt(x, v.vrowOf(p)*m.lh-int(v.scrollY))
}

func (v *CodeEditorView) layout(gtx core.C) core.D {
	v.measure(gtx)
	m := v.metrics
	if v.wantFocus && gtx.Enabled() {
		gtx.Execute(key.FocusCmd{Tag: v})
		v.wantFocus = false
	}
	v.update(gtx)
	v.focused = gtx.Focused(v) && gtx.Enabled()
	v.highlight()
	if v.reveal {
		v.reveal = false
		v.scrollToCaret()
	}
	v.clampScroll()
	v.refreshMatches()

	area := clip.Rect{Max: m.size}.Push(gtx.Ops)
	defer area.Pop()
	editor := clip.Rect{Max: m.size}.Push(gtx.Ops)
	v.semantics(gtx)
	if v.linkTo != v.linkFrom {
		pointer.CursorPointer.Add(gtx.Ops)
	} else {
		pointer.CursorText.Add(gtx.Ops)
	}
	event.Op(gtx.Ops, v)
	key.InputHintOp{Tag: v, Hint: key.HintAny}.Add(gtx.Ops)
	v.scroll.Add(gtx.Ops)
	v.scrollH.Add(gtx.Ops)

	paint.FillShape(gtx.Ops, theme.CodeBg, clip.Rect{Max: m.size}.Op())
	first := max(0, int(v.scrollY)/max(1, m.lh))
	last := min(v.vrowCount()-1, first+v.visible+1)
	for row := first; row <= last; row++ {
		v.paintLine(gtx, row)
	}
	// The gutter covers text scrolled under it.
	paint.FillShape(gtx.Ops, theme.CodeBg, clip.Rect{Max: image.Pt(m.gutter, m.size.Y)}.Op())
	paint.FillShape(gtx.Ops, theme.Border, clip.Rect{Min: image.Pt(m.gutter-1, 0), Max: image.Pt(m.gutter, m.size.Y)}.Op())
	for row := first; row <= last; row++ {
		v.paintGutter(gtx, row)
	}
	v.paintCarets(gtx)
	v.paintScrollbar(gtx)
	editor.Pop() // Floating controls are siblings of the textbox in the semantic tree.
	v.paintCompletions(gtx)
	v.paintHover(gtx)
	v.reportIME(gtx)
	return core.D{Size: m.size}
}

// selectionOn returns the column range each selection covers on a line;
// spill marks a selected line end.
func (v *CodeEditorView) selectionOn(line int, fn func(a, b int, spill bool)) {
	n := len(v.buf.line(line))
	for _, s := range v.sels {
		from, to := s.span()
		if s.empty() || line < from.line || line > to.line {
			continue
		}
		a, b := 0, n
		if line == from.line {
			a = from.col
		}
		if line == to.line {
			b = to.col
		}
		fn(a, b, line < to.line)
	}
}

func (v *CodeEditorView) paintLine(gtx core.C, row int) {
	m := v.metrics
	vr := v.vrow(row)
	line := vr.line
	top := row*m.lh - int(v.scrollY)
	xs := v.colX(line)
	ox := m.textOrigin - int(v.scrollX) - xs[vr.start]
	last := vr.end == len(xs)-1 // the row holding the line's end
	rect := func(a, b int) image.Rectangle {
		// Ranges from diagnostics or a stale pointer may pass the line's end;
		// a wrapped row shows only its own columns.
		a, b = min(max(a, vr.start), vr.end), min(max(b, vr.start), vr.end)
		return image.Rect(ox+xs[a], top, ox+xs[b], top+m.lh)
	}
	if v.focused && len(v.sels) == 1 && v.primary().empty() && v.primary().caret.line == line {
		c := theme.Subtle
		c.A /= 2
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Pt(m.gutter, top), Max: image.Pt(m.size.X, top+m.lh)}.Op())
	}
	decorations := v.lineDecorations(line)
	v.paintDecorations(gtx, row, m.textOrigin-int(v.scrollX), top, decorations)
	// Find results, the chosen one stronger.
	if (v.search.open || v.search.active) && len(v.search.matches) > 0 {
		ms := v.search.matches
		// Matches do not overlap, so their ends ascend with their starts.
		i := sort.Search(len(ms), func(i int) bool { return ms[i].to.line >= line })
		for ; i < len(ms) && ms[i].from.line <= line; i++ {
			c := theme.Warning
			c.A = 0x40
			if i == v.search.current {
				c.A = 0x90
			}
			a, b := 0, len(xs)-1
			if ms[i].from.line == line {
				a = ms[i].from.col
			}
			if ms[i].to.line == line {
				b = ms[i].to.col
			}
			r := rect(a, b)
			if ms[i].to.line > line && last {
				r.Max.X += m.space // the line break is part of the match
			}
			if r.Dx() > 0 {
				paint.FillShape(gtx.Ops, c, clip.UniformRRect(r, gtx.Dp(2)).Op(gtx.Ops))
			}
		}
	}
	v.selectionOn(line, func(a, b int, spill bool) {
		r := rect(a, b)
		if spill && last {
			r.Max.X += m.space
		} else if b > vr.end && a < vr.end && !last {
			r.Max.X = m.size.X // selection continues on the next row
		}
		paint.FillShape(gtx.Ops, theme.Highlight, clip.Rect(r).Op())
	})
	// Text, a piece per color and between tabs.
	l := v.buf.line(line)
	spans := v.buf.spans(line)
	si := 0
	for i := vr.start; i < vr.end; {
		if l[i] == '\t' {
			i++
			continue
		}
		for si < len(spans) && spans[si].end <= i {
			si++
		}
		c := theme.CodeText
		end := vr.end
		if si < len(spans) && spans[si].start <= i {
			end = min(end, spans[si].end)
			if spans[si].color.A != 0 {
				c = spans[si].color
			}
		} else if si < len(spans) {
			end = min(end, spans[si].start)
		}
		face := v.face()
		for _, d := range decorations {
			if d.style != CodeDecorationText {
				continue
			}
			if d.a > i {
				end = min(end, d.a)
			} else if d.b > i {
				end = min(end, d.b)
				if !d.keep {
					c = d.color
				}
				if d.weight != 0 {
					face.Weight = d.weight
				}
				if d.italic {
					face.Style = font.Italic
				}
			}
		}
		j := i
		for j < end && l[j] != '\t' {
			j++
		}
		if x := ox + xs[i]; x < m.size.X && ox+xs[j] > m.gutter {
			v.paintTextFace(gtx, face, string(l[i:j]), image.Pt(x, top+m.baseline), c)
		}
		i = j
	}
	if v.whitespace {
		v.paintWhitespace(gtx, l, xs, ox, top, vr.start, vr.end)
	}
	// A folded region shows as a pill after its header.
	if v.folds[line] && last {
		x := ox + xs[len(xs)-1] + m.space
		r := image.Rect(x, top+m.lh/6, x+v.textWidth("…")+gtx.Dp(12), top+m.lh-m.lh/6)
		paint.FillShape(gtx.Ops, theme.Subtle, clip.UniformRRect(r, r.Dy()/2).Op(gtx.Ops))
		v.paintText(gtx, "…", image.Pt(r.Min.X+gtx.Dp(6), top+m.baseline), theme.Muted)
	}
	// The word a Cmd/Ctrl+click would look up.
	if v.linkFrom != v.linkTo && v.linkFrom.line == line && v.linkFrom.line < v.buf.count() && v.linkTo.col > vr.start && v.linkFrom.col < vr.end+1 {
		r := rect(v.linkFrom.col, v.linkTo.col)
		paint.FillShape(gtx.Ops, theme.Primary, clip.Rect{Min: image.Pt(r.Min.X, r.Max.Y-gtx.Dp(3)), Max: image.Pt(r.Max.X, r.Max.Y-gtx.Dp(2))}.Op())
	}
	// Diagnostics: a wavy line under the range.
	for _, d := range v.diagnostics {
		if line < d.Line || line > d.EndLine {
			continue
		}
		a, b := 0, len(xs)-1
		if line == d.Line {
			a = min(max(d.Col, 0), len(xs)-1)
		}
		if line == d.EndLine {
			b = min(max(d.EndCol, a), len(xs)-1)
		}
		if a == b && b < len(xs)-1 {
			b++
		}
		if b <= vr.start && vr.start > 0 || a >= vr.end && !last {
			continue // on another row of the wrapped line
		}
		a, b = max(a, vr.start), min(b, vr.end)
		x0, x1 := ox+xs[a], max(ox+xs[b], ox+xs[a]+m.px/2)
		squiggle(gtx, x0, x1, top+m.lh-gtx.Dp(3), gtx.Dp(2), codeSeverityColor(d.Severity))
	}
}

// paintWhitespace marks spaces with dots and tabs with arrows.
func (v *CodeEditorView) paintWhitespace(gtx core.C, l []rune, xs []int, ox, top, from, to int) {
	m := v.metrics
	c := theme.Muted
	c.A = 0x70
	y := top + m.lh/2
	for i := from; i < to; i++ {
		r := l[i]
		x0, x1 := ox+xs[i], ox+xs[i+1]
		if x1 < m.gutter || x0 > m.size.X {
			continue
		}
		switch r {
		case ' ':
			d := max(1, gtx.Dp(1))
			cx := (x0 + x1) / 2
			paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Pt(cx-d, y-d), Max: image.Pt(cx+d, y+d)}.Op())
		case '\t':
			if x1-x0 < gtx.Dp(6) {
				continue
			}
			var p clip.Path
			p.Begin(gtx.Ops)
			a, b := float32(x0+gtx.Dp(2)), float32(x1-gtx.Dp(3))
			h := float32(gtx.Dp(3))
			p.MoveTo(f32.Pt(a, float32(y)))
			p.LineTo(f32.Pt(b, float32(y)))
			p.MoveTo(f32.Pt(b-h, float32(y)-h))
			p.LineTo(f32.Pt(b, float32(y)))
			p.LineTo(f32.Pt(b-h, float32(y)+h))
			paint.FillShape(gtx.Ops, c, clip.Stroke{Path: p.End(), Width: float32(max(1, gtx.Dp(1)))}.Op())
		}
	}
}

func (v *CodeEditorView) paintText(gtx core.C, s string, at image.Point, c color.NRGBA) {
	v.paintTextFace(gtx, v.face(), s, at, c)
}

func (v *CodeEditorView) paintTextFace(gtx core.C, face font.Font, s string, at image.Point, c color.NRGBA) {
	glyphs := v.shapeFace(v.metrics.px, face, s)
	if len(glyphs) == 0 {
		return
	}
	sh := theme.Material.Shaper
	t := op.Affine(f32.AffineId().Offset(f32.Pt(float32(at.X)+float32(glyphs[0].X.Round()), float32(at.Y)))).Push(gtx.Ops)
	outline := clip.Outline{Path: sh.Shape(glyphs)}.Op().Push(gtx.Ops)
	paint.ColorOp{Color: c}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	outline.Pop()
	if call := sh.Bitmaps(glyphs); call != (op.CallOp{}) {
		call.Add(gtx.Ops)
	}
	t.Pop()
}

func (v *CodeEditorView) paintGutter(gtx core.C, row int) {
	m := v.metrics
	vr := v.vrow(row)
	if vr.start > 0 {
		return // a wrapped line's continuation
	}
	line := vr.line
	top := row*m.lh - int(v.scrollY)
	num := strconv.Itoa(line + 1)
	c := theme.Muted
	for _, s := range v.sels {
		if s.caret.line == line {
			c = theme.CodeText
		}
	}
	v.paintText(gtx, num, image.Pt(m.gutter-gtx.Dp(10)-v.textWidth(num), top+m.baseline), c)
	// The fold arrow: ▸ folded, ▾ foldable.
	if v.foldEnd(line) > line {
		cx, cy := float32(m.foldW)/2, float32(top+m.lh/2)
		s := float32(gtx.Dp(3))
		var p clip.Path
		p.Begin(gtx.Ops)
		if v.folds[line] {
			p.MoveTo(f32.Pt(cx-s/2, cy-s))
			p.LineTo(f32.Pt(cx+s/2, cy))
			p.LineTo(f32.Pt(cx-s/2, cy+s))
		} else {
			p.MoveTo(f32.Pt(cx-s, cy-s/2))
			p.LineTo(f32.Pt(cx, cy+s/2))
			p.LineTo(f32.Pt(cx+s, cy-s/2))
		}
		paint.FillShape(gtx.Ops, theme.Muted, clip.Stroke{Path: p.End(), Width: float32(max(1, gtx.Dp(1.5)))}.Op())
		// Agents fold and unfold through the arrow.
		r := image.Rect(0, top, m.foldW, top+m.lh)
		a := clip.Rect(r).Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		name := locale.Current().FoldRegion
		if v.folds[line] {
			name = locale.Current().UnfoldRegion
		}
		semantic.LabelOp(name + " " + num).Add(gtx.Ops)
		a.Pop()
	}
	// A dot for the most serious diagnostic on the line.
	worst := -1
	for _, d := range v.diagnostics {
		if line >= d.Line && line <= d.EndLine && (worst < 0 || int(d.Severity) < worst) {
			worst = int(d.Severity)
		}
	}
	if worst >= 0 {
		r := gtx.Dp(3)
		cx, cy := m.foldW+gtx.Dp(3), top+m.lh/2
		paint.FillShape(gtx.Ops, codeSeverityColor(CodeSeverity(worst)), clip.Ellipse{Min: image.Pt(cx-r, cy-r), Max: image.Pt(cx+r, cy+r)}.Op(gtx.Ops))
	}
}

func (v *CodeEditorView) paintCarets(gtx core.C) {
	if !v.focused {
		return
	}
	const interval = 500 * time.Millisecond
	elapsed := gtx.Now.Sub(v.blink)
	if elapsed < 10*time.Second {
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(interval - elapsed%interval)})
		if elapsed/interval%2 != 0 {
			return
		}
	}
	m := v.metrics
	w := max(1, gtx.Dp(2))
	inset := (m.lh - m.ascent - m.descent) / 2
	for _, s := range v.sels {
		if v.hidden(s.caret.line) {
			continue
		}
		p := v.caretPoint(s.caret)
		if p.X < m.gutter || p.Y < -m.lh || p.Y > m.size.Y {
			continue
		}
		paint.FillShape(gtx.Ops, theme.Primary, clip.Rect{Min: image.Pt(p.X-w/2, p.Y+inset), Max: image.Pt(p.X-w/2+w, p.Y+m.lh-inset)}.Op())
	}
}

func (v *CodeEditorView) paintScrollbar(gtx core.C) {
	m := v.metrics
	// The editor permits one extra line of trailing space. Use the same
	// extent as clampScroll so the thumb stays in bounds at the bottom.
	total := v.vrowCount()*m.lh + m.lh
	if total <= m.size.Y {
		return
	}
	h := min(m.size.Y, max(gtx.Dp(24), m.size.Y*m.size.Y/total))
	y := min(max(int(v.scrollY), 0), total-m.size.Y) * (m.size.Y - h) / max(1, total-m.size.Y)
	c := theme.Muted
	c.A = 0x60
	w := gtx.Dp(6)
	paint.FillShape(gtx.Ops, c, clip.UniformRRect(image.Rect(m.size.X-w-gtx.Dp(2), y, m.size.X-gtx.Dp(2), y+h), w/2).Op(gtx.Ops))
	// Find results as ticks beside the scrollbar.
	if v.search.open || v.search.active {
		tc := theme.Warning
		for _, mt := range v.search.matches {
			ty := v.vrowOf(mt.from) * m.size.Y / max(1, v.vrowCount())
			paint.FillShape(gtx.Ops, tc, clip.Rect{Min: image.Pt(m.size.X-w-gtx.Dp(2), ty), Max: image.Pt(m.size.X-gtx.Dp(2), ty+max(1, gtx.Dp(2)))}.Op())
		}
	}
}

// completionRect is where the completion list draws, in editor coordinates.
func (v *CodeEditorView) completionRect(gtx core.C) image.Rectangle {
	m := v.metrics
	rows := min(len(v.comp), 8)
	w := gtx.Dp(260)
	for _, c := range v.comp {
		w = max(w, v.textWidth(c.Label)+v.textWidth(c.Detail)+gtx.Dp(40))
	}
	w = min(w, m.size.X)
	p := v.caretPoint(v.compFrom)
	r := image.Rect(p.X, p.Y+m.lh, p.X+w, p.Y+m.lh+rows*m.lh+gtx.Dp(8))
	if r.Max.Y > m.size.Y && p.Y-r.Dy() >= 0 {
		r = r.Sub(image.Pt(0, r.Dy()+m.lh))
	}
	if r.Max.X > m.size.X {
		r = r.Sub(image.Pt(r.Max.X-m.size.X, 0))
	}
	return r
}

// completionTop is the first completion row shown, keeping the choice in view.
func (v *CodeEditorView) completionTop() int { return max(0, min(v.compSel-7, len(v.comp)-8)) }

func (v *CodeEditorView) paintCompletions(gtx core.C) {
	if len(v.comp) == 0 {
		return
	}
	m := v.metrics
	r := v.completionRect(gtx)
	panel(gtx, r, gtx.Dp(theme.RadiusLg))
	top := v.completionTop()
	for i := top; i < min(len(v.comp), top+8); i++ {
		row := image.Rect(r.Min.X+gtx.Dp(4), r.Min.Y+gtx.Dp(4)+(i-top)*m.lh, r.Max.X-gtx.Dp(4), r.Min.Y+gtx.Dp(4)+(i-top+1)*m.lh)
		area := clip.Rect(row).Push(gtx.Ops)
		core.Role("option").Add(gtx.Ops)
		semantic.LabelOp(v.comp[i].Label).Add(gtx.Ops)
		semantic.SelectedOp(i == v.compSel).Add(gtx.Ops)
		if i == v.compSel {
			paint.FillShape(gtx.Ops, theme.Highlight, clip.UniformRRect(row, gtx.Dp(theme.RadiusMd)).Op(gtx.Ops))
		}
		v.paintText(gtx, v.comp[i].Label, image.Pt(row.Min.X+gtx.Dp(8), row.Min.Y+m.baseline), theme.Text)
		if d := v.comp[i].Detail; d != "" {
			left := row.Min.X + gtx.Dp(20) + v.textWidth(v.comp[i].Label)
			if left < row.Max.X-gtx.Dp(8) {
				detail := clip.Rect(image.Rect(left, row.Min.Y, row.Max.X-gtx.Dp(8), row.Max.Y)).Push(gtx.Ops)
				v.paintText(gtx, d, image.Pt(max(left, row.Max.X-gtx.Dp(8)-v.textWidth(d)), row.Min.Y+m.baseline), theme.Muted)
				detail.Pop()
			}
		}
		area.Pop()
	}
}

func (v *CodeEditorView) paintHover(gtx core.C) {
	if v.hoverText == "" {
		return
	}
	m := v.metrics
	w := v.textWidth(v.hoverText)
	p := image.Pt(int(v.hoverAt.X), int(v.hoverAt.Y)+m.lh)
	r := image.Rectangle{Min: p, Max: p.Add(image.Pt(w+gtx.Dp(20), m.lh+gtx.Dp(8)))}
	if r.Max.X > m.size.X {
		r = r.Sub(image.Pt(r.Max.X-m.size.X, 0))
	}
	if r.Max.Y > m.size.Y {
		r = r.Sub(image.Pt(0, r.Dy()+m.lh*2))
	}
	panel(gtx, r, gtx.Dp(theme.RadiusMd))
	area := clip.Rect(r).Push(gtx.Ops)
	core.Role("tooltip").Add(gtx.Ops)
	semantic.LabelOp(v.hoverText).Add(gtx.Ops)
	area.Pop()
	v.paintText(gtx, v.hoverText, image.Pt(r.Min.X+gtx.Dp(10), r.Min.Y+gtx.Dp(4)+m.baseline), theme.Text)
}

// panel draws a floating surface: shadow, background and border.
func panel(gtx core.C, r image.Rectangle, radius int) {
	sh := theme.Shadow
	sh.A /= 4
	for i := 1; i <= 4; i++ {
		g := gtx.Dp(unit.Dp(i * 2))
		paint.FillShape(gtx.Ops, sh, clip.UniformRRect(image.Rect(r.Min.X-g, r.Min.Y-g+gtx.Dp(3), r.Max.X+g, r.Max.Y+g+gtx.Dp(3)), radius+g).Op(gtx.Ops))
	}
	paint.FillShape(gtx.Ops, theme.Surface, clip.UniformRRect(r, radius).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, theme.Border, clip.Stroke{Path: clip.UniformRRect(r, radius).Path(gtx.Ops), Width: 1}.Op())
}

// squiggle draws a wavy line from x0 to x1 at y.
func squiggle(gtx core.C, x0, x1, y, amp int, c color.NRGBA) {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(float32(x0), float32(y)))
	step := float32(max(2, amp*2))
	up := true
	for x := float32(x0) + step; x <= float32(x1)+step/2; x += step {
		dy := float32(amp)
		if up {
			dy = -dy
		}
		p.LineTo(f32.Pt(x, float32(y)+dy))
		up = !up
	}
	paint.FillShape(gtx.Ops, c, clip.Stroke{Path: p.End(), Width: float32(max(1, gtx.Dp(1)))}.Op())
}

func codeSeverityColor(s CodeSeverity) color.NRGBA {
	switch s {
	case CodeSeverityWarning:
		return theme.Warning
	case CodeSeverityInfo:
		return theme.Info
	}
	return theme.Danger
}

// semantics describes the editor to agents as a text box: its name and, for
// files of reasonable size, its text.
func (v *CodeEditorView) semantics(gtx core.C) {
	semantic.Editor.Add(gtx.Ops)
	name := v.name
	if name == "" {
		name = locale.Current().CodeEditor
	}
	semantic.LabelOp(name).Add(gtx.Ops)
	if v.buf.count() <= 2000 {
		semantic.DescriptionOp(v.buf.text()).Add(gtx.Ops)
	} else {
		semantic.DescriptionOp(locale.Current().Rows(v.buf.count())).Add(gtx.Ops)
	}
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
}

// highlight starts a background pass when the text or style changed and none
// is running; its result replaces the spans if the text is still the same.
func (v *CodeEditorView) highlight() {
	style := codeStyleName()
	if v.hlRunning || v.lang == "" || (v.hlRev == v.buf.revision && v.hlStyle == style) {
		return
	}
	v.hlRunning = true
	rev, lang, src := v.buf.revision, v.lang, v.buf.text()
	go func() {
		time.Sleep(codeHighlightDelay)
		spans := highlightCode(lang, src, style)
		core.Update(func() {
			v.hlRunning = false
			if v.buf.revision == rev && len(spans) == v.buf.count() {
				v.buf.lines.each(0, func(i int, l *codeLine) bool {
					l.spans = spans[i]
					return true
				})
				v.hlRev, v.hlStyle = rev, style
			}
		})
	}()
}

func (v *CodeEditorView) clampScroll() {
	m := v.metrics
	maxY := float32(max(0, v.vrowCount()*m.lh-m.size.Y+m.lh))
	v.scrollY = min(max(v.scrollY, 0), maxY)
	v.scrollX = max(v.scrollX, 0)
	if v.wrap {
		v.scrollX = 0
	}
}

// scrollToCaret keeps the primary caret in view.
func (v *CodeEditorView) scrollToCaret() {
	m := v.metrics
	if m.lh == 0 {
		return
	}
	c := v.primary().caret
	y := float32(v.vrowOf(c) * m.lh)
	if y < v.scrollY {
		v.scrollY = y
	} else if bottom := y + float32(m.lh) - float32(m.size.Y); bottom > v.scrollY {
		v.scrollY = bottom
	}
	if v.wrap {
		return
	}
	x := float32(v.colX(c.line)[c.col])
	view := float32(m.size.X - m.textOrigin - m.pad)
	if x < v.scrollX {
		v.scrollX = max(0, x-view/3)
	} else if x > v.scrollX+view {
		v.scrollX = x - view*2/3
	}
}
