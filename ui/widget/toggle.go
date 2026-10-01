package widget

import (
	"image"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// ToggleButton is a pressable button with a persistent selected state.
type ToggleButton struct {
	text         string
	on, disabled bool
	click        widget.Clickable
	onChange     func(bool)
	icon         *IconView
	size         ComponentSize
	ghost        bool
}

func Toggle(text string, on bool) *ToggleButton { return &ToggleButton{text: text, on: on} }
func (t *ToggleButton) Value() bool             { return t.on }

// SetValue updates the state without calling OnChange.
func (t *ToggleButton) SetValue(on bool)                      { t.on = on }
func (t *ToggleButton) SetDisabled(disabled bool)             { t.disabled = disabled }
func (t *ToggleButton) OnChange(fn func(bool)) *ToggleButton  { t.onChange = fn; return t }
func (t *ToggleButton) Icon(icon *IconView) *ToggleButton     { t.icon = icon; return t }
func (t *ToggleButton) Size(size ComponentSize) *ToggleButton { t.size = size; return t }
func (t *ToggleButton) Ghost() *ToggleButton                  { t.ghost = true; return t }

func (t *ToggleButton) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if t.disabled {
		gtx = gtx.Disabled()
	}
	for t.click.Clicked(gtx) {
		if t.disabled {
			continue
		}
		t.on = !t.on
		gtx.Execute(key.FocusCmd{Tag: &t.click})
		core.Call(gtx, func() {
			if t.onChange != nil {
				t.onChange(t.on)
			}
		})
	}
	bg, fg, border := theme.Surface, theme.Text, theme.Border
	if t.ghost {
		border = bg
	}
	if t.click.Hovered() {
		bg = theme.Subtle
	}
	if t.on {
		bg, fg, border = theme.Highlight, theme.Primary, theme.Primary
	}
	if gtx.Focused(&t.click) {
		border = theme.Primary
	}
	if !gtx.Enabled() {
		bg, fg, border = theme.Subtle, theme.Muted, theme.Border
		if t.on {
			bg = theme.Highlight
		}
	}
	textSize, padding, iconSize := t.size.metrics()
	return core.Semantic(gtx, func(gtx C) D {
		return t.click.Layout(gtx, func(gtx C) D {
			return widget.Border{Color: border, CornerRadius: 6, Width: 1}.Layout(gtx, func(gtx C) D {
				return giolayout.Background{}.Layout(gtx, func(gtx C) D {
					fillRounded(gtx, bg, gtx.Constraints.Min.X, gtx.Constraints.Min.Y, gtx.Dp(6))
					return D{Size: gtx.Constraints.Min}
				}, func(gtx C) D {
					return giolayout.Inset{Top: padding, Bottom: padding, Left: padding * 1.5, Right: padding * 1.5}.Layout(gtx, func(gtx C) D {
						var children []giolayout.FlexChild
						if t.icon != nil {
							children = append(children, giolayout.Rigid(func(gtx C) D { ic := *t.icon; ic.size = iconSize; ic.color = &fg; return ic.Layout(gtx) }))
							if t.text != "" {
								children = append(children, giolayout.Rigid(giolayout.Spacer{Width: 6}.Layout))
							}
						}
						children = append(children, giolayout.Rigid(func(gtx C) D {
							st := material.Label(theme.Material, textSize, t.text)
							st.Color = fg
							return layoutLabel(gtx, st)
						}))
						return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx, children...)
					})
				})
			})
		})
	}, semantic.Button, core.Role("toggle"), semantic.LabelOp(t.text), semantic.SelectedOp(t.on), semantic.EnabledOp(gtx.Enabled()))
}
