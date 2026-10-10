package el

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"github.com/dyike/keel/ui/theme"
	"golang.org/x/image/math/fixed"
)

type textShimmer struct {
	phase, spread float32
	highlight     color.NRGBA
}

// Shimmer paints a highlight across the glyphs, preserving inherited typography.
// Phase is the sweep position (0..1); spread is its half-width relative to the
// text box (0..1). This paint-only primitive does not schedule animation.
func (t *TextEl) Shimmer(phase, spread float32, highlight color.NRGBA) *TextEl {
	if math.IsNaN(float64(phase)) || math.IsInf(float64(phase), 0) || spread <= 0 || spread > 1 {
		return t
	}
	if math.IsNaN(float64(spread)) {
		return t
	}
	t.n.shimmer = &textShimmer{min(max(phase, 0), 1), spread, highlight}
	return t
}

func (e *engine) paintShimmerText(n *Node, g layout.Context, size image.Point) {
	s := n.shimmer
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
		if int(glyph.Y)+glyph.Descent.Ceil() < 0 || (glyph.X+glyph.Advance).Ceil() < 0 || glyph.X.Floor() > size.X {
			continue
		}
		glyphs = append(glyphs, glyph)
	}
	bounds.Min = bounds.Min.Add(padding.Min)
	bounds.Max = bounds.Max.Add(padding.Max)
	defer clip.Rect(bounds).Push(g.Ops).Pop()
	semantic.LabelOp(n.text).Add(g.Ops)
	band := max(float32(size.X)*s.spread, 1)
	center := s.phase*(float32(size.X)+2*band) - band
	split := int(math.Round(float64(center)))
	fill := func(r image.Rectangle, x1, x2 float32, c1, c2 color.NRGBA) {
		if r.Empty() {
			return
		}
		defer clip.Rect(r).Push(g.Ops).Pop()
		paint.LinearGradientOp{Stop1: f32.Pt(x1, 0), Stop2: f32.Pt(x2, 0), Color1: c1, Color2: c2}.Add(g.Ops)
		paint.PaintOp{}.Add(g.Ops)
	}
	platform := theme.PlatformText(shaper)
	for start := 0; start < len(glyphs); {
		end := start + 1
		for end < len(glyphs) && end-start < 32 && glyphs[end-1].Flags&text.FlagLineBreak == 0 && glyphs[end].Y == glyphs[start].Y {
			end++
		}
		run := glyphs[start:end]
		origin := f32.Pt(float32(run[0].X)/64, float32(run[0].Y))
		if platform {
			paintShimmerStripes(g.Ops, shaper, run, origin, bounds, center, band, lb.Color, s.highlight)
			start = end
			continue
		}
		offset := op.Affine(f32.AffineId().Offset(origin)).Push(g.Ops)
		outline := clip.Outline{Path: shaper.Shape(run)}.Op().Push(g.Ops)
		// Gradients use text-box coordinates even when a glyph run begins mid-line.
		undo := op.Affine(f32.AffineId().Offset(origin.Mul(-1))).Push(g.Ops)
		left, right := bounds, bounds
		left.Max.X = min(left.Max.X, split)
		right.Min.X = max(right.Min.X, split)
		fill(left, center-band, center, lb.Color, s.highlight)
		fill(right, center, center+band, s.highlight, lb.Color)
		undo.Pop()
		outline.Pop()
		if bitmap := shaper.Bitmaps(run); bitmap != (op.CallOp{}) {
			bitmap.Add(g.Ops)
		}
		offset.Pop()
		start = end
	}
}

// paintShimmerStripes draws a shimmer with the platform rasterizer, which
// tints glyphs with one color: the base color outside the band, and inside
// it narrow stripes, each with the gradient's color at its center, so the
// text keeps the same glyphs as when it rests.
func paintShimmerStripes(ops *op.Ops, sh *text.Shaper, run []text.Glyph, origin f32.Point, bounds image.Rectangle, center, band float32, base, highlight color.NRGBA) {
	lo, hi := int(math.Floor(float64(center-band))), int(math.Ceil(float64(center+band)))
	plain := func(r image.Rectangle) {
		if r = r.Intersect(bounds); r.Empty() {
			return
		}
		cl := clip.Rect(r).Push(ops)
		theme.PaintGlyphs(ops, sh, run, base, origin)
		cl.Pop()
	}
	plain(image.Rect(bounds.Min.X, bounds.Min.Y, lo, bounds.Max.Y))
	plain(image.Rect(hi, bounds.Min.Y, bounds.Max.X, bounds.Max.Y))
	step := max(1, int(math.Ceil(float64(band)/12)))
	for x := lo; x < hi; x += step {
		r := image.Rect(x, bounds.Min.Y, min(x+step, hi), bounds.Max.Y).Intersect(bounds)
		if r.Empty() {
			continue
		}
		mid := float32(r.Min.X+r.Max.X) / 2
		t := 1 - min(float32(math.Abs(float64(mid-center)))/band, 1)
		cl := clip.Rect(r).Push(ops)
		theme.PaintGlyphs(ops, sh, run, mixLinear(base, highlight, t), origin)
		cl.Pop()
	}
}

// mixLinear interpolates two colors in linear light, as gradients do.
func mixLinear(a, b color.NRGBA, t float32) color.NRGBA {
	ch := func(x, y uint8) uint8 {
		la, lb := srgbLinear(x), srgbLinear(y)
		v := la + (lb-la)*float64(t)
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		return uint8(math.Round(255 * v))
	}
	return color.NRGBA{R: ch(a.R, b.R), G: ch(a.G, b.G), B: ch(a.B, b.B), A: uint8(math.Round(float64(a.A) + float64(int(b.A)-int(a.A))*float64(t)))}
}

func srgbLinear(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}
