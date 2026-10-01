package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("radio", "inputs", radioGallery) }
func radioGallery() core.Widget {
	r := widget.RadioGroup("通知频率", "立即", "每天", "每周")
	r.SetValue("每天")
	return layout.Card(widget.Heading("单选组"), r, widget.Checkbox("禁用单选组", false).OnChange(r.SetDisabled))
}
