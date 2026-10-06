package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("tabs", "controls", func() core.Widget {
		tabs := kit.Tabs().
			Add(demoText("General", "基本"), kit.Input(demoText("Name", "名称")).Placeholder(demoText("Content is preserved after switching", "切换后内容保留"))).
			Add(demoText("Notifications", "通知"), kit.Switch(demoText("Email reminders", "邮件提醒"), true)).
			Add(demoText("About", "关于 About"), el.ViewFunc(func(*el.Context) el.Element {
				return el.Text(demoText("Keel component examples 1.0", "Keel 组件示例 1.0"))
			}))
		tabs.Closable(tabs.Remove).Reorderable(nil).Trailing(kit.Button(demoText("Add", "新增"), func() { tabs.Add(demoText("New tag", "新标签"), kit.Input(demoText("Content", "内容"))) }).Variant(kit.ButtonGhost))
		variants := []*kit.TabsView{}
		for _, variant := range []kit.TabsVariant{kit.TabsUnderline, kit.TabsPill, kit.TabsOutline, kit.TabsSegmented} {
			variants = append(variants, kit.Tabs().Variant(variant).AddItem(kit.TabItem{Title: demoText("Overview", "概览"), Icon: kit.IconHome}).AddItem(kit.TabItem{Title: demoText("Message", "消息"), Icon: kit.IconInbox, Content: el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Row().Gap(theme.SpaceSm).Child(el.Text(demoText("Message", "消息")), el.Text("3").Bold())
			})}).AddItem(kit.TabItem{Title: demoText("Disabled", "已禁用"), Icon: kit.IconLock, Disabled: true}))
		}
		scrolling := kit.Tabs().Scrollable(true).MaxWidth(100).Variant(kit.TabsPill)
		for i := range 12 {
			scrolling.Add(fmt.Sprintf(demoText("Document %02d", "文档 %02d"), i+1), nil)
		}
		scrolling.Trailing(kit.Button(demoText("Last item", "最后一项"), func() { scrolling.ScrollTo(scrolling.Len() - 1) }).Compact(true))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			out := el.Div().P(24).Gap(theme.SpaceXl).W(el.Dp(460)).MaxW(el.Full).Child(tabs.Render(cx))
			for i, v := range variants {
				out.Child(el.Text([]string{demoText("Underline", "下划线"), demoText("Pill", "胶囊"), demoText("Outline", "描边"), demoText("Segmented", "分段")}[i]).TextColor(theme.Muted), v.Render(cx))
			}
			out.Child(el.Text(demoText("Horizontal scrolling · Labels limited to 100dp", "横向滚动 · 标签宽度上限 100dp")).TextColor(theme.Muted), scrolling.Render(cx))
			return out
		}))
	})
}
