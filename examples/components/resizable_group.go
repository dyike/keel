package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("resizable_group", "shell", func() core.Widget {
		pane := func(label string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element { return el.Div().Grow().Bg(theme.Surface).P(12).Child(el.Text(label)) })
		}
		g := kit.ResizableGroup(kit.ResizablePanel{ID: "files", Content: pane(demoText("Files", "文件")), Size: 120, Min: 60, Max: 180}, kit.ResizablePanel{ID: "editor", Content: kit.Input(demoText("Edit content", "编辑内容")), Min: 100}, kit.ResizablePanel{ID: "preview", Content: pane(demoText("Preview", "预览")), Size: 120, Min: 60})
		visible := true
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(kit.Button(demoText("Show / hide preview", "显示 / 隐藏预览"), func() { visible = !visible; g.SetVisible("preview", visible) }).Render(cx), el.Div().H(el.Dp(320)).Items(el.Stretch).Child(g.Render(cx)))
		}))
	})
}
