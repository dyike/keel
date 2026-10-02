package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("accordion", "controls", func() core.Widget {
		txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		name := kit.Input("显示名称").Placeholder("收起后保留输入内容")
		one := kit.Accordion().Add("个人资料", name).Add("通知设置", kit.Switch("接收通知", true)).Add("暂不可用", txt("禁用项目"))
		one.Heading(0, el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Gap(8).Child(kit.Icon(kit.IconUser).Size(16).Render(cx), el.Text("个人资料").Bold(), kit.Tag("可编辑").Render(cx))
		}))
		one.SetValue(0)
		one.SetItemDisabled(2, true)
		faq := kit.Accordion().Multiple().
			Add("如何用键盘操作？", txt("Tab 聚焦标题，↑ ↓ 切换标题，Enter 或空格展开，Home / End 跳到首尾。")).
			Add("可以同时展开吗？ Multiple", txt("这个面板启用了 Multiple，可以同时展开两项。"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(480)).Child(one.Render(cx), faq.Render(cx))
		}))
	})
}
