package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("command", "overlays", func() core.Widget {
		msg := "按 ⌘K / Ctrl+K 或点击按钮打开命令面板"
		do := func(s string) func() { return func() { msg = "已执行：" + s } }
		items := []kit.CommandItem{
			kit.CommandItem{Title: "新建订单", Group: "订单", Shortcut: "mod+n", Action: do("新建订单")},
			kit.CommandItem{Title: "导出 CSV（暂无权限）", Group: "订单", Disabled: true, Action: do("导出 CSV")},
			kit.CommandItem{Title: "打开设置", Group: "偏好", Shortcut: "mod+,", Action: do("打开设置")},
			kit.CommandItem{Title: "切换深色模式 Dark mode", Group: "偏好", Action: func() { theme.Apply(theme.Dark()) }},
			kit.CommandItem{Title: "New window", Shortcut: "mod+shift+n", Action: do("New window")},
		}
		for i := range 10000 {
			title := fmt.Sprintf("命令 %05d", i)
			items = append(items, kit.CommandItem{Title: title, Group: "更多", Action: do(title)})
		}
		cmd := kit.Command(items...).Header(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().P(12).Child(el.Text("搜索并执行命令"))
		})).Footer(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().P(12).Child(el.Text("↑↓ 选择 · Enter 执行 · Esc 关闭").TextColor(theme.Muted))
		}))
		quick := kit.Command(items[:3]...).Searchable(false).Inline(true).RowHeight(48).
			RenderItem(func(item kit.CommandItem, active bool) el.View {
				return el.ViewFunc(func(*el.Context) el.Element {
					return el.Div().Child(el.Text(item.Title), el.Text("快速操作").TextSize(theme.TextSm).TextColor(theme.Muted))
				})
			})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			cx.Shortcut("mod+k", cmd.Toggle)
			return el.Div().P(24).Gap(12).Items(el.Start).Child(
				kit.Button("打开命令面板", cmd.Toggle).Render(cx), el.Text(msg).TextColor(theme.Muted), quick.Render(cx), cmd.Render(cx))
		}))
	})
}
