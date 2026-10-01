package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("select", "inputs", selectGallery) }
func selectGallery() core.Widget {
	s := widget.Select("状态", "待处理", "处理中", "已完成")
	s.SetValue("待处理")
	return layout.Card(widget.Heading("下拉选择"), s, widget.Checkbox("禁用下拉选择", false).OnChange(s.SetDisabled))
}
