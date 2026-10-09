package el

import (
	"image"
	"image/color"
	"slices"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

// TextRange colors a half-open Unicode rune interval. Later ranges take priority.
// Any overlapping shaping cluster is colored as a whole; bitmap glyphs keep their colors.
type TextRange struct {
	Start, End int
	Color      color.NRGBA
}

// Ranges copies color ranges without changing text measurement or shaping.
// Shimmer, when present, takes precedence over ranges.
func (t *TextEl) Ranges(ranges ...TextRange) *TextEl {
	t.n.textRanges = slices.Clone(ranges)
	return t
}

func (e *engine) paintRangeText(n *Node, g layout.Context, size image.Point) {
	lb := e.label(n, n.text)
	shaper := lb.Shaper
	shaper.LayoutString(text.Parameters{Font: lb.Font, PxPerEm: fixed.I(g.Sp(lb.TextSize)), MaxLines: lb.MaxLines, Truncator: lb.Truncator, Alignment: lb.Alignment, WrapPolicy: lb.WrapPolicy, MaxWidth: size.X, Locale: g.Locale, LineHeight: fixed.I(g.Sp(lb.LineHeight)), LineHeightScale: lb.LineHeightScale}, lb.Text)
	glyphs := make([]text.Glyph, 0, 64)
	bounds := image.Rectangle{Max: size}
	padding := image.Rectangle{}
	for glyph, ok := shaper.NextGlyph(); ok; glyph, ok = shaper.NextGlyph() {
		padding.Min.X = min(padding.Min.X, glyph.Bounds.Min.X.Floor())
		padding.Max.X = max(padding.Max.X, (glyph.Bounds.Max.X - glyph.Advance).Ceil())
		padding.Min.Y = min(padding.Min.Y, (glyph.Bounds.Min.Y + glyph.Ascent).Floor())
		padding.Max.Y = max(padding.Max.Y, (glyph.Bounds.Max.Y - glyph.Descent).Ceil())
		if int(glyph.Y)-glyph.Ascent.Ceil() > size.Y {
			break
		}
		glyphs = append(glyphs, glyph)
	}
	bounds.Min = bounds.Min.Add(padding.Min)
	bounds.Max = bounds.Max.Add(padding.Max)
	defer clip.Rect(bounds).Push(g.Ops).Pop()
	semantic.LabelOp(n.text).Add(g.Ops)

	colors := make([]color.NRGBA, len(glyphs))
	runePos := 0
	for start := 0; start < len(glyphs); {
		end := start
		for end < len(glyphs) {
			end++
			if glyphs[end-1].Flags&text.FlagClusterBreak != 0 {
				break
			}
		}
		run := glyphs[start:end]
		next := runePos + int(run[len(run)-1].Runes)
		c := lb.Color
		if run[0].Flags&text.FlagTruncator == 0 {
			for _, r := range n.textRanges {
				if r.Start < r.End && r.Start < next && r.End > runePos {
					c = r.Color
				}
			}
		}
		for i := start; i < end; i++ {
			colors[i] = c
		}
		runePos = next
		start = end
	}
	for start := 0; start < len(glyphs); {
		end := start + 1
		for end < len(glyphs) && end-start < 32 && glyphs[end-1].Flags&text.FlagLineBreak == 0 && colors[end] == colors[start] {
			end++
		}
		run := glyphs[start:end]
		origin := f32.Pt(float32(run[0].X)/64, float32(run[0].Y))
		offset := op.Affine(f32.AffineId().Offset(origin)).Push(g.Ops)
		outline := clip.Outline{Path: shaper.Shape(run)}.Op().Push(g.Ops)
		paint.Fill(g.Ops, colors[start])
		outline.Pop()
		if bitmap := shaper.Bitmaps(run); bitmap != (op.CallOp{}) {
			bitmap.Add(g.Ops)
		}
		offset.Pop()
		start = end
	}
}
