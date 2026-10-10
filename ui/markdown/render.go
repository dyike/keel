package markdown

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/dyike/keel/ui/locale"
	"golang.org/x/image/math/fixed"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Colors and fonts of rendered Markdown. Change them before rendering.
var (
	CodeBg       color.NRGBA
	CodeBorder   color.NRGBA
	CodeHover    color.NRGBA
	InlineCode   color.NRGBA
	InlineCodeBg color.NRGBA
	CodeStyle    = "" // a chroma style name
	MonoFace     = theme.MonoFace
	caret        = "▍"
)

var headingSizes = [...]unit.Sp{0, 24, 20, 17, 15, 15, 15}

// Render draws the document. Put it in a view's tree like any element.
func (d *Doc) Render(cx *el.Context) el.Element {
	d.refreshPalette()
	root := el.Div().Gap(12)
	if d.preview.lines > 0 {
		root.MaxH(el.Sp(float32(theme.BodySize) * 1.6 * float32(d.preview.lines))).Reveal(1)
	}
	if d.preview.maxHeight > 0 {
		root.MaxH(el.Dp(float32(d.preview.maxHeight))).Reveal(1)
	}
	var all []*block
	for i := range d.chunks {
		for j := range d.chunks[i].blocks {
			all = append(all, &d.chunks[i].blocks[j])
		}
	}
	tailIndex := len(all) - 1
	if tailIndex >= 0 && all[tailIndex].kind == footnoteList {
		tailIndex--
	}
	for i, b := range all {
		if d.streaming && b.kind == footnoteList && (tailIndex < 0 || !endsInText(all[tailIndex])) {
			root.Child(el.Text(caret).TextColor(theme.Primary))
		}
		tail := d.streaming && i == tailIndex
		if tail || d.plugins != nil && hasCustomContent(b) || d.dynamicCode(b) {
			root.Child(d.block(cx, b, tail)) // streaming or custom controls: rebuild
			continue
		}
		// A finished block looks the same until its code is copied: reuse
		// its elements and layout across frames.
		b := b
		cacheKey := keyForBlock(b)
		cacheKey.extensions = d.extensionRevision
		root.Child(cx.Cache(cacheKey, func() el.Element { return d.block(cx, b, false) }))
	}
	if d.streaming && (len(all) == 0 || all[len(all)-1].kind != footnoteList && !endsInText(all[len(all)-1])) {
		root.Child(el.Text(caret).TextColor(theme.Primary))
	}
	root.Decorate(func(gtx core.C, draw func()) {
		d.beginPreview(cx, root, gtx)
		d.navigate(cx, root, gtx)
		d.selection.paint(gtx, cx, root, func() {
			d.syncStreamFade(gtx)
			d.syncRanges()
			d.paintPreview(gtx, draw)
		})
		d.revealRange(cx, gtx)
	})
	return root
}

// endsInText reports whether a block's last line is text the caret can follow.
func endsInText(b *block) bool { return b.kind == paragraph || b.kind == heading }

func (d *Doc) block(cx *el.Context, b *block, last bool) el.Element {
	if b.custom != nil {
		return b.custom.Render(cx)
	}
	switch b.kind {
	case footnoteList:
		footer := el.Div().Role("footnotes").Gap(8).Pt(8).Child(el.Div().H(el.Dp(1)).Bg(theme.Border))
		for i := range b.children {
			footer.Child(d.block(cx, &b.children[i], false))
		}
		return footer
	case footnoteItem:
		content := el.Div().Grow().Gap(8)
		for i := range b.children {
			content.Child(d.block(cx, &b.children[i], false))
		}
		return el.Div().Row().Gap(6).Child(el.Text(strconv.Itoa(b.level)+".").W(el.Dp(22)).TextColor(theme.Muted), content)
	case paragraph:
		return el.Widget(d.rich(b, b.spans, theme.BodySize, false, theme.Text, last))
	case heading:
		size := headingSizes[min(b.level, 6)]
		return el.Div().Pt(4).Child(el.Widget(d.rich(b, b.spans, size, true, theme.Text, last)))
	case codeBlock:
		return d.code(cx, b)
	case quote:
		content := el.Div().Grow().Gap(8)
		for i := range b.children {
			content.Child(d.block(cx, &b.children[i], false))
		}
		return el.Div().Row().Gap(12).Child(el.Div().W(el.Dp(3)).Rounded(1.5).Bg(theme.Border), content)
	case list:
		return d.list(cx, b)
	case table:
		return d.table(cx, b)
	case rule:
		return el.Div().H(el.Dp(1)).Bg(theme.Border).My(4)
	case frontMatter:
		return d.frontMatterTable(b)
	case group:
		content := el.Div().Gap(12)
		for i := range b.children {
			content.Child(d.block(cx, &b.children[i], last && i == len(b.children)-1))
		}
		return content
	}
	return nil
}

