package markdown

// Rich text with decorations and clickable links.
//
// The line breaking follows gioui.org/x/styledtext (Unlicense OR MIT): each
// run is shaped with the width left on the current line, and what does not fit
// continues on the next. On top of that this keeps every piece of a run it
// placed, so it can draw what styledtext cannot: backgrounds behind inline
// code, strikethrough, link underlines, and per-link hit areas that agents see
// as separate links. Lines center their glyphs vertically in the line height,
// and runes of different sizes share a baseline.

import (
	"image"
	"image/color"
	"strings"
	"unicode/utf8"

	"io"
	"math"
	"unicode"

	"gioui.org/f32"
	"gioui.org/font"

	"gioui.org/gesture"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/image/math/fixed"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/imageload"
	"github.com/dyike/keel/ui/theme"
)

// run is a stretch of text in one style.
type run struct {
	object    *InlineObject
	text      string
	font      font.Font
	size      unit.Sp
	color     color.NRGBA
	bg        *color.NRGBA // e.g. inline code
	strike    bool
	underline bool
	link      string
	math      *mathExpr
	display   bool
	anchor    string
	rise      unit.Sp
	image     *imageload.View
}

// piece is the part of a run placed on one line.
type piece struct {
	run      int
	text     string
	rect     image.Rectangle // logical box: x extent, the line's full height
	baseline int
	call     op.CallOp // the glyphs, drawn at rect.Min.X and the run's own baseline
	ascent   int
	descent  int
	ink      [2]int   // visible glyph top and bottom relative to the baseline
	start    int      // index of the piece's first rune in the whole text
	runes    int      // runes shown
	glyphs   []glyphX // for mapping between x positions and runes
}

// glyphX is one glyph cluster of a piece: where it starts, how wide, how many runes.
type glyphX struct{ x, adv, runes int }

// runeAt returns how many of the piece's runes lie left of x (relative to the piece).
func (p piece) runeAt(x int) int {
	n := 0
	for _, g := range p.glyphs {
		if x < g.x+g.adv/2 {
			return n
		}
		n += g.runes
	}
	return p.runes
}

// xAt returns where the piece's rune number n starts, relative to the piece.
func (p piece) xAt(n int) int {
	k := 0
	for _, g := range p.glyphs {
		if k >= n {
			return g.x
		}
		k += g.runes
	}
	return p.rect.Dx()
}

// richText lays out runs and remembers where it put them.
type richText struct {
	runs  []run
	base  unit.Sp // the paragraph's text size; lines are base×lineH tall at least
	lineH float32 // line height as a multiple of the text size
	align text.Alignment
	// inkCenter centers each line's own ink rather than a representative
	// body: symbols such as ☑ come from faces whose ink sits apart from
	// letters.
	inkCenter bool
	pieces    []piece
	links     map[int]*gesture.Click // by run index
	size      image.Point

	// Selection, in rune indexes into the text of the first textRuns runs
	// (later runs, like a streaming caret, are decoration).
	textRuns       int
	sel            [2]int // anchor, focus
	dragging       bool
	document       *documentSelection
	decoration     bool
	offset, length int // rune range in the document, excluding decorations
}

func (r *richText) link(i int) *gesture.Click {
	if r.links == nil {
		r.links = map[int]*gesture.Click{}
	}
	c := r.links[i]
	if c == nil {
		c = new(gesture.Click)
		r.links[i] = c
	}
	return c
}

// updateLinks reports clicked link URLs since the last frame.
func (r *richText) updateLinks(gtx layout.Context, fn func(url string)) {
	for i, c := range r.links {
		for {
			ev, ok := c.Update(gtx.Source)
			if !ok {
				break
			}
			if ev.Kind == gesture.KindClick && fn != nil && i < len(r.runs) && (r.document == nil || !r.document.moved) {
				url := r.runs[i].link
				core.Call(gtx, func() { fn(url) })
			}
		}
	}
}

