package widget

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// Label is read-only text that wraps to the available width.
type Label struct {
	text  string
	size  unit.Sp
	bold  bool
	color *color.NRGBA
}

func Text(s string) *Label    { return &Label{text: s, size: theme.BodySize} }
func Heading(s string) *Label { return &Label{text: s, size: theme.HeadingSize, bold: true} }

// Muted is smaller, secondary text.
func Muted(s string) *Label { return &Label{text: s, size: theme.SmallSize, color: &theme.Muted} }

func (l *Label) Text() string     { return l.text }
func (l *Label) SetText(s string) { l.text = s }

func (l *Label) Layout(gtx C) D {
	lb := material.Label(theme.Material, l.size, l.text)
	lb.Color = theme.Text
	if l.color != nil {
		lb.Color = *l.color
	}
	if l.bold {
		lb.Font.Weight = font.Bold
	}
	// Gio's label node spans the whole available height; this one has the
	// text's real bounds, and the inner node becomes its child.
	return core.Semantic(gtx, func(gtx C) D { return layoutLabel(gtx, lb) }, semantic.LabelOp(l.text))
}

// Center visible font ink in the logical line box. Keep Gio's line spacing and
// dimensions (including blank lines), moving only the paint and baseline.
// A stable probe prevents the text from jumping when its contents change.
func layoutLabel(gtx C, lb material.LabelStyle) D {
	dy := labelShift(gtx, lb)
	m := op.Record(gtx.Ops)
	d := lb.Layout(gtx)
	call := m.Stop()
	defer op.Offset(image.Pt(0, dy)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	d.Baseline -= dy
	return d
}

func labelShift(gtx C, lb material.LabelStyle) int {
	lb.Shaper.LayoutString(text.Parameters{Font: lb.Font, PxPerEm: fixed.I(gtx.Sp(lb.TextSize)), MaxWidth: 1 << 20}, "国Ag")
	var ascent, descent, top, bottom int
	for {
		g, ok := lb.Shaper.NextGlyph()
		if !ok {
			break
		}
		ascent = max(ascent, g.Ascent.Ceil())
		descent = max(descent, g.Descent.Ceil())
		top = min(top, g.Bounds.Min.Y.Floor())
		bottom = max(bottom, g.Bounds.Max.Y.Ceil())
	}
	return (descent - ascent - top - bottom) / 2
}
