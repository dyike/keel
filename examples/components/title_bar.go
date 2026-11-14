package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func framelessOptions() window.Options {
	return window.Options{
		Title: demoText("Keel Frameless Window", "Keel 无边框窗口"),
		Width: 820, Height: 520, Frameless: true, Content: el.Root(newFramelessDemo()),
	}
}

func init() {
	registerSection("title_bar", "shell", func() core.Widget {
		demo := newFramelessDemo()
		var preview *window.Window
		open := func() {
			if preview != nil && !preview.Closed() {
				preview.Raise()
				return
			}
			options := framelessOptions()
			options.OnClose = func() { preview = nil }
			preview = window.Open(options)
		}
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Items(el.Stretch).HFull().Child(
				el.Div().P(16).Gap(12).NoShrink().Child(
					el.Text(demoText("Frameless windows use an app-drawn title bar. Open the preview to try dragging, window controls, search and commands.", "无边框窗口由应用绘制标题栏。打开预览，体验拖动窗口、窗口按钮、搜索框和命令面板。")).TextColor(theme.Muted),
					kit.Button(demoText("Open frameless window", "打开无边框窗口"), open).Render(cx),
				),
				el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(demo.Render(cx)),
			)
		}))
	})
}