// Layout places and draws the runs within gtx.Constraints.Max.X.
func (r *richText) Layout(gtx layout.Context, shaper *text.Shaper) layout.Dimensions {
	r.pieces = r.pieces[:0]
	maxW := gtx.Constraints.Max.X
	size := r.base
	for _, rn := range r.runs {
		size = max(size, rn.size)
	}
	lineH := int(float32(gtx.Sp(size))*r.lineH + 0.5)
	sc := scriptOfRuns(r.runs)

	var line []piece
	y, width := 0, 0
	lineW := 0
	pos := 0 // runes of text consumed so far
	flush := func() {
		if len(line) == 0 {
			y += lineH
			return
		}
		asc, desc := 0, 0
		for _, p := range line {
			asc, desc = max(asc, p.ascent), max(desc, p.descent)
		}
		// Share one baseline. Text lines center the glyphs' ink body, as
		// el.Text does: centering the font's logical box set CJK text high,
		// since PingFang reserves a deep empty descent, leaving selection
		// and highlight backgrounds visibly below the text. Lines with
		// objects, math or images keep centering the logical box.
		height := lineH
		objects := false
		bodyTop, bodyBottom, body := 0, 0, false
		for _, p := range line {
			rn := r.runs[p.run]
			if rn.object != nil || rn.math != nil || rn.image != nil {
				objects = true
				height = max(height, asc+desc+gtx.Dp(2))
				continue
			}
			if rn.rise != 0 {
				continue
			}
			t, b := inkBody(shaper, rn.font, gtx.Sp(rn.size), sc)
			if r.inkCenter {
				t, b = p.ink[0], p.ink[1]
			}
			if !body {
				bodyTop, bodyBottom, body = t, b, true
			} else {
				bodyTop, bodyBottom = min(bodyTop, t), max(bodyBottom, b)
			}
		}
		baseline := y + (height-(asc+desc))/2 + asc
		if body && !objects {
			// Round halves down, as el.Text does.
			baseline = y + int(math.Floor(float64(height-(bodyTop+bodyBottom))/2+0.5))
		}
		pad := 0
		switch r.align {
		case text.Middle:
			pad = (maxW - lineW) / 2
		case text.End:
			pad = maxW - lineW
		}
		for _, p := range line {
			p.rect = image.Rect(p.rect.Min.X+pad, y, p.rect.Max.X+pad, y+height)
			p.baseline = baseline
			r.pieces = append(r.pieces, p)
		}
		width = max(width, lineW)
		y += height
		line, lineW = line[:0], 0
	}

	for i := 0; i < len(r.runs); i++ {
		rn := r.runs[i]
		if rn.object != nil {
			measure := gtx.Disabled()
			measure.Constraints = layout.Constraints{Max: image.Pt(maxW, codeWidth)}
			scratch := new(op.Ops)
			measure.Ops = scratch
			natural := rn.object.Widget.Layout(measure)
			if lineW > 0 && natural.Size.X > maxW-lineW {
				flush()
			}
			g := gtx
			g.Constraints = layout.Constraints{Max: image.Pt(maxW-lineW, codeWidth)}
			m := op.Record(gtx.Ops)
			dims := rn.object.Widget.Layout(g)
			call := m.Stop()
			count := utf8.RuneCountInString(rn.text)
			line = append(line, piece{run: i, text: rn.text, rect: image.Rect(lineW, 0, lineW+dims.Size.X, 0), call: call, ascent: dims.Size.Y, start: pos, runes: count, glyphs: []glyphX{{adv: dims.Size.X, runes: count}}})
			lineW += dims.Size.X
			pos += count
			continue
		}
		if rn.image != nil {
			// Images keep their aspect ratio, break as one unit, and use alt
			// text as their atomic range in document selection.
			natural := rn.image.Asset.Size().X
			if natural == 0 {
				natural = 260
			}
			if lineW > 0 && gtx.Dp(unit.Dp(natural)) > maxW-lineW {
				flush()
			}
			g := gtx
			g.Constraints = layout.Constraints{Max: image.Pt(maxW-lineW, codeWidth)}
			m := op.Record(gtx.Ops)
			dims := rn.image.Layout(g)
			call := m.Stop()
			count := utf8.RuneCountInString(rn.text)
			line = append(line, piece{run: i, text: rn.text, rect: image.Rect(lineW, 0, lineW+dims.Size.X, 0), call: call, ascent: dims.Size.Y, start: pos, runes: count, glyphs: []glyphX{{adv: dims.Size.X, runes: count}}})
			lineW += dims.Size.X
			pos += count
			continue
		}
		if rn.math != nil {
			b := layoutMath(gtx, shaper, rn.math, rn, rn.display)
			if len(line) > 0 && (rn.display || lineW+b.w > maxW) {
				flush()
			}
			x := lineW
			if rn.display {
				x = max(0, (maxW-b.w)/2)
			}
			count := utf8.RuneCountInString(rn.text)
			line = append(line, piece{run: i, text: rn.text, rect: image.Rect(x, 0, x+b.w, 0), call: b.call, ascent: b.a, descent: b.d, start: pos, runes: count, glyphs: []glyphX{{adv: b.w, runes: count}}})
			lineW = x + b.w
			pos += count
			if rn.display {
				flush()
			}
			continue
		}
		content := rn.text
		for content != "" {
			if content[0] == '\n' {
				// A hard break: Gio shapes nothing for it on a one-line
				// layout, so handle it here. Two in a row make an empty line.
				flush()
				content = content[1:]
				pos++
				continue
			}
			res := shapeLine(gtx, shaper, rn, content, maxW-lineW, true)
			if res.runes == 0 {
				if lineW > 0 {
					flush() // nothing fits after what is on the line: next line
					continue
				}
				// Nothing fits even on an empty line: place the first word anyway.
				res = shapeLine(gtx, shaper, rn, content, maxW, false)
				if res.runes == 0 {
					break
				}
			}
			n := byteLen(content, res.runes)
			// A line break right after what was shaped belongs to this line:
			// with MaxLines 1 Gio stops before it rather than counting it.
			if n < len(content) && content[n] == '\n' {
				n++
				res.newline = true
			}
			shown := strings.TrimSuffix(content[:n], "\n")
			if shown != "" {
				count := utf8.RuneCountInString(shown)
				line = append(line, piece{run: i, text: shown, rect: image.Rect(lineW, 0, lineW+res.width, 0),
					call: res.call, ascent: res.ascent + gtx.Sp(rn.rise), descent: max(0, res.descent-gtx.Sp(rn.rise)), ink: [2]int{-res.inkAscent, res.inkDescent}, start: pos, runes: count, glyphs: res.glyphs})
				lineW += res.width
			}
			pos += utf8.RuneCountInString(content[:n])
			content = content[n:]
			if content != "" || res.newline {
				flush() // the run wrapped, or hit a hard line break
			}
		}
	}
	if len(line) > 0 {
		flush()
	}

	r.size = gtx.Constraints.Constrain(image.Pt(width, y))
	r.paint(gtx)
	return layout.Dimensions{Size: r.size}
}

