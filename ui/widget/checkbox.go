package widget

import (
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"image"
)

type Check struct {
	label                          string
	onChange                       func(bool)
	value, disabled, indeterminate bool
	click                          widget.Clickable
}

func Checkbox(label string, checked bool) *Check { return &Check{label: label, value: checked} }
func (c *Check) OnChange(fn func(bool)) *Check   { c.onChange = fn; return c }
func (c *Check) Value() bool                     { return c.value }
func (c *Check) SetValue(v bool)                 { c.value = v; c.indeterminate = false }
func (c *Check) SetDisabled(v bool)              { c.disabled = v }

// SetIndeterminate preserves Value; user activation selects the mixed checkbox.
func (c *Check) SetIndeterminate(v bool) { c.indeterminate = v }
func (c *Check) Indeterminate() bool     { return c.indeterminate }
func (c *Check) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if c.disabled {
		gtx = gtx.Disabled()
	}
	for c.click.Clicked(gtx) {
		if c.disabled {
			continue
		}
		c.value = c.indeterminate || !c.value
		c.indeterminate = false
		gtx.Execute(key.FocusCmd{Tag: &c.click})
		value := c.value
		core.Call(gtx, func() {
			if c.onChange != nil {
				c.onChange(value)
			}
		})
	}
	state := ""
	if c.indeterminate {
		state = "mixed"
	}
	return core.Semantic(gtx, func(gtx C) D {
		return c.click.Layout(gtx, func(gtx C) D {
			return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
				giolayout.Rigid(func(gtx C) D {
					icon, col := theme.Material.Icon.CheckBoxUnchecked, theme.Muted
					if c.value || c.indeterminate {
						icon, col = theme.Material.Icon.CheckBoxChecked, theme.Primary
					}
					if c.disabled {
						col = theme.Muted
					}
					size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(22), gtx.Dp(22)))
					if c.indeterminate {
						fillRounded(gtx, col, size.X, size.Y, gtx.Dp(3))
						paintOp(gtx, theme.OnColor, clip.Rect(image.Rect(size.X/4, size.Y/2-gtx.Dp(1), size.X*3/4, size.Y/2+gtx.Dp(1))).Op())
						return D{Size: size}
					}
					gtx.Constraints.Min, gtx.Constraints.Max = size, size
					return icon.Layout(gtx, col)
				}),
				giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout),
				giolayout.Rigid(func(gtx C) D {
					lb := material.Label(theme.Material, theme.BodySize, c.label)
					lb.Color = theme.Text
					if c.disabled {
						lb.Color = theme.Muted
					}
					return layoutLabel(gtx, lb)
				}),
			)
		})
	}, semantic.CheckBox, semantic.LabelOp(c.label), semantic.SelectedOp(c.value), semantic.EnabledOp(gtx.Enabled()), core.Role("checkbox", state))
}
