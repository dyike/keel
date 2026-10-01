package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("progress", "controls", progressGallery) }
func progressGallery() core.Widget {
	progress := widget.Progress("上传进度")
	progress.SetValue(.4)
	return layout.Card(widget.Heading("确定与不确定进度"), progress, widget.Checkbox("无法确定进度", false).OnChange(progress.SetIndeterminate))
}