func (r *richText) paint(gtx layout.Context) {
	ops := gtx.Ops
	// Backgrounds first, under the glyphs.
	for _, p := range r.pieces {
		rn := r.runs[p.run]
		if rn.bg == nil {
			continue
		}
		em := gtx.Sp(rn.size)
		bg := image.Rect(p.rect.Min.X-gtx.Dp(2), p.baseline-em*9/10, p.rect.Max.X+gtx.Dp(2), p.baseline+em*3/10)
		paint.FillShape(ops, *rn.bg, clip.UniformRRect(bg, gtx.Dp(3)).Op(ops))
	}
	if d := r.document; d != nil && !r.decoration {
		for _, h := range d.highlights {
			r.paintRange(gtx, max(0, h.start-r.offset), min(r.length, h.end-r.offset), h.background)
		}
	}
	lo, hi := r.selRange()
	r.paintRange(gtx, lo, hi, SelectionBg)
	for _, p := range r.pieces {
		rn := r.runs[p.run]
		r.paintFadedPiece(gtx, p, func() {
			stk := op.Offset(image.Pt(p.rect.Min.X, p.baseline-p.ascent)).Push(ops)
			paint.ColorOp{Color: rn.color}.Add(ops)
			p.call.Add(ops)
			stk.Pop()
			em := gtx.Sp(rn.size)
			thick := max(1, em/14)
			switch {
			case rn.strike:
				yy := p.baseline - em*3/10
				paint.FillShape(ops, rn.color, clip.Rect(image.Rect(p.rect.Min.X, yy, p.rect.Max.X, yy+thick)).Op())
			case rn.underline || rn.link != "" && r.link(p.run).Hovered():
				yy := p.baseline + em/8
				paint.FillShape(ops, rn.color, clip.Rect(image.Rect(p.rect.Min.X, yy, p.rect.Max.X, yy+thick)).Op())
			}
		})
	}
	// The whole text takes pointer input for selecting; links sit on top.
	area := clip.Rect(image.Rectangle{Max: r.size}).Push(ops)
	if r.document == nil && !r.decoration {
		event.Op(ops, r)
	}
	if !r.decoration {
		pointer.CursorText.Add(ops)
	}
	area.Pop()

	// Links: a hit area and a semantic node per piece, so agents can click them.
	for _, p := range r.pieces {
		rn := r.runs[p.run]
		if rn.link == "" {
			continue
		}
		c := r.link(p.run)
		area := clip.Rect(p.rect).Push(ops)
		semantic.Button.Add(ops)
		core.Role("link", rn.link).Add(ops)
		semantic.LabelOp(strings.TrimSpace(rn.text)).Add(ops)
		pointer.CursorPointer.Add(ops)
		c.Add(ops)
		area.Pop()
	}
}

