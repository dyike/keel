package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("toolbar", "shell", func() core.Widget {
		msg := demoText("Narrow the window to move overflowing buttons into More; ←→ moves between buttons", "缩小窗口，放不下的按钮会移到「更多」里；←→ 在按钮间移动")
		do := func(s string) func() { return func() { msg = demoText("Executed: ", "已执行：") + s } }
		zoom := kit.Select("", "75%", "100%", "125%", "150%").OnChange(func(value string) { msg = demoText("Zoom: ", "缩放：") + value })
		zoom.SetValue("100%")
		bar := kit.Toolbar(
			kit.ToolbarItem{Label: demoText("New", "新建"), Icon: kit.IconPlus, Action: do(demoText("New", "新建"))},
			kit.ToolbarItem{Label: demoText("Search", "搜索"), Icon: kit.IconSearch, IconOnly: true, Action: do(demoText("Search", "搜索"))},
			kit.ToolbarItem{Separator: true},
			kit.ToolbarItem{Label: demoText("Zoom", "缩放"), Width: 150, Content: zoom},
			kit.ToolbarItem{Label: demoText("Copy", "复制 Copy"), Action: do(demoText("Copy", "复制"))},
			kit.ToolbarItem{Label: demoText("Export", "导出"), Action: do(demoText("Export", "导出"))},
			kit.ToolbarItem{Label: demoText("Print", "打印"), Action: do(demoText("Print", "打印"))},
			kit.ToolbarItem{Label: demoText("Share", "分享"), Action: do(demoText("Share", "分享"))},
			kit.ToolbarItem{Label: demoText("Archive", "归档"), Action: do(demoText("Archive", "归档")), Disabled: true},
		)
		bar.Leading(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Px(8).Child(el.Text(demoText("Files", "文件"))) })).Trailing(kit.Button(demoText("Help", "帮助"), do(demoText("Help", "帮助"))).Variant(kit.ButtonGhost))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(
				el.Div().Border(1, theme.Border).Rounded(8).P(4).Items(el.Stretch).Child(bar.Render(cx)),
				el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
