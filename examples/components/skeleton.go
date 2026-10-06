package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("skeleton", "controls", func() core.Widget { return el.Embed(&skeletonGallery{}) })
}

type skeletonGallery struct{ loaded bool }

func (v *skeletonGallery) Render(cx *el.Context) el.Element {
	card := el.Div().P(20).MinH(el.Dp(200)).Gap(12).W(el.Dp(360)).MaxW(el.Full).Rounded(theme.RadiusMd).Bg(theme.Surface).Border(1, theme.Border)
	label := demoText("Simulate loading complete", "模拟加载完成")
	if v.loaded {
		label = demoText("Reload", "重新加载")
		card.Child(el.Text(demoText("Keel component gallery", "Keel 组件库")).Bold(), el.Text(demoText("Loaded content replaces the gray placeholders.", "加载完成后，真实内容替换灰色占位。")), el.Text(demoText("Reserve space for the avatar, title, and body.", "头像、标题和正文都可以预先保留空间。")).TextColor(theme.Muted))
	} else {
		card.Child(kit.Skeleton().W(el.Dp(48)).H(el.Dp(48)).Circle().Render(cx), kit.Skeleton().W(el.Dp(180)).Shimmer().Render(cx), kit.Skeleton().Shimmer().Render(cx), kit.Skeleton().W(el.Dp(240)).Secondary(true).Render(cx))
	}
	return el.Div().P(24).Gap(16).Child(
		el.Text(demoText("Skeleton · Loading placeholder", "Skeleton · 加载占位")).Bold(),
		el.Text(demoText("Gray blocks reserve space while content loads, reducing layout shifts when it appears.", "灰块表示内容正在加载，不是正文缺失。它们提前占位，减少内容出现时的布局跳动。")),
		card,
		kit.Button(label, func() { v.loaded = !v.loaded }).Render(cx),
		el.Text(demoText("Basic shapes: circle, rounded rectangle, rectangle", "基本形状：圆形、圆角矩形、直角矩形")).TextColor(theme.Muted),
		el.Div().Row().Gap(16).Child(kit.Skeleton().W(el.Dp(48)).H(el.Dp(48)).Circle().Render(cx), kit.Skeleton().W(el.Dp(120)).H(el.Dp(48)).Rounded(24).Render(cx), kit.Skeleton().W(el.Dp(120)).H(el.Dp(48)).Rounded(0).Render(cx)),
		kit.Button(demoText("Toggle reduced motion", "切换减少动画"), func() { theme.SetReducedMotion(!theme.ReducedMotion) }).Render(cx),
	)
}
