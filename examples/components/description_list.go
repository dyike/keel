package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("description_list", "controls", func() core.Widget {
		order := kit.DescriptionList().Columns(2).Vertical().Bordered(true).
			Item("订单号", "SO-123").Item("客户", "张三 / Ada Lovelace").
			Separator().Item("备注", "中英文 mixed description；长值可以换行，这条说明跨两列展示。").Span(2).
			Item("状态", "待发货").Item("渠道", "线上商店")
		basic := kit.DescriptionList().Size(theme.TextSm).Item("项目", "Keel").Item("版本", "0.1.0")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).Child(el.Text("订单信息 · 两列布局"), order.Render(cx), el.Text("紧凑列表"), basic.Render(cx))
		}))
	})
}
