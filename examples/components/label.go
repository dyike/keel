package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("label", "inputs", labelGallery) }
func labelGallery() core.Widget {
	field := widget.Input("").Hint("外部标签关联的输入框")
	field.SetName("关联输入")
	return layout.Card(widget.Heading("外部标签聚焦"), widget.Text("点击标签聚焦下面的输入框").For(field), field)
}