func (d *Doc) list(cx *el.Context, b *block) el.Element {
	lv, _ := b.view.(*listView)
	if lv == nil {
		lv = &listView{}
		b.view = lv
	}
	out := el.Div().Gap(6)
	for i := range b.items {
		it := &b.items[i]
		marker := "•"
		switch {
		case it.task != nil && *it.task:
			marker = "☑"
		case it.task != nil:
			marker = "☐"
		case b.ordered:
			marker = strconv.Itoa(b.start+i) + "."
		}
		content := el.Div().Grow().Gap(6)
		for j := range it.blocks {
			content.Child(d.block(cx, &it.blocks[j], false))
		}
		out.Child(el.Div().Row().Gap(6).Items(el.Start).Child(
			// Drawn by the same rich text as the item, so it shares its line
			// height and vertical centering and lines up with the first line.
			el.Div().W(el.Dp(22)).Items(el.End).NoShrink().Child(el.Widget(lv.marker(i, marker))),
			content,
		))
	}
	return out
}

func (d *Doc) table(cx *el.Context, b *block) el.Element {
	t := b.tbl
	st, _ := b.view.(*tableView)
	if st == nil {
		st = &tableView{}
		b.view = st
	}
	cols := len(t.header)
	divider := theme.Border
	if d.hideTableDividers {
		divider = color.NRGBA{}
	}
	cell := func(key string, spans []span, i int, bold bool) el.Element {
		align := alignLeft
		if i < len(t.align) {
			align = t.align[i]
		}
		r := st.cell(key, spans, bold, align, d.followLink)
		d.bindImages(r, spans)
		r.separator = "\t"
		if i == 0 {
			r.separator = "\n"
			if key == "h0" {
				r.separator = "\n\n"
			}
		}
		box := el.Div().W(el.Dp(st.basis(t)[i])).Px(10).Py(8).Child(el.Widget(r))
		if i < cols-1 && !d.hideTableDividers {
			// A separator, rather than a cell outline: keep it visible in a
			// frameless palette and span the tallest (possibly wrapped) cell.
			box.Child(el.Div().Absolute().Right(0).Top(0).Bottom(0).W(el.Dp(1)).Bg(theme.Border))
		}
		return box
	}
	// Columns take their widest cell's natural width, the same in every row,
	// so they line up and the table is as wide as its content, as in a web
	// page. Wider than the document, rows shrink alike and cells wrap.
	grid := el.Div().Role("table").Value(locale.Current().Rows(len(t.rows))).Border(1, theme.Border).Rounded(6)
	head := el.Div().Row().Items(el.Stretch).Bg(theme.Subtle).Role("row").Name(joinCells(t.header))
	for i := range cols {
		head.Child(cell("h"+strconv.Itoa(i), t.header[i], i, true))
	}
	grid.Child(head)
	for r, row := range t.rows {
		line := el.Div().Row().Items(el.Stretch).Role("row").Name(joinCells(row))
		for i := range cols {
			var spans []span
			if i < len(row) {
				spans = row[i]
			}
			line.Child(cell(strconv.Itoa(r)+"/"+strconv.Itoa(i), spans, i, false))
		}
		row := el.Div().Child(el.Div().H(el.Dp(1)).Bg(divider), line)
		d.guardPreview(cx, row)
		grid.Child(row)
	}
	d.guardPreview(cx, grid)
	// A row keeps the column from stretching the table to its width.
	return el.Div().Row().Child(grid)
}

func joinCells(cells [][]span) string {
	var parts []string
	for _, c := range cells {
		parts = append(parts, plain(c))
	}
	return strings.Join(parts, " | ")
}

// tableView keeps each cell's rich text state across frames.
type tableView struct {
	cells  map[string]*richBlock
	widths []float32 // column widths in dp, cell padding included
}

// tableTextScale measures cell text at this many pixels per em, then scales.
const tableTextScale = 100

