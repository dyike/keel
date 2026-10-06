package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("theme", "controls", func() core.Widget { return el.Embed(&themeGallery{}) })
}

type themeGallery struct{}

func (*themeGallery) Render(cx *el.Context) el.Element {
	return el.Div().Bg(theme.Surface).P(16).Gap(12).Child(
		el.Text(demoText("Global theme", "全局主题")).TextSize(float32(theme.HeadingSize)),
		el.Div().Row().Gap(12).Child(
			el.Div().P(10).Bg(theme.Subtle).OnClick(func() { theme.Apply(theme.Light()) }).Child(el.Text(demoText("Light", "浅色"))),
			el.Div().P(10).Bg(theme.Subtle).OnClick(func() { theme.Apply(theme.Dark()) }).Child(el.Text(demoText("Dark", "深色"))),
		),
		cx.Cache("semantic-colors", func() el.Element {
			return el.Div().Gap(8).Child(
				el.Text(demoText("Success: saved", "成功：保存完成")).TextColor(theme.Success),
				el.Text(demoText("Warning: check your input", "警告：请检查输入")).TextColor(theme.Warning),
				el.Text(demoText("Info: new messages", "提示：有新消息")).TextColor(theme.Info),
			)
		}),
		el.Text(demoText("Local theme: this region uses nord; the rest of the page is unchanged", "局部主题：下面这块用 nord，页面其余部分不变")).TextColor(theme.Muted),
		el.Div().Row().Gap(12).Child(themedCard(cx, "nord"), themedCard(cx, "high-contrast")),
	)
}

// themedCard shows kit components in a named theme, inside this page.
func themedCard(cx *el.Context, name string) el.Element {
	p, _ := theme.Named(name)
	return cx.Themed(p, el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(260)).P(16).Gap(10).Rounded(theme.RadiusLg).Bg(theme.Surface).Border(1, theme.Border).Child(
			el.Text(name).Bold(),
			kit.Input("").Placeholder(demoText("Input field", "输入框")).Render(cx),
			el.Div().Row().Gap(8).Child(kit.Button(demoText("Primary", "主要"), nil).Render(cx), kit.Button(demoText("Secondary", "次要"), nil).Variant(kit.ButtonSecondary).Render(cx)),
		)
	}))
}
