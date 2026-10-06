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
		core.Bind("demo.command.new", "mod+n")
		msg := demoText("Press ⌘K / Ctrl+K or click the button to open the command palette", "按 ⌘K / Ctrl+K 或点击按钮打开命令面板")
		do := func(s string) func() { return func() { msg = demoText("Executed: ", "已执行：") + s } }
		items := []kit.CommandItem{
			kit.CommandItem{Title: demoText("New order", "新建订单"), Group: demoText("Orders", "订单"), ActionName: "demo.command.new", Icon: kit.IconPlus, Action: do(demoText("New order", "新建订单"))},
			kit.CommandItem{Title: demoText("Export CSV (permission required)", "导出 CSV（暂无权限）"), Group: demoText("Orders", "订单"), Disabled: true, Action: do(demoText("Export CSV", "导出 CSV"))},
			kit.CommandItem{Separator: true},
			kit.CommandItem{Title: demoText("Open settings", "打开设置"), Keywords: []string{"preferences", "settings"}, Group: demoText("Preferences", "偏好"), Shortcut: "mod+,", Action: do(demoText("Open settings", "打开设置"))},
			kit.CommandItem{Title: demoText("Toggle dark mode", "切换深色模式 Dark mode"), Group: demoText("Preferences", "偏好"), Action: func() { theme.Apply(theme.Dark()) }},
			kit.CommandItem{Title: "New window", Shortcut: "mod+shift+n", Action: do("New window")},
		}
		for i := range 10000 {
			title := fmt.Sprintf(demoText("Command %05d", "命令 %05d"), i)
			items = append(items, kit.CommandItem{Title: title, Group: demoText("More", "更多"), Action: do(title)})
		}
		cmd := kit.Command(items...).AutoRowHeight(true).Placeholder(demoText("Search commands or settings", "搜索命令或 settings")).OnCancel(func() { msg = demoText("Command palette closed", "已关闭命令面板") }).OnConfirm(func(index int) { msg += fmt.Sprintf(demoText("(item %d)", "（条目 %d）"), index) }).Header(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().P(12).Child(el.Text(demoText("Search and run commands", "搜索并执行命令")))
		})).Footer(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().P(12).Child(el.Text(demoText("↑↓ selects · Enter runs · Esc closes", "↑↓ 选择 · Enter 执行 · Esc 关闭")).TextColor(theme.Muted))
		}))
		quick := kit.Command(items[:4]...).Searchable(false).Inline(true).AutoRowHeight(true).MaxHeight(240).Bordered(false).
			RenderItem(func(item kit.CommandItem, active bool) el.View {
				return el.ViewFunc(func(*el.Context) el.Element {
					return el.Div().Child(el.Text(item.Title), el.Text(demoText("Quick actions", "快速操作")).TextSize(theme.TextSm).TextColor(theme.Muted))
				})
			})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			cx.Shortcut("mod+k", cmd.Toggle)
			return el.Div().P(24).Gap(12).Items(el.Start).Child(
				kit.Button(demoText("Open command palette", "打开命令面板"), cmd.Toggle).Render(cx), el.Text(msg).TextColor(theme.Muted), quick.Render(cx), cmd.Render(cx))
		}))
	})
}
