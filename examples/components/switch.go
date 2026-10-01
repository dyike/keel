package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("switch", "controls", switchGallery) }
func switchGallery() core.Widget {
	sync := widget.Switch("同步", true).Size(widget.Large)
	return layout.Card(widget.Heading("开关"), layout.Row(widget.Switch("小", false).Size(widget.Small), widget.Switch("中", true), sync), widget.Checkbox("同步加载中", false).OnChange(sync.SetLoading), widget.Checkbox("禁用同步", false).OnChange(sync.SetDisabled))
}
