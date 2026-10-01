package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("table", "controls", tableGallery) }
func tableGallery() core.Widget {
	t := widget.Table(widget.Col("名称", 1), widget.Col("数量", 1)).Height(160)
	t.SetRows([][]string{{"橙子", "12"}, {"苹果", "5"}, {"香蕉", "9"}})
	return layout.Card(widget.Heading("表格"), t, widget.Checkbox("禁用表格", false).OnChange(t.SetDisabled))
}
