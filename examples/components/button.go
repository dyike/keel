package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("button", "controls", buttonGallery) }
func buttonGallery() core.Widget {
	feedback := widget.Muted("点击按钮查看反馈")
	n := 0
	action := func() { n++; feedback.SetText(fmt.Sprintf("点击 %d 次", n)) }
	save := widget.Button("保存", action).Icon(widget.Icon(widget.IconCheck))
	return layout.Card(widget.Heading("按钮尺寸、图标和加载"), layout.Row(widget.Button("小", action).Size(widget.Small), save, widget.Button("大", action).Size(widget.Large).Secondary()), widget.Checkbox("保存按钮加载中", false).OnChange(save.SetLoading), feedback)
}