type lineResult struct {
	call                  op.CallOp
	width                 int
	inkAscent, inkDescent int // visible glyph bounds relative to the baseline
	ascent, descent       int // baseline below the piece's top; extent below the baseline
	runes                 int
	newline               bool // ended at a hard line break
	glyphs                []glyphX
}

// shapeLine shapes as much of content as fits on one line of width maxW. With
// truncate false it wraps instead and keeps the first line, to place a word
// that is wider than the line.
func shapeLine(gtx layout.Context, shaper *text.Shaper, rn run, content string, maxW int, truncate bool) lineResult {
	maxLines := 1
	if !truncate {
		maxLines = 0
	}
	m := op.Record(gtx.Ops)
	shaper.LayoutString(text.Parameters{
		Font:       rn.font,
		PxPerEm:    fixed.I(gtx.Sp(rn.size)),
		MaxLines:   maxLines,
		MaxWidth:   max(maxW, 0),
		Truncator:  "\u200b",
		Locale:     gtx.Locale,
		WrapPolicy: text.WrapHeuristically,
	}, content)
	var res lineResult
	var glyphs [32]text.Glyph
	buf := glyphs[:0]
	first := true
	var firstX fixed.Int26_6
	var lineY int32
	flushGlyphs := func() {
		if len(buf) == 0 {
			return
		}
		// Shape draws relative to the batch's first glyph: its pen position
		// and baseline. Place that from the line start and the piece's top.
		theme.PaintGlyphs(gtx.Ops, shaper, buf, rn.color, f32.Pt(float32((buf[0].X-firstX).Floor()), float32(buf[0].Y)))
		buf = buf[:0]
	}
	for g, ok := shaper.NextGlyph(); ok; g, ok = shaper.NextGlyph() {
		if g.Flags&text.FlagTruncator != 0 {
			continue // marks where the line was cut; not text
		}
		if first {
			first, firstX, lineY = false, g.X, g.Y
			res.ascent = int(g.Y)
		}
		if g.Y != lineY {
			break // the second line of a wrapped layout
		}
		res.runes += int(g.Runes)
		res.glyphs = append(res.glyphs, glyphX{x: (g.X - firstX).Floor(), adv: g.Advance.Ceil(), runes: int(g.Runes)})
		res.descent = max(res.descent, g.Descent.Ceil())
		res.inkAscent = max(res.inkAscent, -g.Bounds.Min.Y.Floor())
		res.inkDescent = max(res.inkDescent, g.Bounds.Max.Y.Ceil())
		res.width = max(res.width, (g.X + g.Advance - firstX).Ceil())
		if g.Flags&text.FlagParagraphBreak != 0 {
			res.newline = true
		}
		buf = append(buf, g)
		if cap(buf) == len(buf) {
			flushGlyphs()
		}
	}
	flushGlyphs()
	res.call = m.Stop()
	return res
}

func byteLen(s string, runes int) int {
	n := 0
	for i := 0; i < runes && n < len(s); i++ {
		_, sz := utf8.DecodeRuneInString(s[n:])
		n += sz
	}
	return n
}

// SelectionBg is the color behind selected text.
var SelectionBg = color.NRGBA{R: 0xb4, G: 0xd5, B: 0xfe, A: 0xff}

func (r *richText) selRange() (int, int) {
	if d := r.document; d != nil {
		lo, hi := d.selRange()
		return min(max(lo-r.offset, 0), r.length), min(max(hi-r.offset, 0), r.length)
	}
	return min(r.sel[0], r.sel[1]), max(r.sel[0], r.sel[1])
}

// updateSelection handles dragging to select, Cmd/Ctrl+C to copy and
// Cmd/Ctrl+A to select all. Clicking elsewhere clears the selection.
func (r *richText) updateSelection(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(
			pointer.Filter{Target: r, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel},
			key.FocusFilter{Target: r},
			key.Filter{Focus: r, Name: "C", Required: key.ModShortcut},
			key.Filter{Focus: r, Name: "A", Required: key.ModShortcut},
		)
		if !ok {
			break
		}
		switch e := ev.(type) {
		case pointer.Event:
			pt := image.Pt(int(e.Position.X), int(e.Position.Y))
			switch e.Kind {
			case pointer.Press:
				if e.Buttons.Contain(pointer.ButtonPrimary) {
					i := r.hit(pt)
					r.sel, r.dragging = [2]int{i, i}, true
					gtx.Execute(key.FocusCmd{Tag: r})
				}
			case pointer.Drag:
				if r.dragging {
					r.sel[1] = r.hit(pt)
				}
			case pointer.Release, pointer.Cancel:
				r.dragging = false
			}
		case key.FocusEvent:
			if !e.Focus {
				r.sel = [2]int{}
			}
		case key.Event:
			if e.State != key.Press {
				break
			}
			switch e.Name {
			case "C":
				if s := r.selectedText(); s != "" {
					gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(s))})
				}
			case "A":
				r.sel = [2]int{0, utf8.RuneCountInString(r.text())}
			}
		}
	}
}

