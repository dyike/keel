package widget

import (
	"image"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// SwitchView is an on/off toggle with a label; the whole row is clickable.
type SwitchView struct {
	label             string
	disabled, loading bool
	size              ComponentSize
	onChange          func(bool)
	b                 widget.Bool
}

func Switch(label string, on bool) *SwitchView {
	s := &SwitchView{label: label}
	s.b.Value = on
	return s
}

func (s *SwitchView) OnChange(fn func(on bool)) *SwitchView { s.onChange = fn; return s }
func (s *SwitchView) Value() bool                           { return s.b.Value }
func (s *SwitchView) SetValue(on bool)                      { s.b.Value = on }

func (s *SwitchView) Size(size ComponentSize) *SwitchView { s.size = size; return s }
func (s *SwitchView) SetDisabled(v bool)                  { s.disabled = v }
func (s *SwitchView) SetLoading(v bool)                   { s.loading = v }
func (s *SwitchView) Loading() bool                       { return s.loading }

func (s *SwitchView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if s.disabled || s.loading {
		gtx = gtx.Disabled()
	}
	before := s.b.Value
	changed := s.b.Update(gtx)
	if s.disabled || s.loading {
		s.b.Value = before
		changed = false
	}
	if changed {
		v := s.b.Value
		gtx.Execute(key.FocusCmd{Tag: &s.b})
		core.Call(gtx, func() {
			if s.onChange != nil {
				s.onChange(v)
			}
		})
	}
	return core.Semantic(gtx, func(gtx C) D {
		return s.b.Layout(gtx, func(gtx C) D {
			semantic.Switch.Add(gtx.Ops)
			semantic.LabelOp(s.label).Add(gtx.Ops)
			return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
				giolayout.Rigid(s.track),
				giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout),
				giolayout.Rigid(func(gtx C) D {
					lb := material.Label(theme.Material, theme.BodySize, s.label)
					lb.Color = theme.Text
					if s.disabled {
						lb.Color = theme.Muted
					}
					return layoutLabel(gtx, lb)
				}),
			)
		})
	}, semantic.Switch, semantic.LabelOp(s.label), semantic.SelectedOp(s.b.Value), semantic.EnabledOp(gtx.Enabled()))
}

func (s *SwitchView) track(gtx C) D {
	width, height := 36, 20
	if s.size == Small {
		width, height = 28, 16
	}
	if s.size == Large {
		width, height = 44, 24
	}
	w, h := gtx.Dp(unit.Dp(width)), gtx.Dp(unit.Dp(height))
	col := theme.Border
	if s.b.Value {
		col = theme.Primary
	}
	if s.disabled {
		col = theme.SubtleHover
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
	if s.loading {
		shift := op.Offset(image.Pt(x, pad)).Push(gtx.Ops)
		loadingIndicator(gtx, unit.Dp(height-4), theme.Primary)
		shift.Pop()
	}
	return D{Size: gtx.Constraints.Constrain(image.Pt(w, h))}
}
