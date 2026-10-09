package kit

import (
	"image"
	"image/color"
	"slices"
	"sort"
	"strings"

	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

type CodeDecorationStyle uint8

const (
	CodeDecorationFrame CodeDecorationStyle = iota
	CodeDecorationFill
	CodeDecorationText
	CodeDecorationUnderline
)

// CodeDecoration paints a tracked range without affecting layout or input.
// Nil Color uses theme.CodeText (12% alpha for fills). Coordinates count runes.
type CodeDecoration struct {
	Range CodeRange
	Style CodeDecorationStyle
	Color *color.NRGBA
	// Weight and Italic restyle the glyphs of a CodeDecorationText range.
	// Glyphs keep the regular positions, so carets and clicks do not move.
	// With either set and Color nil, the syntax color stays.
	Weight font.Weight
	Italic bool
}

// CodeDecorationCollection owns annotations independently of other extensions.
// Dropping the handle leaves its annotations; Dispose removes them permanently.
// Call all methods on the UI thread, or inside core.Update.
type CodeDecorationCollection struct {
	editor        *CodeEditorView
	entries       []CodeDecoration
	order, maxEnd []int
}

func (v *CodeEditorView) Decorations(entries ...CodeDecoration) *CodeDecorationCollection {
	c := &CodeDecorationCollection{editor: v}
	v.decorations = append(v.decorations, c)
	v.buf.onEdit = v.trackDecorations
	c.Set(entries...)
	return c
}
func cloneCodeDecoration(d CodeDecoration) CodeDecoration {
	if d.Color != nil {
		color := *d.Color
		d.Color = &color
	}
	return d
}

// Set replaces this collection only. Empty/reversed/outside ranges are discarded;
// partially valid ranges are clipped to the document.
func (c *CodeDecorationCollection) Set(entries ...CodeDecoration) {
	if c.editor == nil {
		return
	}
	c.entries = nil
	c.Append(entries...)
}
func (c *CodeDecorationCollection) Append(entries ...CodeDecoration) {
	if c.editor == nil {
		return
	}
	b := c.editor.buf
	last := codePos{b.count() - 1, len(b.line(b.count() - 1))}
	for _, d := range entries {
		a, z := d.Range.positions()
		if !a.less(z) || !a.less(last) || !((codePos{}).less(z)) || d.Style > CodeDecorationUnderline {
			continue
		}
		clip := func(p codePos) codePos {
			if p.line < 0 {
				return codePos{}
			}
			if p.line >= b.count() {
				return last
			}
			return b.clamp(p)
		}
		a, z = clip(a), clip(z)
		if !a.less(z) {
			continue
		}
		d.Range = publicCodeRange(a, z)
		c.entries = append(c.entries, cloneCodeDecoration(d))
	}
	c.order = make([]int, len(c.entries))
	for i := range c.order {
		c.order[i] = i
	}
	slices.SortStableFunc(c.order, func(i, j int) int {
		a, _ := c.entries[i].Range.positions()
		b, _ := c.entries[j].Range.positions()
		if a.less(b) {
			return -1
		}
		if b.less(a) {
			return 1
		}
		return 0
	})
	c.reindex()
}
func (c *CodeDecorationCollection) Get() []CodeDecoration {
	out := make([]CodeDecoration, len(c.entries))
	for i, d := range c.entries {
		out[i] = cloneCodeDecoration(d)
	}
	return out
}
func (c *CodeDecorationCollection) Clear() { c.Set() }
func (c *CodeDecorationCollection) Dispose() {
	if c.editor == nil {
		return
	}
	v := c.editor
	v.decorations = slices.DeleteFunc(v.decorations, func(other *CodeDecorationCollection) bool { return c == other })
	c.editor = nil
	c.entries = nil
	c.order = nil
	c.maxEnd = nil
	if len(v.decorations) == 0 {
		v.buf.onEdit = nil
	}
}

// A balanced interval index rejects subtrees ending before a visible line.
func (c *CodeDecorationCollection) reindex() {
	c.maxEnd = make([]int, len(c.order))
	var build func(int, int) int
	build = func(lo, hi int) int {
		if lo >= hi {
			return -1
		}
		mid := (lo + hi) / 2
		end := max(c.entries[c.order[mid]].Range.EndLine, build(lo, mid), build(mid+1, hi))
		c.maxEnd[mid] = end
		return end
	}
	build(0, len(c.order))
}
func (c *CodeDecorationCollection) atLine(line int) []int {
	out := []int{}
	var visit func(int, int)
	visit = func(lo, hi int) {
		if lo >= hi {
			return
		}
		mid := (lo + hi) / 2
		if c.maxEnd[mid] < line {
			return
		}
		visit(lo, mid)
		d := c.entries[c.order[mid]]
		if d.Range.Line <= line {
			if d.Range.EndLine >= line {
				out = append(out, c.order[mid])
			}
			visit(mid+1, hi)
		}
	}
	visit(0, len(c.order))
	sort.Ints(out)
	return out
}
func trackCodeAnchor(p, from, to, end codePos, start bool) codePos {
	if p.less(from) {
		return p
	}
	if from == to && p == from {
		if start {
			return end
		}
		return from
	}
	if !p.less(to) {
		return shiftPos(p, to, end)
	}
	if start {
		return from
	}
	return end
}
func (v *CodeEditorView) trackDecorations(from, to, end codePos) {
	for _, c := range v.decorations {
		mapping := make([]int, len(c.entries))
		kept := c.entries[:0]
		for i, d := range c.entries {
			a, z := d.Range.positions()
			a = trackCodeAnchor(a, from, to, end, true)
			z = trackCodeAnchor(z, from, to, end, false)
			mapping[i] = -1
			if a.less(z) {
				d.Range = publicCodeRange(a, z)
				mapping[i] = len(kept)
				kept = append(kept, d)
			}
		}
		c.entries = kept
		order := c.order[:0]
		for _, i := range c.order {
			if mapping[i] >= 0 {
				order = append(order, mapping[i])
			}
		}
		c.order = order
		c.reindex()
	}
}

// SetValue can preserve annotations in an unchanged prefix/suffix. Only the
// minimal changed span transforms anchors; undo history remains reset.
func codeTextChange(before, after string) (codePos, codePos, codePos) {
	a, b := []rune(before), []rune(strings.ReplaceAll(after, "\r\n", "\n"))
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	from := endOf(codePos{}, string(a[:prefix]))
	to := endOf(from, string(a[prefix:len(a)-suffix]))
	end := endOf(from, string(b[prefix:len(b)-suffix]))
	return from, to, end
}

type codeLineDecoration struct {
	a, b     int // columns on this line; text styles use them
	style    CodeDecorationStyle
	color    color.NRGBA
	keep     bool // keep the syntax color: a weight or italic only
	weight   font.Weight
	italic   bool
	from, to codePos // the whole range, for shapes that span rows
}

func (v *CodeEditorView) lineDecorations(line int) []codeLineDecoration {
	var out []codeLineDecoration
	n := len(v.buf.line(line))
	for _, c := range v.decorations {
		for _, i := range c.atLine(line) {
			d := c.entries[i]
			from, to := d.Range.positions()
			a, b := 0, n
			if d.Range.Line == line {
				a = d.Range.Col
			}
			if d.Range.EndLine == line {
				b = d.Range.EndCol
			}
			a, b = min(max(a, 0), n), min(max(b, 0), n)
			if a == b && (line >= d.Range.EndLine || d.Style == CodeDecorationText || d.Style == CodeDecorationUnderline) {
				continue // an empty line inside a fill or frame still joins it up
			}
			color := theme.CodeText
			if d.Style == CodeDecorationFill {
				color.A = 31
			}
			if d.Color != nil {
				color = *d.Color
			}
			keep := d.Color == nil && d.Style == CodeDecorationText && (d.Weight != 0 || d.Italic)
			out = append(out, codeLineDecoration{a: a, b: b, style: d.Style, color: color, keep: keep, weight: d.Weight, italic: d.Italic, from: from, to: to})
		}
	}
	return out
}

// decorationSpan is the x range, from the row's start, that a range covers
// on visual row r. A fill or frame that runs past a line's end covers one
// space more, like a selection, so ranges over several lines connect.
func (v *CodeEditorView) decorationSpan(from, to codePos, style CodeDecorationStyle, r int) (x0, x1 int, ok bool) {
	if r < 0 || r >= v.vrowCount() {
		return 0, 0, false
	}
	vr := v.vrow(r)
	if vr.line < from.line || vr.line > to.line {
		return 0, 0, false
	}
	xs := v.colX(vr.line)
	n := len(xs) - 1
	a, b := 0, n
	if vr.line == from.line {
		a = min(from.col, n)
	}
	if vr.line == to.line {
		b = min(to.col, n)
	}
	a, b = max(a, vr.start), min(b, vr.end)
	spill := vr.line < to.line && vr.end == n && style != CodeDecorationUnderline
	if a > b || a == b && !spill {
		return 0, 0, false
	}
	x0, x1 = xs[a]-xs[vr.start], xs[b]-xs[vr.start]
	if spill {
		x1 += v.metrics.space
	}
	return x0, x1, true
}

// paintDecorations draws fills, frames and underlines on visual row r, whose
// text starts at x ox. A frame over several rows is one outline: each row
// draws its sides, and its top and bottom only where the neighbouring row
// does not continue the shape.
func (v *CodeEditorView) paintDecorations(gtx core.C, r, ox, top int, ds []codeLineDecoration) {
	lh := v.metrics.lh
	w := max(1, gtx.Dp(1))
	for _, style := range []CodeDecorationStyle{CodeDecorationFill, CodeDecorationFrame, CodeDecorationUnderline} {
		for _, d := range ds {
			if d.style != style {
				continue
			}
			x0, x1, ok := v.decorationSpan(d.from, d.to, style, r)
			if !ok {
				continue
			}
			rect := image.Rect(ox+x0, top, ox+x1, top+lh)
			switch style {
			case CodeDecorationFill:
				paint.FillShape(gtx.Ops, d.color, clip.Rect(rect).Op())
			case CodeDecorationUnderline:
				rect.Min.Y = rect.Max.Y - gtx.Dp(2)
				rect.Max.Y = rect.Min.Y + w
				paint.FillShape(gtx.Ops, d.color, clip.Rect(rect).Op())
			case CodeDecorationFrame:
				fill := func(r image.Rectangle) { paint.FillShape(gtx.Ops, d.color, clip.Rect(r).Op()) }
				fill(image.Rect(rect.Min.X, top, rect.Min.X+w, top+lh))
				fill(image.Rect(rect.Max.X-w, top, rect.Max.X, top+lh))
				edge := func(y, neighbour int) {
					p0, p1, joined := v.decorationSpan(d.from, d.to, style, neighbour)
					if !joined || p1 <= x0 || p0 >= x1 {
						fill(image.Rect(rect.Min.X, y, rect.Max.X, y+w))
						return
					}
					if p0 > x0 {
						fill(image.Rect(rect.Min.X, y, ox+p0+w, y+w))
					}
					if p1 < x1 {
						fill(image.Rect(ox+p1-w, y, rect.Max.X, y+w))
					}
				}
				edge(top, r-1)
				edge(top+lh-w, r+1)
			}
		}
	}
}
