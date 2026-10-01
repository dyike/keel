package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("toggle", "controls", toggleGallery) }
func toggleGallery() core.Widget {
	feedback := widget.Muted("未固定")
	var pin *widget.ToggleView = widget.Toggle("固定面板", false).Icon(widget.Icon(widget.IconCheck)).OnChange(func(v bool) {
		if v {
			feedback.SetText("已固定")
		} else {
			feedback.SetText("未固定")
		}
	})
	disabled := widget.Toggle("禁用且选中", true)
	disabled.SetDisabled(true)
	return layout.Card(widget.Heading("状态按钮"), layout.Row(pin, widget.Toggle("Ghost", false).Ghost(), disabled), feedback)
}
