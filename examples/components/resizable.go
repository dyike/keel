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
				return el.Div().Grow().ScrollY().P(16).Bg(theme.Surface).Child(el.Text(s).TextColor(theme.Muted))
			})
		}
		inner := kit.Resizable(pane(demoText("Editor", "编辑区")), pane(demoText("Terminal", "终端"))).Vertical().Min(60, 60).Max(240, 240)
		inner.SetValue(180)
		outer := kit.Resizable(pane(demoText("Drag the divider, or focus it and use the arrow keys", "拖动中间的分隔条，或聚焦后按方向键")), inner).Min(120, 160).Max(260, 0)
		outer.SetValue(220)
		disabled := false
		showSidebar := true
		compact := false
		sizes := [2]float32{220, 180}
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			width, _ := cx.ViewportSize()
			narrow := width < 640*max(float32(1), cx.TextScale())
			if narrow != compact {
				old, next := 0, 1
				if compact {
					old, next = 1, 0
				}
				sizes[old] = outer.Value()
				outer.SetVertical(narrow).SetValue(sizes[next])
				compact = narrow
			}
			height := float32(360)
			if compact {
				height = 600 * max(float32(1), cx.TextScale())
			}
			toggleLabel := demoText("Disable panels", "禁用面板")
			if disabled {
				toggleLabel = demoText("Enable panels", "启用面板")
			}
			sidebarLabel := demoText("Hide sidebar", "隐藏侧栏")
			if !showSidebar {
				sidebarLabel = demoText("Show sidebar", "显示侧栏")
			}
			return el.Div().ScrollY().P(24).Gap(12).Items(el.Stretch).Child(kit.Button(toggleLabel, func() { disabled = !disabled; outer.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Size(44).Render(cx), kit.Button(sidebarLabel, func() { showSidebar = !showSidebar; outer.Visible(showSidebar, true) }).Size(44).Render(cx), el.Div().H(el.Dp(height)).Border(1, theme.Outline).Items(el.Stretch).Child(outer.Render(cx)))
		}))
	})
}
