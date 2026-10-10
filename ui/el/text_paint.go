package el

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/theme"
	"golang.org/x/image/math/fixed"
)

// paintLabel reuses short vector fragments for single-line numeric labels.
// Gio caches Shape by the complete glyph sequence: a changing counter misses
// that cache every frame. Two-glyph fragments keep common digit combinations
// reusable without introducing a renderer or font dependency fork.
// Complex layout, overlapping ink and translucent text use the standard label.
func (e *engine) paintLabel(gtx layout.Context, lb material.LabelStyle) {
	var buf [64]text.Glyph
	params, gs, viewport, ok := shapeNumericLabel(gtx, lb, buf[:0])
	if !ok {
		lb.Layout(gtx)
		return
	}
	defer clip.Rect(viewport).Push(gtx.Ops).Pop()
	semantic.LabelOp(lb.Text).Add(gtx.Ops)
	if len(gs) > 0 {
		origin := f32.Pt(float32(gs[0].X)/64, float32(gs[0].Y))
		if theme.PlatformText(lb.Shaper) {
			// The platform rasterizer caches glyph masks itself.
			theme.PaintGlyphs(gtx.Ops, lb.Shaper, gs, lb.Color, origin)
			return
		}
		tr := op.Affine(f32.AffineId().Offset(origin)).Push(gtx.Ops)
		e.textPainter.Paint(gtx.Ops, lb.Shaper, params, gs, lb.Color)
		tr.Pop()
	}
}

func labelParameters(gtx layout.Context, lb material.LabelStyle) text.Parameters {
	cs := gtx.Constraints
	return text.Parameters{Font: lb.Font, PxPerEm: fixed.I(gtx.Sp(lb.TextSize)), MaxLines: lb.MaxLines, Truncator: lb.Truncator, Alignment: lb.Alignment, WrapPolicy: lb.WrapPolicy, MinWidth: cs.Min.X, MaxWidth: cs.Max.X, Locale: gtx.Locale, LineHeight: fixed.I(gtx.Sp(lb.LineHeight)), LineHeightScale: lb.LineHeightScale}
}

func shapeNumericLabel(gtx layout.Context, lb material.LabelStyle, buf []text.Glyph) (text.Parameters, []text.Glyph, image.Rectangle, bool) {
	params := labelParameters(gtx, lb)
	numeric := false
	for _, r := range lb.Text {
		if r >= '0' && r <= '9' {
			numeric = true
			break
		}
	}
	if !numeric || len(lb.Text) > 128 || lb.Color.A != 255 || lb.State != nil {
		return params, nil, image.Rectangle{}, false
	}
	cs := gtx.Constraints
	sh := lb.Shaper
	sh.LayoutString(params, lb.Text)
	gs := buf[:0]
	var padding image.Rectangle
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		if len(gs) == 64 || g.Offset != (fixed.Point26_6{}) || g.Flags&(text.FlagTowardOrigin|text.FlagParagraphBreak) != 0 || len(gs) > 0 && g.Y != gs[0].Y {
			return params, nil, image.Rectangle{}, false
		}
		if g.X+g.Advance < 0 || g.X > fixed.I(cs.Max.X) || int(g.Y)+g.Descent.Ceil() < 0 || int(g.Y)-g.Ascent.Ceil() > cs.Max.Y {
			return params, nil, image.Rectangle{}, false
		}
		if len(gs) > 0 {
			prev := gs[len(gs)-1]
			if prev.X+prev.Bounds.Max.X > g.X+g.Bounds.Min.X {
				return params, nil, image.Rectangle{}, false
			}
		}
		padding.Min.X = min(padding.Min.X, g.Bounds.Min.X.Floor())
		padding.Max.X = max(padding.Max.X, (g.Bounds.Max.X - g.Advance).Ceil())
		padding.Min.Y = min(padding.Min.Y, (g.Bounds.Min.Y + g.Ascent).Floor())
		padding.Max.Y = max(padding.Max.Y, (g.Bounds.Max.Y - g.Descent).Ceil())
		gs = append(gs, g)
	}
	viewport := image.Rectangle{Min: padding.Min, Max: cs.Max.Add(padding.Max)}
	return params, gs, viewport, true
}
