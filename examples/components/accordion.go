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
		small := kit.Accordion().Size(kit.AccordionSizeSmall).Bordered(false).
			Add("Small · 无边框", txt("隐藏外框和分节线，保留背景与圆角。")).
			Add("第二节", txt("尺寸切换不改变展开状态。"))
		small.SetValue(0)
		large := kit.Accordion().Size(kit.AccordionSizeLarge).Add("Large · 大尺寸", txt("标题、箭头、间距与正文一起调整。"))
		large.SetValue(0)
		tiny := kit.Accordion().Size(kit.AccordionSizeXSmall).Add("XSmall · 紧凑", txt("适用于信息密集的面板。"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(480)).MaxW(el.Full).Child(one.Render(cx), faq.Render(cx), small.Render(cx), large.Render(cx), tiny.Render(cx))
		}))
	})
}
