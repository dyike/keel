package widget

import (
	"fmt"
	"image"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// ProgressView is a horizontal bar from 0 to 1 with an optional label.
type ProgressView struct {
	label string
	value float32
}

func Progress(label string) *ProgressView { return &ProgressView{label: label} }

// SetValue sets the progress, clamped to [0, 1].
func (p *ProgressView) SetValue(v float32) { p.value = min(max(v, 0), 1) }
func (p *ProgressView) Value() float32     { return p.value }

func (p *ProgressView) Layout(gtx C) D {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	pct := fmt.Sprintf("%d%%", int(p.value*100+0.5))
	return core.Semantic(gtx, func(gtx C) D {
		return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
			giolayout.Rigid(func(gtx C) D {
				if p.label == "" {
					return D{}
				}
				return giolayout.Flex{}.Layout(gtx,
					giolayout.Flexed(1, Muted(p.label).Layout),
					giolayout.Rigid(func(gtx C) D {
						lb := material.Label(theme.Material, theme.SmallSize, pct)
						lb.Color = theme.Muted
						return lb.Layout(gtx)
					}),
				)
			}),
			giolayout.Rigid(giolayout.Spacer{Height: 6}.Layout),
			giolayout.Rigid(func(gtx C) D {
				w, h := gtx.Constraints.Max.X, gtx.Dp(8)
				fillRounded(gtx, theme.Subtle, w, h, h/2)
				if fill := int(float32(w) * p.value); fill > 0 {
					fillRounded(gtx, theme.Primary, max(fill, h), h, h/2)
				}
				return D{Size: image.Pt(w, h)}
			}),
		)
	}, core.Role("progressbar", pct), semantic.LabelOp(p.label))
}
