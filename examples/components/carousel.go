package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("carousel", "shell", func() core.Widget {
		slide := func(title, body string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().P(24).Gap(8).Child(el.Text(title).TextSize(20).Bold(), el.Text(body).TextColor(theme.Muted))
			})
		}
		car := kit.Carousel(
			slide("欢迎使用 Keel", "用 Go 写桌面界面，不需要 HTML。"),
			slide("组件 Components", "70 多个 kit 组件，支持深色和多语言。"),
			slide("Agent 测试", "keel-mcp 让 Agent 像用户一样操作界面。"),
		).Autoplay(4 * time.Second).Height(160)
		disabled := false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(el.Div().W(el.Dp(480)).Child(car.Render(cx)), kit.Button("启用 / 禁用轮播", func() { disabled = !disabled; car.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx))
		}))
	})
}
