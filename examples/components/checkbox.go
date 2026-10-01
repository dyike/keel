package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("checkbox", "controls", checkboxGallery) }
func checkboxGallery() core.Widget {
	mixed := widget.Checkbox("半选：点击后选中全部", false)
	mixed.SetIndeterminate(true)
	disabled := widget.Checkbox("禁用复选框", true)
	disabled.SetDisabled(true)
	return layout.Card(widget.Heading("复选框"), mixed, disabled, widget.Checkbox("禁用半选项", false).OnChange(mixed.SetDisabled))
}
