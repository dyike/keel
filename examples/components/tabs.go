package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("tabs", "controls", tabsGallery) }
func tabsGallery() core.Widget {
	tabs := widget.Tabs().Add("概览", widget.Text("概览内容")).Add("设置", widget.Input("名称"))
	return layout.Card(widget.Heading("标签页"), tabs, widget.Checkbox("禁用标签页", false).OnChange(tabs.SetDisabled))
}
