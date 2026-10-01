package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("command", "overlays", func() core.Widget {
		msg := "按 ⌘K / Ctrl+K 或点击按钮打开命令面板"
		do := func(s string) func() { return func() { msg = "已执行：" + s } }
		cmd := kit.Command(
			kit.CommandItem{Title: "新建订单", Group: "订单", Shortcut: "mod+n", Action: do("新建订单")},
			kit.CommandItem{Title: "导出 CSV", Group: "订单", Action: do("导出 CSV")},
			kit.CommandItem{Title: "打开设置", Group: "偏好", Shortcut: "mod+,", Action: do("打开设置")},
			kit.CommandItem{Title: "切换深色模式 Dark mode", Group: "偏好", Action: func() { theme.Apply(theme.Dark()) }},
			kit.CommandItem{Title: "New window", Shortcut: "mod+shift+n", Action: do("New window")},
		)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			cx.Shortcut("mod+k", cmd.Toggle)
			return el.Div().P(24).Gap(12).Items(el.Start).Child(
				kit.Button("打开命令面板", cmd.Toggle).Render(cx), el.Text(msg).TextColor(theme.Muted), cmd.Render(cx))
		}))
	})
}
