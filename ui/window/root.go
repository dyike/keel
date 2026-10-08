package window

import (
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// root paints the window background and scrolls the content when it is taller
// than the window.
type root struct {
	list        widget.List
	transparent bool
}

func (r *root) Layout(gtx core.C, content core.Widget) core.D {
	if !r.transparent {
		paint.Fill(gtx.Ops, theme.Bg)
	}
	if content == nil {
		return core.D{Size: gtx.Constraints.Max}
	}
	// Content that lays out the whole window itself (el.Root) gets it as is:
	// no padding, no outer scroll.
	if f, ok := content.(interface{ FillsWindow() bool }); ok && f.FillsWindow() {
		gtx.Constraints = layout.Exact(gtx.Constraints.Max)
		return content.Layout(gtx)
	}
	r.list.Axis = layout.Vertical
	return material.List(theme.Material, &r.list).Layout(gtx, 1, func(gtx core.C, _ int) core.D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.UniformInset(24).Layout(gtx, content.Layout)
	})
}
