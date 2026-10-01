package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("link", "controls", linkGallery) }
func linkGallery() core.Widget {
	status := widget.Text("尚未点击")
	link := widget.Link("查看文档", func() { status.SetText("已点击链接") })
	return layout.Card(widget.Heading("链接"), link, status, widget.Checkbox("禁用链接", false).OnChange(link.SetDisabled))
}
