package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("description_list", "controls", func() core.Widget {
		return el.Embed(kit.DescriptionList().Item("订单号", "SO-123").Item("客户", "张三 / Ada Lovelace").Item("备注", "中英文 mixed description；长值可以换行。").ItemView("操作", el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().OnClick(func() {}).Child(el.Text("查看订单")) })))
	})
}
