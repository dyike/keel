package kit

import (
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"slices"
	"sort"
	"strings"
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
	a, b  int
	style CodeDecorationStyle
	color color.NRGBA
}

func (v *CodeEditorView) lineDecorations(line int) []codeLineDecoration {
	var out []codeLineDecoration
	n := len(v.buf.line(line))
	for _, c := range v.decorations {
		for _, i := range c.atLine(line) {
			d := c.entries[i]
			a, b := 0, n
			if d.Range.Line == line {
				a = d.Range.Col
			}
			if d.Range.EndLine == line {
				b = d.Range.EndCol
			}
			a, b = min(max(a, 0), n), min(max(b, 0), n)
			if a == b {
				continue
			}
			color := theme.CodeText
			if d.Style == CodeDecorationFill {
				color.A = 31
			}
			if d.Color != nil {
				color = *d.Color
			}
			out = append(out, codeLineDecoration{a, b, d.Style, color})
		}
	}
	return out
}
func paintCodeDecorations(gtx core.C, ds []codeLineDecoration, rect func(int, int) image.Rectangle) {
	for _, style := range []CodeDecorationStyle{CodeDecorationFill, CodeDecorationFrame, CodeDecorationUnderline} {
		for _, d := range ds {
			if d.style != style {
				continue
			}
			r := rect(d.a, d.b)
			switch style {
			case CodeDecorationFill:
				paint.FillShape(gtx.Ops, d.color, clip.Rect(r).Op())
			case CodeDecorationFrame:
				paint.FillShape(gtx.Ops, d.color, clip.Stroke{Path: clip.Rect(r).Path(), Width: float32(gtx.Dp(1))}.Op())
			case CodeDecorationUnderline:
				r.Min.Y = r.Max.Y - gtx.Dp(2)
				r.Max.Y = r.Min.Y + max(1, gtx.Dp(1))
				paint.FillShape(gtx.Ops, d.color, clip.Rect(r).Op())
			}
		}
	}
}