// basis returns each column's width: its widest single-line cell plus the
// cell's horizontal padding. Measured once; a table's cells do not change.
func (st *tableView) basis(t *tableData) []float32 {
	if len(st.widths) == len(t.header) {
		return st.widths
	}
	sh := theme.Material.Shaper
	measure := func(spans []span, bold bool) float32 {
		f := font.Font{Typeface: theme.Face}
		if bold {
			f.Weight = font.Bold
		}
		sh.LayoutString(text.Parameters{Font: f, PxPerEm: fixed.I(tableTextScale), MaxWidth: 1 << 24}, plain(spans))
		width := 0
		for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
			width = max(width, (g.X + g.Advance).Ceil())
		}
		return float32(width) * float32(theme.BodySize) / tableTextScale
	}
	st.widths = make([]float32, len(t.header))
	for i := range t.header {
		w := measure(t.header[i], true)
		for _, row := range t.rows {
			if i < len(row) {
				w = max(w, measure(row[i], false))
			}
		}
		st.widths[i] = w + 20 + 2 // Px(10) each side, and rounding room
	}
	return st.widths
}

func (t *tableView) cell(key string, spans []span, bold bool, align cellAlign, onLink func(string)) *richBlock {
	if t.cells == nil {
		t.cells = map[string]*richBlock{}
	}
	r := t.cells[key]
	if r == nil {
		r = newRich(spans, theme.BodySize, bold, theme.Text, onLink)
		r.align = [...]text.Alignment{text.Start, text.Middle, text.End}[align]
		t.cells[key] = r
	}
	return r
}

// rich returns the paragraph's rich text, kept in the block across frames.
// The streaming tail gets a caret appended.
func (d *Doc) rich(b *block, spans []span, size unit.Sp, bold bool, c color.NRGBA, withCaret bool) *richBlock {
	r, _ := b.view.(*richBlock)
	if r == nil {
		r = newRich(spans, size, bold, c, d.followLink)
		b.view = r
		d.bindImages(r, spans)
	}
	r.anchor = b.anchor
	r.caret = withCaret
	return r
}

// richBlock draws styled runs with clickable links; see text.go.
type richBlock struct {
	measured      [8]measurement // recent measurements by constraints, see Layout
	nextSlot      int
	runs          []run
	base          unit.Sp
	lineH         float32 // line height as a multiple of the text size
	plain         string
	align         text.Alignment
	caret         bool
	links         bool
	rt            richText
	onLink        func(string)
	imageRevision uint64
	anchor        string // internal navigation destination
	separator     string // plain-text boundary before this block; default: blank line
	code          bool   // triple-click selects a source line in code
	decoration    bool   // list markers are not part of selectable text
}

func newRich(spans []span, size unit.Sp, bold bool, c color.NRGBA, onLink func(string)) *richBlock {
	r := &richBlock{plain: plain(spans), onLink: onLink, lineH: 1.6, base: size}
	for _, s := range spans {
		rn := run{object: s.object, text: s.text, math: s.math, display: s.display, size: size, color: c, font: font.Font{Typeface: theme.Face}, strike: s.strike, anchor: s.anchor}
		if s.superscript {
			rn.size = size * 0.75
			rn.rise = size * 0.3
		}
		if s.subscript {
			rn.size = size * 0.75
			rn.rise = -size * 0.15
		}
		rn.underline = s.underline
		if s.mark {
			bg := theme.Highlight
			rn.bg = &bg
		}
		if s.bold || bold {
			rn.font.Weight = font.Bold
		}
		if s.italic {
			rn.font.Style = font.Italic
		}
		if s.code {
			rn.font.Typeface = MonoFace
			rn.color = followColor(InlineCode, theme.CodeText)
			bg := followColor(InlineCodeBg, theme.CodeBg)
			rn.bg = &bg
			rn.size = size * 0.92
		}
		if s.link != "" {
			rn.color = theme.Primary
			rn.link = s.link
			r.links = true
		}
		r.runs = append(r.runs, rn)
	}
	return r
}

