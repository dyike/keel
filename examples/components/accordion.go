package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("accordion", "controls", func() core.Widget {
		txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		name := kit.Input(demoText("Display name", "显示名称")).Placeholder(demoText("Input is preserved when collapsed", "收起后保留输入内容"))
		one := kit.Accordion().Add(demoText("Profile", "个人资料"), name).Add(demoText("Notification settings", "通知设置"), kit.Switch(demoText("Receive notifications", "接收通知"), true)).Add(demoText("Unavailable", "暂不可用"), txt(demoText("Disabled item", "禁用项目")))
		one.Heading(0, el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Gap(8).Child(kit.Icon(kit.IconUser).Size(16).Render(cx), el.Text(demoText("Profile", "个人资料")).Bold(), kit.Tag(demoText("Editable", "可编辑")).Render(cx))
		}))
		one.SetValue(0)
		one.SetItemDisabled(2, true)
		faq := kit.Accordion().Multiple().
			Add(demoText("How do I use the keyboard?", "如何用键盘操作？"), txt(demoText("Tab focuses a heading; ↑ ↓ changes headings; Enter or Space expands; Home / End jumps to the first or last heading.", "Tab 聚焦标题，↑ ↓ 切换标题，Enter 或空格展开，Home / End 跳到首尾。"))).
			Add(demoText("Can multiple items be expanded?", "可以同时展开吗？ Multiple"), txt(demoText("Multiple is enabled, so two items can be expanded at once.", "这个面板启用了 Multiple，可以同时展开两项。")))
		small := kit.Accordion().Size(kit.AccordionSizeSmall).Bordered(false).
			Add(demoText("Small · Borderless", "Small · 无边框"), txt(demoText("Hide the outer border and section dividers; keep the background and rounded corners.", "隐藏外框和分节线，保留背景与圆角。"))).
			Add(demoText("Second section", "第二节"), txt(demoText("Changing size preserves the expanded state.", "尺寸切换不改变展开状态。")))
		small.SetValue(0)
		large := kit.Accordion().Size(kit.AccordionSizeLarge).Add(demoText("Large", "Large · 大尺寸"), txt(demoText("The heading, arrow, spacing, and body adjust together.", "标题、箭头、间距与正文一起调整。")))
		large.SetValue(0)
		tiny := kit.Accordion().Size(kit.AccordionSizeXSmall).Add(demoText("XSmall · Compact", "XSmall · 紧凑"), txt(demoText("Suitable for information-dense panels.", "适用于信息密集的面板。")))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(480)).MaxW(el.Full).Child(one.Render(cx), faq.Render(cx), small.Render(cx), large.Render(cx), tiny.Render(cx))
		}))
	})
}
