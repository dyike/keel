package widget

import (
	"image"
	"image/color"
	"strconv"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// BadgeView displays a count or dot, optionally above a child's top right corner.
// Counts <= 0 hide the badge. Dot mode uses the same visibility rule.
type BadgeView struct {
	count, max             int
	dot                    bool
	child                  core.Widget
	size                   ComponentSize
	icon                   *IconView
	background, foreground *color.NRGBA
}

func Badge(count int) *BadgeView { return &BadgeView{count: count, max: 99} }
func (b *BadgeView) Max(max int) *BadgeView {
	if max > 0 {
		b.max = max
	}
	return b
}
func (b *BadgeView) Dot() *BadgeView                    { b.dot = true; return b }
func (b *BadgeView) Child(child core.Widget) *BadgeView { b.child = child; return b }
func (b *BadgeView) SetCount(count int)                 { b.count = count }
func (b *BadgeView) Count() int                         { return b.count }
func (b *BadgeView) Size(size ComponentSize) *BadgeView { b.size = size; return b }
func (b *BadgeView) Icon(icon *IconView) *BadgeView     { b.icon = icon; b.dot = false; return b }
func (b *BadgeView) Color(background, foreground color.NRGBA) *BadgeView {
	b.background = &background
	b.foreground = &foreground
	return b
}
func (b *BadgeView) label() string {
	if b.count > b.max {
		return strconv.Itoa(b.max) + "+"
	}
	return strconv.Itoa(b.count)
}
func (b *BadgeView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if b.count <= 0 {
		if b.child != nil {
			return b.child.Layout(gtx)
		}
		return D{}
	}
	rec := op.Record(gtx.Ops)
	bg, fg := theme.Danger, theme.OnColor
	if b.background != nil {
		bg = *b.background
	}
	if b.foreground != nil {
		fg = *b.foreground
	}
	textSize, _, iconSize := b.size.metrics()
	textSize -= 2
	badge := core.Semantic(gtx, func(gtx C) D {
		return giolayout.Background{}.Layout(gtx, func(gtx C) D {
			s := gtx.Constraints.Min
			fillRounded(gtx, bg, s.X, s.Y, s.Y/2)
			return D{Size: s}
		}, func(gtx C) D {
			if b.dot {
				dp := iconSize / 2
				return D{Size: gtx.Constraints.Constrain(image.Pt(gtx.Dp(dp), gtx.Dp(dp)))}
			}
			if b.icon != nil {
				return giolayout.UniformInset(2).Layout(gtx, func(gtx C) D { ic := *b.icon; ic.size = iconSize; ic.color = &fg; return ic.Layout(gtx) })
			}
			return giolayout.Inset{Left: 6, Right: 6, Top: 2, Bottom: 2}.Layout(gtx, func(gtx C) D {
				st := material.Label(theme.Material, textSize, b.label())
				st.Color = fg
				return layoutLabel(gtx, st)
			})
		})
	}, semantic.LabelOp(strconv.Itoa(b.count)))
	call := rec.Stop()
	if b.child == nil {
		call.Add(gtx.Ops)
		return badge
	}
	// Reserve space for the overhang so a parent's clip cannot cut off the badge.
	pad := image.Pt((badge.Size.X+1)/2, (badge.Size.Y+1)/2)
	cg := gtx
	cg.Constraints.Max.X = max(0, cg.Constraints.Max.X-pad.X)
	cg.Constraints.Max.Y = max(0, cg.Constraints.Max.Y-pad.Y)
	shift := op.Offset(image.Pt(0, pad.Y)).Push(gtx.Ops)
	child := b.child.Layout(cg)
	shift.Pop()
	size := gtx.Constraints.Constrain(image.Pt(max(child.Size.X+pad.X, badge.Size.X), max(child.Size.Y+pad.Y, badge.Size.Y)))
	shift = op.Offset(image.Pt(max(0, size.X-badge.Size.X), 0)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	shift.Pop()
	return D{Size: size, Baseline: child.Baseline + size.Y - pad.Y - child.Size.Y}
}
