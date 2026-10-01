package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("input", "inputs", inputGallery) }
func inputGallery() core.Widget {
	search := widget.Input("搜索").Hint("点击标签也能聚焦").Prefix(widget.Icon(widget.IconSearch)).Suffix(widget.Kbd("mod+k").Plain()).Clearable()
	feedback := widget.Muted("尚未编辑")
	search.OnChange(func(v string) { feedback.SetText(fmt.Sprintf("搜索内容：%q", v)) })
	price := widget.Input("金额").Prefix(widget.Text("¥")).Suffix(widget.Muted("元")).Clearable()
	price.SetValue("128.00")
	ro := widget.Input("只读").Clearable()
	ro.SetValue("无法清空")
	ro.SetReadOnly(true)
	locked := widget.Input("禁用").Clearable()
	locked.SetValue("不会接收输入")
	locked.SetDisabled(true)
	return layout.Card(widget.Heading("输入框"), search, feedback, widget.Checkbox("禁用搜索", false).OnChange(search.SetDisabled), widget.Checkbox("搜索只读", false).OnChange(search.SetReadOnly), price, ro, locked)
}
