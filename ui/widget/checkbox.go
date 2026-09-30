package widget

import (
	"image"

	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

type Check struct {
	label    string
	onChange func(bool)
	b        widget.Bool
}

func Checkbox(label string, checked bool) *Check {
	c := &Check{label: label}
	c.b.Value = checked
	return c
}

func (c *Check) OnChange(fn func(bool)) *Check { c.onChange = fn; return c }
func (c *Check) Value() bool                   { return c.b.Value }
func (c *Check) SetValue(v bool)               { c.b.Value = v }

func (c *Check) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if c.b.Update(gtx) && c.onChange != nil {
		core.Call(gtx, func() { c.onChange(c.b.Value) })
	}
	return c.b.Layout(gtx, func(gtx C) D {
		icon, col := theme.Material.Icon.CheckBoxUnchecked, theme.Muted
		if c.b.Value {
			icon, col = theme.Material.Icon.CheckBoxChecked, theme.Primary
		}
		return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
			giolayout.Rigid(func(gtx C) D {
				gtx.Constraints.Min = image.Pt(gtx.Dp(22), gtx.Dp(22))
				return icon.Layout(gtx, col)
			}),
			giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout),
			giolayout.Rigid(func(gtx C) D {
				// Flex centers the padded label, so only half of the inset moves the text.
				lb := material.Label(theme.Material, theme.BodySize, c.label)
				lb.Color = theme.Text
				return giolayout.Inset{Top: 2 * theme.CJKNudge}.Layout(gtx, lb.Layout)
			}),
		)
	})
}
