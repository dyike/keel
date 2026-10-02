package kit

import (
	"image"
	"image/color"
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
	lh         int // line height
	baseline   int // from a line's top
	ascent     int
	descent    int
	gutter     int // width of the line numbers column
	pad        int // space between gutter and text
	tab        int // width of a tab stop
	size       image.Point
	textOrigin int // x of column 0 before scrolling
}

const codeHighlightDelay = 150 * time.Millisecond

func (v *CodeEditorView) face() font.Font { return font.Font{Typeface: theme.MonoFace} }

// shape lays out one line of text without wrapping.
func (v *CodeEditorView) shape(px int, s string) []gtext.Glyph {
	sh := theme.Material.Shaper
	sh.LayoutString(gtext.Parameters{Font: v.face(), PxPerEm: fixed.I(px), MaxWidth: 1 << 24}, s)
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
	space := 0
	for _, g := range v.shape(px, " ") {
		space += g.Advance.Round()
	}
	m.tab = max(1, 4*space)
	digit := 0
	for _, g := range v.shape(px, "0") {
		digit += g.Advance.Round()
	}
	m.pad = gtx.Dp(12)
	m.gutter = len(strconv.Itoa(len(v.buf.lines)))*digit + gtx.Dp(24)
	m.size = gtx.Constraints.Max
	m.textOrigin = m.gutter + m.pad
	v.visible = max(1, m.size.Y/max(1, m.lh))
}

// colX returns the x of every column boundary of a line (len+1 values),
// expanding tabs to stops. Cached by line text.
func (v *CodeEditorView) colX(line int) []int {
	l := v.buf.lines[line]
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
			stop := (x.Round()/v.metrics.tab + 1) * v.metrics.tab
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
	line := int((p.Y + v.scrollY) / float32(max(1, m.lh)))
	line = min(max(line, 0), len(v.buf.lines)-1)
	x := int(p.X+v.scrollX) - m.textOrigin
	xs := v.colX(line)
	col := 0
	for col < len(xs)-1 && x > (xs[col]+xs[col+1])/2 {
		col++
	}
	return codePos{line, col}
}

// caretPoint is the caret's top-left in editor coordinates.
func (v *CodeEditorView) caretPoint(p codePos) image.Point {
	m := v.metrics
	x := m.textOrigin + v.colX(p.line)[p.col] - int(v.scrollX)
	return image.Pt(x, p.line*m.lh-int(v.scrollY))
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

	area := clip.Rect{Max: m.size}.Push(gtx.Ops)
	defer area.Pop()
	editor := clip.Rect{Max: m.size}.Push(gtx.Ops)
	v.semantics(gtx)
	pointer.CursorText.Add(gtx.Ops)
	event.Op(gtx.Ops, v)
	key.InputHintOp{Tag: v, Hint: key.HintAny}.Add(gtx.Ops)
	v.scroll.Add(gtx.Ops)
	v.scrollH.Add(gtx.Ops)

	paint.FillShape(gtx.Ops, theme.CodeBg, clip.Rect{Max: m.size}.Op())
	first := max(0, int(v.scrollY)/max(1, m.lh))
	last := min(len(v.buf.lines)-1, first+v.visible+1)
	from, to := ordered(v.anchor, v.caret)
	for line := first; line <= last; line++ {
		v.paintLine(gtx, line, from, to)
	}
	// The gutter covers text scrolled under it.
	paint.FillShape(gtx.Ops, theme.CodeBg, clip.Rect{Max: image.Pt(m.gutter, m.size.Y)}.Op())
	paint.FillShape(gtx.Ops, theme.Border, clip.Rect{Min: image.Pt(m.gutter-1, 0), Max: image.Pt(m.gutter, m.size.Y)}.Op())
	for line := first; line <= last; line++ {
		v.paintGutter(gtx, line)
	}
	v.paintCaret(gtx)
	v.paintScrollbar(gtx)
	editor.Pop() // Floating controls are siblings of the textbox in the semantic tree.
	v.paintCompletions(gtx)
	v.paintHover(gtx)
	v.reportIME(gtx)
	return core.D{Size: m.size}
}

func (v *CodeEditorView) paintLine(gtx core.C, line int, from, to codePos) {
	m := v.metrics
	top := line*m.lh - int(v.scrollY)
	xs := v.colX(line)
	ox := m.textOrigin - int(v.scrollX)
	if line == v.caret.line && from == to && v.focused {
		c := theme.Subtle
		c.A /= 2
		paint.FillShape(gtx.Ops, c, clip.Rect{Min: image.Pt(m.gutter, top), Max: image.Pt(m.size.X, top+m.lh)}.Op())
	}
	// Selection: the selected columns, plus a sliver for a selected newline.
	if from != to && line >= from.line && line <= to.line {
		a, b := 0, len(xs)-1
		if line == from.line {
			a = from.col
		}
		if line == to.line {
			b = to.col
		}
		x0, x1 := ox+xs[a], ox+xs[b]
		if line < to.line {
			x1 += m.px / 2
		}
		paint.FillShape(gtx.Ops, theme.Highlight, clip.Rect{Min: image.Pt(x0, top), Max: image.Pt(x1, top+m.lh)}.Op())
	}
	// Text, a piece per color and between tabs.
	l := v.buf.lines[line]
	var spans []codeSpan
	if line < len(v.buf.spans) {
		spans = v.buf.spans[line]
	}
	si := 0
	for i := 0; i < len(l); {
		if l[i] == '\t' {
			i++
			continue
		}
		for si < len(spans) && spans[si].end <= i {
			si++
		}
		c := theme.CodeText
		end := len(l)
		if si < len(spans) && spans[si].start <= i {
			c, end = spans[si].color, min(end, spans[si].end)
		} else if si < len(spans) {
			end = min(end, spans[si].start)
		}
		j := i
		for j < end && l[j] != '\t' {
			j++
		}
		if x := ox + xs[i]; x < m.size.X && ox+xs[j] > m.gutter {
			v.paintText(gtx, string(l[i:j]), image.Pt(x, top+m.baseline), c)
		}
		i = j
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
		x0, x1 := ox+xs[a], max(ox+xs[b], ox+xs[a]+m.px/2)
		squiggle(gtx, x0, x1, top+m.lh-gtx.Dp(3), gtx.Dp(2), codeSeverityColor(d.Severity))
	}
}

