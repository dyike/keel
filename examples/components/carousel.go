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
			slide(demoText("Welcome to Keel", "欢迎使用 Keel"), demoText("Build desktop UIs in Go without HTML.", "用 Go 写桌面界面，不需要 HTML。")),
			slide(demoText("Components", "组件 Components"), demoText("80+ kit components with dark mode and localization.", "80 多个 kit 组件，支持深色和多语言。")),
			slide(demoText("Agent testing", "Agent 测试"), demoText("keel-mcp lets agents interact with the UI like users.", "keel-mcp 让 Agent 像用户一样操作界面。")),
		).Autoplay(4*time.Second).Height(180).Loop(false).Basis(2.0/3).ItemBasis(1, 0.5).ItemSize(0, 260).ItemSize(2, 640)
		content := car.Content()
		previous := car.PreviousControl(kit.Button(demoText("Previous item", "上一项"), nil).Variant(kit.ButtonSecondary))
		next := car.NextControl(kit.Button(demoText("Next item", "下一项"), nil).Variant(kit.ButtonSecondary))
		pages := []el.View{car.PaginationItem(0, nil), car.PaginationItem(1, nil), car.PaginationItem(2, nil)}
		disabled := false
		vertical := false
		wheelStep := false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(el.Div().W(el.Dp(480)).MaxW(el.Full).Gap(theme.SpaceMd).Child(content.Render(cx), el.Div().Row().Wrap().Items(el.Center).Gap(theme.SpaceSm).Child(previous.Render(cx), pages[0].Render(cx), pages[1].Render(cx), pages[2].Render(cx), next.Render(cx))), kit.Button(demoText("Toggle horizontal / vertical", "切换横向 / 竖向"), func() { vertical = !vertical; car.Vertical(vertical) }).Variant(kit.ButtonSecondary).Render(cx), kit.Button(demoText("Toggle continuous / item scrolling", "切换连续 / 逐项滚动"), func() { wheelStep = !wheelStep; car.WheelStep(wheelStep) }).Variant(kit.ButtonSecondary).Render(cx), kit.Button(demoText("Enable / disable carousel", "启用 / 禁用轮播"), func() { disabled = !disabled; car.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx))
		}))
	})
}
