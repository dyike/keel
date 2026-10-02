package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("toolbar", "shell", func() core.Widget {
		msg := "缩小窗口，放不下的按钮会移到「更多」里；←→ 在按钮间移动"
		do := func(s string) func() { return func() { msg = "已执行：" + s } }
		bar := kit.Toolbar(
			kit.ToolbarItem{Label: "新建", Icon: kit.IconPlus, HasIcon: true, Action: do("新建")},
			kit.ToolbarItem{Label: "搜索", Icon: kit.IconSearch, HasIcon: true, IconOnly: true, Action: do("搜索")},
			kit.ToolbarItem{Separator: true},
			kit.ToolbarItem{Label: "复制 Copy", Action: do("复制")},
			kit.ToolbarItem{Label: "导出", Action: do("导出")},
			kit.ToolbarItem{Label: "打印", Action: do("打印")},
			kit.ToolbarItem{Label: "分享", Action: do("分享")},
			kit.ToolbarItem{Label: "归档", Action: do("归档"), Disabled: true},
		)
		bar.Leading(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Px(8).Child(el.Text("文件")) })).Trailing(kit.Button("帮助", do("帮助")).Variant(kit.ButtonGhost))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(
				el.Div().Border(1, theme.Border).Rounded(8).P(4).Items(el.Stretch).Child(bar.Render(cx)),
				el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