func (r *richBlock) Layout(gtx core.C) core.D {
	revision := r.imagesRevision()
	if revision != r.imageRevision {
		r.imageRevision = revision
		clear(r.measured[:])
	}
	// el measures an element by laying it out with input disabled, often
	// several times a frame, then paints it. Laying out rich text is costly
	// and a block's text does not change, so a repeated measurement returns
	// the size from before.
	if !gtx.Enabled() && !r.hasInlineObjects() {
		for _, m := range r.measured {
			if m.ok && m.cs == gtx.Constraints && m.caret == r.caret {
				return core.D{Size: m.size}
			}
		}
	}
	r.rt.updateLinks(gtx, r.onLink)
	r.rt.decoration = r.decoration
	if r.rt.document == nil && !r.decoration {
		r.rt.updateSelection(gtx)
	}
	runs := r.runs
	if r.caret {
		size := theme.BodySize
		if len(runs) > 0 {
			size = runs[len(runs)-1].size
		}
		runs = append(runs[:len(runs):len(runs)], run{text: caret, size: size, color: theme.Primary, font: font.Font{Typeface: theme.Face}})
	}
	r.rt.runs, r.rt.base, r.rt.lineH, r.rt.align = runs, r.base, r.lineH, r.align
	r.rt.textRuns = len(r.runs) // not the caret
	sem := []interface{ Add(*op.Ops) }{semantic.LabelOp(r.plain)}
	if r.links || r.imagesRevision() > 0 {
		// A container, so agents see each link inside it on its own.
		sem = append(sem, core.Role("paragraph"))
	}
	d := core.Semantic(gtx, func(gtx core.C) core.D {
		return r.rt.Layout(gtx, theme.Material.Shaper)
	}, sem...)
	if !gtx.Enabled() {
		r.measured[r.nextSlot] = measurement{gtx.Constraints, r.caret, d.Size, true}
		r.nextSlot = (r.nextSlot + 1) % len(r.measured)
	}
	return d
}

// highlight colors code with the installed core.Highlighter (import
// ui/highlight); unknown languages are guessed, then plain.
func highlight(lang, code string) *richBlock {
	// Tabs render as a narrow space in proportional-width shaping; expand them.
	code = strings.ReplaceAll(code, "\t", "    ")
	r := &richBlock{plain: code, code: true, lineH: 1.45, base: theme.BodySize * 0.9}
	base := run{size: theme.BodySize * 0.9, color: theme.CodeText, font: font.Font{Typeface: MonoFace}}
	var tokens []core.CodeToken
	ok := false
	if h := core.CurrentHighlighter(); h != nil {
		tokens, ok = h.Highlight(code, core.HighlightOptions{Language: lang, Guess: lang == "", Style: codeStyle()})
	}
	if !ok {
		base.text = code
		r.runs = []run{base}
		return r
	}
	for _, tok := range tokens {
		rn := base
		rn.text = tok.Text
		if tok.Color.A != 0 {
			rn.color = tok.Color
		}
		if tok.Bold {
			rn.font.Weight = font.Bold
		}
		if tok.Italic {
			rn.font.Style = font.Italic
		}
		r.runs = append(r.runs, rn)
	}
	return r
}

type blockKey struct {
	extensions uint64
	palette    paletteKey
	b          *block
	state      uint64
}

func keyForBlock(b *block) blockKey {
	k := blockKey{b: b, state: 14695981039346656037, palette: currentPaletteKey()}
	var visit func(*block)
	visit = func(b *block) {
		switch view := b.view.(type) {
		case *richBlock:
			k.state = (k.state ^ view.imagesRevision()) * 1099511628211
		case *tableView:
			var revision uint64
			for _, r := range view.cells {
				revision += r.imagesRevision()
			}
			k.state = (k.state ^ revision) * 1099511628211
		}
		if cv, ok := b.view.(*codeView); ok {
			var state uint64
			if cv.wrap {
				state |= 1
			}
			if time.Since(cv.copied) < 2*time.Second {
				state |= 2
			}
			k.state = (k.state ^ state) * 1099511628211
		}
		for i := range b.children {
			visit(&b.children[i])
		}
		for i := range b.items {
			for j := range b.items[i].blocks {
				visit(&b.items[i].blocks[j])
			}
		}
	}
	visit(b)
	return k
}

type measurement struct {
	cs    layout.Constraints
	caret bool
	size  image.Point
	ok    bool
}

// listView keeps a list's marker text across frames.
type listView struct{ markers map[int]*richBlock }

func (l *listView) marker(i int, text string) *richBlock {
	if l.markers == nil {
		l.markers = map[int]*richBlock{}
	}
	r := l.markers[i]
	if r == nil {
		r = newRich([]span{{text: text}}, theme.BodySize, false, theme.Muted, nil)
		r.decoration = true
		l.markers[i] = r
	}
	return r
}
