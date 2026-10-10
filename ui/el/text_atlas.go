package el

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/theme"
)

const maxAtlasLabelGlyphs = 32768

type atlasLabelKey struct {
	textMeasureKey
	constraints layout.Constraints
	color       color.NRGBA
}

type atlasLabel struct {
	key      atlasLabelKey
	glyphs   []text.Glyph
	viewport image.Rectangle
}

func labelAtlasKey(gtx layout.Context, lb material.LabelStyle) atlasLabelKey {
	return atlasLabelKey{textMeasureKey{lb.Shaper, labelParameters(gtx, lb), lb.Text}, gtx.Constraints, lb.Color}
}

// Prepare visible ordinary labels once, then commit all image pages before
// painting any of them. Shaped runs live for this frame only.
func (e *engine) prepareTextAtlas(tree *Node) {
	e.textAtlasActive = false
	clear(e.atlasLabels)
	e.atlasGlyphs = e.atlasGlyphs[:0]
	// The platform rasterizer caches its own glyph images; vector atlas masks
	// beside it would draw labels differently from their neighbours.
	if e.textAtlas == nil || theme.PlatformText(theme.Material.Shaper) {
		return
	}
	e.textAtlas.BeginFrame(theme.Material.Shaper)
	e.prepareAtlasNode(tree, image.Point{}, e.visible)
	e.textAtlas.Commit()
	e.textAtlasActive = true
}

func (e *engine) prepareAtlasNode(n *Node, origin image.Point, visible image.Rectangle) {
	if n.style.hidden || n.decorate != nil || n.style.scrollX || n.style.scrollY {
		return
	}
	pos := origin.Add(n.pos).Add(image.Pt(e.dp(n.style.translate[0]), e.dp(n.style.translate[1])))
	abs := image.Rectangle{Min: pos, Max: pos.Add(n.size)}
	if !n.style.absolute && !abs.Overlaps(visible) {
		return
	}
	if n.style.revealSet {
		visible = visible.Intersect(abs)
	}
	if n.isText {
		if n.shimmer != nil || len(n.textRanges) != 0 || len(e.atlasLabels) >= 2048 || len(e.atlasGlyphs)+64 > maxAtlasLabelGlyphs {
			return
		}
		pl, pt, pr, pb := e.edges(n.style.pad)
		bw := e.dp(n.style.borderWidth)
		g := e.gtx
		g.Constraints = layout.Constraints{Max: n.size.Sub(image.Pt(pl+pr+2*bw, pt+pb+2*bw))}
		lb := e.label(n, n.text)
		if cap(e.atlasGlyphs)-len(e.atlasGlyphs) < 64 {
			buf := make([]text.Glyph, len(e.atlasGlyphs), min(max(1024, cap(e.atlasGlyphs)*2), maxAtlasLabelGlyphs))
			copy(buf, e.atlasGlyphs)
			e.atlasGlyphs = buf
		}
		params, gs, viewport, ok := shapeNumericLabel(g, lb, e.atlasGlyphs[len(e.atlasGlyphs):])
		// Alignment can give the whole label a fractional origin. Preserve
		// its original vector antialiasing rather than rounding the origin.
		if !ok || len(gs) == 0 || gs[0].X%64 != 0 {
			return
		}
		e.atlasGlyphs = e.atlasGlyphs[:len(e.atlasGlyphs)+len(gs)]
		e.textAtlas.Prepare(params, gs, lb.Color)
		if e.atlasLabels == nil {
			e.atlasLabels = make(map[*Node]atlasLabel)
		}
		e.atlasLabels[n] = atlasLabel{labelAtlasKey(g, lb), gs, viewport}
		return
	}
	if n.widget != nil || n.input != nil {
		return
	}
	for _, child := range n.children {
		e.prepareAtlasNode(child.node(), pos, visible)
	}
}

func (e *engine) paintAtlasLabel(n *Node, gtx layout.Context, lb material.LabelStyle) bool {
	if !e.textAtlasActive {
		return false
	}
	p, ok := e.atlasLabels[n]
	// Paint-time hover, focus, disabled, theme and padding changes may
	// differ from layout. Only use a run whose full drawing inputs match.
	if !ok || lb.State != nil || p.key != labelAtlasKey(gtx, lb) {
		return false
	}
	defer clip.Rect(p.viewport).Push(gtx.Ops).Pop()
	semantic.LabelOp(lb.Text).Add(gtx.Ops)
	g := p.glyphs[0]
	tr := op.Offset(image.Pt(g.X.Round(), int(g.Y))).Push(gtx.Ops)
	e.textAtlas.Paint(gtx.Ops, p.key.params, p.glyphs, lb.Color)
	tr.Pop()
	return true
}
