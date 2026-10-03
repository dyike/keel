package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("resizable", "shell", func() core.Widget {
		pane := func(s string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Grow().P(16).Bg(theme.Surface).Child(el.Text(s).TextColor(theme.Muted))
			})
		}
		inner := kit.Resizable(pane("编辑区 Editor"), pane("终端 Terminal")).Vertical().Min(60, 60).Max(240, 240)
		inner.SetValue(180)
		outer := kit.Resizable(pane("拖动中间的分隔条，或聚焦后按方向键"), inner).Min(120, 160).Max(260, 0)
		outer.SetValue(220)
		disabled := false
		showSidebar := true
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(kit.Button("启用 / 禁用分隔面板", func() { disabled = !disabled; outer.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx), kit.Button("显示 / 隐藏侧栏", func() { showSidebar = !showSidebar; outer.Visible(showSidebar, true) }).Render(cx), el.Div().H(el.Dp(360)).Border(1, theme.Border).Items(el.Stretch).Child(outer.Render(cx)))
		}))
	})
}
