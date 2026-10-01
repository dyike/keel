package widget

import (
	"image"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// RadioView chooses exactly one of a few options, all visible at once.
type RadioView struct {
	label      string
	options    []string
	horizontal bool
	disabled   bool
	onChange   func(string)
	enum       widget.Enum
}

// RadioGroup creates a vertical group; label may be empty. Nothing is chosen at first.
func RadioGroup(label string, options ...string) *RadioView {
	return &RadioView{label: label, options: options}
}

func (r *RadioView) Horizontal() *RadioView                    { r.horizontal = true; return r }
func (r *RadioView) OnChange(fn func(value string)) *RadioView { r.onChange = fn; return r }
func (r *RadioView) Value() string                             { return r.enum.Value }

// SetValue chooses v without calling OnChange.
func (r *RadioView) SetValue(v string)  { r.enum.Value = v }
func (r *RadioView) SetDisabled(v bool) { r.disabled = v }

func (r *RadioView) Layout(gtx C) D {
	if r.disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min = image.Point{}
	if r.enum.Update(gtx) && r.onChange != nil {
		v := r.enum.Value
		core.Call(gtx, func() { r.onChange(v) })
	}
	axis, gap := giolayout.Vertical, giolayout.Spacer{Height: 6}
	if r.horizontal {
		axis, gap = giolayout.Horizontal, giolayout.Spacer{Width: 16}
	}
	var items []giolayout.FlexChild
	if r.label != "" {
		items = append(items, giolayout.Rigid(Muted(r.label).Layout), giolayout.Rigid(giolayout.Spacer{Height: 6, Width: 12}.Layout))
	}
	for i, o := range r.options {
		if i > 0 {
			items = append(items, giolayout.Rigid(gap.Layout))
		}
		items = append(items, giolayout.Rigid(func(gtx C) D { return r.option(gtx, o) }))
	}
	d := giolayout.Flex{Axis: axis, Alignment: giolayout.Middle}.Layout(gtx, items...)
	// Enum registers each option's input filters in the Update that runs as
	// the next option is laid out, so the last option would have none in the
	// first frame, and Gio drops areas without handlers from hit testing and
	// the semantic tree. A final Update covers every option.
	if r.enum.Update(gtx) && r.onChange != nil {
		v := r.enum.Value
		core.Call(gtx, func() { r.onChange(v) })
	}
	return d
}

func (r *RadioView) option(gtx C, o string) D {
	on := r.enum.Value == o
	return core.Semantic(gtx, func(gtx C) D {
		return r.enum.Layout(gtx, o, func(gtx C) D {
			icon, col := theme.Material.Icon.RadioUnchecked, theme.Muted
			if on {
				icon, col = theme.Material.Icon.RadioChecked, theme.Primary
			}
			if !gtx.Enabled() {
				col = theme.Muted
			}
			return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
				giolayout.Rigid(func(gtx C) D {
					gtx.Constraints.Min = image.Pt(gtx.Dp(22), gtx.Dp(22))
					return icon.Layout(gtx, col)
				}),
				giolayout.Rigid(giolayout.Spacer{Width: 6}.Layout),
				giolayout.Rigid(func(gtx C) D {
					lb := material.Label(theme.Material, theme.BodySize, o)
					lb.Color = theme.Text
					if !gtx.Enabled() {
						lb.Color = theme.Muted
					}
					return giolayout.Inset{Top: 2 * theme.CJKNudge}.Layout(gtx, lb.Layout)
				}),
			)
		})
	}, semantic.RadioButton, semantic.LabelOp(o), semantic.SelectedOp(on), semantic.EnabledOp(gtx.Enabled()))
}