func (v *CodeEditorView) paintText(gtx core.C, s string, at image.Point, c color.NRGBA) {
	glyphs := v.shape(v.metrics.px, s)
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

func (v *CodeEditorView) paintGutter(gtx core.C, line int) {
	m := v.metrics
	top := line*m.lh - int(v.scrollY)
	num := strconv.Itoa(line + 1)
	w := 0
	for _, g := range v.shape(m.px, num) {
		w += g.Advance.Round()
	}
	c := theme.Muted
	if line == v.caret.line {
		c = theme.CodeText
	}
	v.paintText(gtx, num, image.Pt(m.gutter-gtx.Dp(10)-w, top+m.baseline), c)
	// A dot for the most serious diagnostic on the line.
	worst := -1
	for _, d := range v.diagnostics {
		if line >= d.Line && line <= d.EndLine && (worst < 0 || int(d.Severity) < worst) {
			worst = int(d.Severity)
		}
	}
	if worst >= 0 {
		r := gtx.Dp(3)
		cy := top + m.lh/2
		paint.FillShape(gtx.Ops, codeSeverityColor(CodeSeverity(worst)), clip.Ellipse{Min: image.Pt(gtx.Dp(6)-r, cy-r), Max: image.Pt(gtx.Dp(6)+r, cy+r)}.Op(gtx.Ops))
	}
}

func (v *CodeEditorView) paintCaret(gtx core.C) {
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
	p := v.caretPoint(v.caret)
	if p.X < m.gutter {
		return
	}
	w := max(1, gtx.Dp(2))
	inset := (m.lh - m.ascent - m.descent) / 2
	paint.FillShape(gtx.Ops, theme.Primary, clip.Rect{Min: image.Pt(p.X-w/2, p.Y+inset), Max: image.Pt(p.X-w/2+w, p.Y+m.lh-inset)}.Op())
}

func (v *CodeEditorView) paintScrollbar(gtx core.C) {
	m := v.metrics
	total := len(v.buf.lines) * m.lh
	if total <= m.size.Y {
		return
	}
	h := max(gtx.Dp(24), m.size.Y*m.size.Y/total)
	y := int(v.scrollY) * (m.size.Y - h) / max(1, total-m.size.Y)
	c := theme.Muted
	c.A = 0x60
	w := gtx.Dp(6)
	paint.FillShape(gtx.Ops, c, clip.UniformRRect(image.Rect(m.size.X-w-gtx.Dp(2), y, m.size.X-gtx.Dp(2), y+h), w/2).Op(gtx.Ops))
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

func (v *CodeEditorView) textWidth(s string) int {
	w := 0
	for _, g := range v.shape(v.metrics.px, s) {
		w += g.Advance.Round()
	}
	return w
}

func (v *CodeEditorView) paintHover(gtx core.C) {
	if v.hoverText == "" {
		return
	}
	m := v.metrics
	w := 0
	for _, g := range v.shape(m.px, v.hoverText) {
		w += g.Advance.Round()
	}
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
	if len(v.buf.lines) <= 2000 {
		semantic.DescriptionOp(v.buf.text()).Add(gtx.Ops)
	} else {
		semantic.DescriptionOp(locale.Current().Rows(len(v.buf.lines))).Add(gtx.Ops)
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
			if v.buf.revision == rev && len(spans) == len(v.buf.lines) {
				v.buf.spans, v.hlRev, v.hlStyle = spans, rev, style
			}
		})
	}()
}

func (v *CodeEditorView) clampScroll() {
	m := v.metrics
	maxY := float32(max(0, len(v.buf.lines)*m.lh-m.size.Y+m.lh))
	v.scrollY = min(max(v.scrollY, 0), maxY)
	v.scrollX = max(v.scrollX, 0)
}

// scrollToCaret keeps the caret a line away from the edges.
func (v *CodeEditorView) scrollToCaret() {
	m := v.metrics
	if m.lh == 0 {
		return
	}
	y := float32(v.caret.line * m.lh)
	if y < v.scrollY {
		v.scrollY = y
	} else if bottom := y + float32(m.lh) - float32(m.size.Y); bottom > v.scrollY {
		v.scrollY = bottom
	}
	x := float32(v.colX(v.caret.line)[v.caret.col])
	view := float32(m.size.X - m.textOrigin - m.pad)
	if x < v.scrollX {
		v.scrollX = max(0, x-view/3)
	} else if x > v.scrollX+view {
		v.scrollX = x - view*2/3
	}
}
