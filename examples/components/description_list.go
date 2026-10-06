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
			Item(demoText("Order number", "订单号"), "SO-123").Item(demoText("Customer", "客户"), demoText("Alex / Ada Lovelace", "张三 / Ada Lovelace")).
			Separator().Item(demoText("Notes", "备注"), demoText("Mixed description; long values wrap. This description spans two columns.", "中英文 mixed description；长值可以换行，这条说明跨两列展示。")).Span(2).
			Item(demoText("Status", "状态"), demoText("Awaiting shipment", "待发货")).Item(demoText("Channel", "渠道"), demoText("Online store", "线上商店"))
		basic := kit.DescriptionList().Size(theme.TextSm).Item(demoText("Project", "项目"), "Keel").Item(demoText("Version", "版本"), "0.1.0")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).Child(el.Text(demoText("Order information · Two-column layout", "订单信息 · 两列布局")), order.Render(cx), el.Text(demoText("Compact list", "紧凑列表")), basic.Render(cx))
		}))
	})
}
