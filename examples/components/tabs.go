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
			Add("基本", kit.Input("名称").Placeholder("切换后内容保留")).
			Add("通知", kit.Switch("邮件提醒", true)).
			Add("关于 About", el.ViewFunc(func(*el.Context) el.Element { return el.Text("Keel 组件示例 1.0") }))
		tabs.Closable(tabs.Remove).Reorderable(nil).Trailing(kit.Button("新增", func() { tabs.Add("新标签", kit.Input("内容")) }).Variant(kit.ButtonGhost))
		variants := []*kit.TabsView{}
		for _, variant := range []kit.TabsVariant{kit.TabsUnderline, kit.TabsPill, kit.TabsOutline, kit.TabsSegmented} {
			variants = append(variants, kit.Tabs().Variant(variant).AddItem(kit.TabItem{Title: "概览", Icon: kit.IconHome}).AddItem(kit.TabItem{Title: "消息", Icon: kit.IconInbox, Content: el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Row().Gap(theme.SpaceSm).Child(el.Text("消息"), el.Text("3").Bold())
			})}).AddItem(kit.TabItem{Title: "已禁用", Icon: kit.IconLock, Disabled: true}))
		}
		scrolling := kit.Tabs().Scrollable(true).MaxWidth(100).Variant(kit.TabsPill)
		for i := range 12 {
			scrolling.Add(fmt.Sprintf("文档 %02d", i+1), nil)
		}
		scrolling.Trailing(kit.Button("最后一项", func() { scrolling.ScrollTo(scrolling.Len() - 1) }).Compact(true))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			out := el.Div().P(24).Gap(theme.SpaceXl).W(el.Dp(460)).MaxW(el.Full).Child(tabs.Render(cx))
			for i, v := range variants {
				out.Child(el.Text([]string{"下划线", "胶囊", "描边", "分段"}[i]).TextColor(theme.Muted), v.Render(cx))
			}
			out.Child(el.Text("横向滚动 · 标签宽度上限 100dp").TextColor(theme.Muted), scrolling.Render(cx))
			return out
		}))
	})
}