// hit maps a point, relative to the text, to a rune index.
func (r *richText) hit(pt image.Point) int {
	if len(r.pieces) == 0 {
		return 0
	}
	first, last := r.pieces[0], r.pieces[len(r.pieces)-1]
	if pt.Y < first.rect.Min.Y {
		return first.start
	}
	if pt.Y >= last.rect.Max.Y {
		return last.start + last.runes
	}
	var line []piece
	for _, p := range r.pieces {
		if pt.Y >= p.rect.Min.Y && pt.Y < p.rect.Max.Y {
			line = append(line, p)
		}
	}
	if len(line) == 0 { // between lines: an empty line
		for _, p := range r.pieces {
			if p.rect.Min.Y > pt.Y {
				return p.start
			}
		}
		return last.start + last.runes
	}
	if pt.X <= line[0].rect.Min.X {
		return line[0].start
	}
	for _, p := range line {
		if pt.X < p.rect.Max.X {
			return p.start + p.runeAt(pt.X-p.rect.Min.X)
		}
	}
	end := line[len(line)-1]
	return end.start + end.runes
}

func (r *richText) text() string {
	var b strings.Builder
	for i, rn := range r.runs {
		if i >= r.textRuns {
			break
		}
		b.WriteString(rn.text)
	}
	return b.String()
}

func (r *richText) selectedText() string {
	lo, hi := r.selRange()
	rs := []rune(r.text())
	lo, hi = min(lo, len(rs)), min(hi, len(rs))
	return string(rs[lo:hi])
}

// script says which faces set a text: East Asian text usually comes from a
// fallback face (PingFang behind SF Pro) whose body sits differently in its
// line box. It mirrors el's text script detection.
type script uint8

const (
	scriptLatin script = iota
	scriptWide
	scriptMixed
)

func scriptOfRuns(runs []run) script {
	var latin, wide bool
	for _, rn := range runs {
		for _, r := range rn.text {
			switch {
			case unicode.IsSpace(r):
			case r >= 0x1100 && (unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) ||
				r >= 0x3000 && r <= 0x303f || r >= 0xff00 && r <= 0xffef):
				wide = true
			default:
				latin = true
			}
		}
		if latin && wide {
			return scriptMixed
		}
	}
	if wide {
		return scriptWide
	}
	return scriptLatin
}

type inkBodyKey struct {
	shaper *text.Shaper
	font   font.Font
	px     int
	script script
}

// inkBodies caches inkBody. Touched only under the frame lock.
var inkBodies = map[inkBodyKey][2]int{}

// inkBody returns the top and bottom of a representative glyph body relative
// to the baseline: the cap height for Latin, 国 for East Asian text and both
// for mixed text, matching the bodies el.Text centers beside icons.
func inkBody(shaper *text.Shaper, f font.Font, px int, sc script) (top, bottom int) {
	key := inkBodyKey{shaper, f, px, sc}
	if b, ok := inkBodies[key]; ok {
		return b[0], b[1]
	}
	// The same samples as el's textShift; only the first body glyphs count.
	sample, body := "Ag", 1
	switch sc {
	case scriptWide:
		sample = "国"
	case scriptMixed:
		sample, body = "国Ag", 2
	}
	shaper.LayoutString(text.Parameters{Font: f, PxPerEm: fixed.I(px), MaxWidth: 1 << 20}, sample)
	top, bottom = -px*7/10, 0
	found := false
	for i := 0; ; i++ {
		g, ok := shaper.NextGlyph()
		if !ok {
			break
		}
		if i >= body || g.Bounds.Min.Y == g.Bounds.Max.Y {
			continue
		}
		t, b := g.Bounds.Min.Y.Floor(), g.Bounds.Max.Y.Ceil()
		if !found {
			top, bottom, found = t, b, true
		} else {
			top, bottom = min(top, t), max(bottom, b)
		}
	}
	inkBodies[key] = [2]int{top, bottom}
	return top, bottom
}
