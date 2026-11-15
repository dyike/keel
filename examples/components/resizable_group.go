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
			return el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Grow().ScrollY().Bg(theme.Surface).P(12).Child(el.Text(label))
			})
		}
		g := kit.ResizableGroup(kit.ResizablePanel{ID: "files", Content: pane(demoText("Files", "文件")), Size: 120, Min: 60, Max: 180}, kit.ResizablePanel{ID: "editor", Content: kit.Input(demoText("Edit content", "编辑内容")), Min: 100}, kit.ResizablePanel{ID: "preview", Content: pane(demoText("Preview", "预览")), Size: 120, Min: 60})
		visible := true
		compact := false
		sizes := [2]map[string]float32{g.Sizes(), {"files": 100, "editor": 160, "preview": 120}}
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			width, _ := cx.ViewportSize()
			narrow := width < 640*max(float32(1), cx.TextScale())
			if narrow != compact {
				old, next := 0, 1
				if compact {
					old, next = 1, 0
				}
				sizes[old] = g.Sizes()
				g.SetVertical(narrow).SetSizes(sizes[next])
				compact = narrow
			}
			height := float32(320)
			if compact {
				height = 480 * max(float32(1), cx.TextScale())
			}
			label := demoText("Hide preview", "隐藏预览")
			if !visible {
				label = demoText("Show preview", "显示预览")
			}
			return el.Div().ScrollY().P(24).Gap(12).Items(el.Stretch).Child(kit.Button(label, func() { visible = !visible; g.SetVisible("preview", visible) }).Size(44).Render(cx), el.Div().H(el.Dp(height)).Items(el.Stretch).Child(g.Render(cx)))
		}))
	})
}
