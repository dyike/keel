package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("theme", "controls", func() core.Widget { return el.Embed(&themeGallery{}) })
}

type themeGallery struct{}

func (*themeGallery) Render(cx *el.Context) el.Element {
	return el.Div().Bg(theme.Surface).P(16).Gap(12).Child(
		el.Text("全局主题").TextSize(float32(theme.HeadingSize)),
		el.Div().Row().Gap(12).Child(
			el.Div().P(10).Bg(theme.Subtle).OnClick(func() { theme.Apply(theme.Light()) }).Child(el.Text("浅色")),
			el.Div().P(10).Bg(theme.Subtle).OnClick(func() { theme.Apply(theme.Dark()) }).Child(el.Text("深色")),
		),
		cx.Cache("semantic-colors", func() el.Element {
			return el.Div().Gap(8).Child(
				el.Text("成功：保存完成").TextColor(theme.Success),
				el.Text("警告：请检查输入").TextColor(theme.Warning),
				el.Text("提示：有新消息").TextColor(theme.Info),
			)
		}),
	)
}
