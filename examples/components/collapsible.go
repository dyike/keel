package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("collapsible", "shell", func() core.Widget {
		field := kit.Input(demoText("Order number", "订单编号"))
		field.SetValue("SO-1001")
		details := kit.Collapsible(demoText("Order details", "订单详情"), field)
		details.SetValue(true)
		separate := kit.Collapsible(demoText("Advanced filters", "高级筛选"), kit.Input("").Placeholder(demoText("Customer name", "客户名称"))).Heading(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Gap(8).Child(kit.Icon(kit.IconSearch).Size(16).Render(cx), el.Text(demoText("Advanced filters", "高级筛选")).Bold(), kit.Tag(demoText("Optional", "可选")).Render(cx))
		}))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Items(el.Stretch).Child(
				el.Text("Collapsible").TextSize(24).Bold(), details.Render(cx),
				el.Div().Rounded(8).Border(1, theme.Border).Child(separate.Trigger().Render(cx), el.Text(demoText("Description text can appear between the trigger and content.", "触发器与内容之间可以放说明文字。")).Px(14).TextColor(theme.Muted), separate.Content().Render(cx)),
				el.Text(demoText("Repeated clicks reverse the animation. Reduced motion expands immediately.", "快速重复点击可反转动画；减少动画设置下立即展开。")),
			)
		}))
	})
}
