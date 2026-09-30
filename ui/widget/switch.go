package widget

import (
	"image"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// SwitchView is an on/off toggle with a label; the whole row is clickable.
type SwitchView struct {
	label    string
	onChange func(bool)
	b        widget.Bool
}

func Switch(label string, on bool) *SwitchView {
	s := &SwitchView{label: label}
	s.b.Value = on
	return s
}

func (s *SwitchView) OnChange(fn func(on bool)) *SwitchView { s.onChange = fn; return s }
func (s *SwitchView) Value() bool                           { return s.b.Value }
func (s *SwitchView) SetValue(on bool)                      { s.b.Value = on }

func (s *SwitchView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if s.b.Update(gtx) && s.onChange != nil {
		v := s.b.Value
		core.Call(gtx, func() { s.onChange(v) })
	}
	return s.b.Layout(gtx, func(gtx C) D {
		semantic.Switch.Add(gtx.Ops)
		semantic.LabelOp(s.label).Add(gtx.Ops)
		return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
			giolayout.Rigid(s.track),
			giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout),
			giolayout.Rigid(func(gtx C) D {
				lb := material.Label(theme.Material, theme.BodySize, s.label)
				lb.Color = theme.Text
				return giolayout.Inset{Top: 2 * theme.CJKNudge}.Layout(gtx, lb.Layout)
			}),
		)
	})
}

func (s *SwitchView) track(gtx C) D {
	w, h := gtx.Dp(36), gtx.Dp(20)
	col := theme.Border
	if s.b.Value {
		col = theme.Primary
	}
	fillRounded(gtx, col, w, h, h/2)
	pad := gtx.Dp(2)
	d := h - 2*pad
	x := pad
	if s.b.Value {
		x = w - pad - d
	}
	knob := clip.Ellipse(image.Rect(x, pad, x+d, pad+d)).Op(gtx.Ops)
	paintOp(gtx, theme.Surface, knob)
	return D{Size: image.Pt(w, h)}
}
