package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"strconv"
)

func init() {
	registerSection("scrollable", "shell", func() core.Widget {
		mode := el.ScrollbarAlways
		names := []string{demoText("Always visible", "常显"), demoText("Show on hover", "悬停显示"), demoText("Show while scrolling", "滚动时显示")}
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Gap(12)
			for i := 0; i < 8; i++ {
				row.Child(el.Div().W(el.Dp(150)).H(el.Dp(100)).NoShrink().Rounded(8).Bg(theme.Subtle).Center().Child(el.Text(demoText("Card ", "卡片 ") + strconv.Itoa(i+1))))
			}
			return el.Div().P(24).Gap(16).Items(el.Stretch).Child(
				el.Text(demoText("Horizontal scrolling", "横向滚动")).TextSize(24).Bold(),
				el.Text(demoText("Use the trackpad, drag the bottom thumb, or click the track. After focusing with Tab, use arrow keys or Home / End.", "支持触控板、拖动底部滑块或点击轨道；Tab 聚焦后可用方向键、Home / End。")),
				el.Div().ID("cards").Focusable(true).ScrollX().Scrollbars(mode).H(el.Dp(120)).Child(row),
				kit.Button(demoText("Display policy: ", "显示策略：")+names[int(mode)], func() { mode = (mode + 1) % 3 }).Render(cx),
				kit.Button(demoText("View last image", "查看最后一张"), func() { cx.ScrollIntoViewX("cards", 7*162, 7*162+150) }).Render(cx))
		}))
	})
}
