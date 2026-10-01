package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("toggle_group", "controls", toggle_groupGallery) }
func toggle_groupGallery() core.Widget {
	align := widget.ToggleGroup("左", "中", "右").Size(widget.Small)
	align.SetValues("中")
	format := widget.ToggleGroup("粗体", "斜体", "下划线").Multiple().Ghost()
	format.SetValues("粗体", "下划线")
	return layout.Card(widget.Heading("单选与多选状态按钮组"), align, format, widget.Checkbox("禁用单选组", false).OnChange(align.SetDisabled))
}
