package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("accordion", "controls", accordionGallery) }
func accordionGallery() core.Widget {
	a := widget.Accordion().Multiple().Add("个人资料", widget.Input("名称")).Add("通知设置", widget.Switch("接收通知", true))
	a.SetValue([]int{0})
	return layout.Card(widget.Heading("折叠面板"), a, widget.Checkbox("禁用折叠面板", false).OnChange(a.SetDisabled))
}
