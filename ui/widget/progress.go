package widget

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// ProgressView is a horizontal bar from 0 to 1 with an optional label.
type ProgressView struct {
	label         string
	value         float32
	indeterminate bool
}

func Progress(label string) *ProgressView { return &ProgressView{label: label} }

// SetValue sets the progress, clamped to [0, 1].
func (p *ProgressView) SetValue(v float32) {
	if math.IsNaN(float64(v)) {
		v = 0
	}
	p.value = min(max(v, 0), 1)
}
func (p *ProgressView) Value() float32          { return p.value }
func (p *ProgressView) SetIndeterminate(v bool) { p.indeterminate = v }
func (p *ProgressView) Indeterminate() bool     { return p.indeterminate }

func (p *ProgressView) Layout(gtx C) D {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	pct := fmt.Sprintf("%d%%", int(p.value*100+0.5))
	if p.indeterminate {
		pct = "indeterminate"
	}
	return core.Semantic(gtx, func(gtx C) D {
		return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
			giolayout.Rigid(func(gtx C) D {
				if p.label == "" {
					return D{}
				}
				return giolayout.Flex{}.Layout(gtx,
					giolayout.Flexed(1, Muted(p.label).Layout),
					giolayout.Rigid(func(gtx C) D {
						if p.indeterminate {
							return D{}
						}
						lb := material.Label(theme.Material, theme.SmallSize, pct)
						lb.Color = theme.Muted
						return layoutLabel(gtx, lb)
					}),
				)
			}),
			giolayout.Rigid(giolayout.Spacer{Height: 6}.Layout),
			giolayout.Rigid(func(gtx C) D {
				w, h := gtx.Constraints.Max.X, min(gtx.Constraints.Max.Y, gtx.Dp(8))
				fillRounded(gtx, theme.Subtle, w, h, h/2)
				defer clip.Rect(image.Rect(0, 0, w, h)).Push(gtx.Ops).Pop()
				if p.indeterminate {
					phase := float32(gtx.Now.UnixNano()%int64(2*time.Second)) / float32(2*time.Second)
					length := max(w/3, h)
					x := int(phase*float32(w+length)) - length
					defer op.Offset(image.Pt(x, 0)).Push(gtx.Ops).Pop()
					fillRounded(gtx, theme.Primary, length, h, h/2)
					gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 30)})
				} else if fill := int(float32(w) * p.value); fill > 0 {
					fillRounded(gtx, theme.Primary, fill, h, min(fill, h)/2)
				}
				return D{Size: image.Pt(w, h)}
			}),
		)
	}, core.Role("progressbar", pct), semantic.LabelOp(p.label))
}
