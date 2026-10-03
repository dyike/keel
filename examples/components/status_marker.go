package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("status_marker", "controls", func() core.Widget {
		loading := kit.StatusMarker("正在生成回答…").Loading(true).LoadingStyle(kit.StatusMarkerLoadingStyleShimmer).Role("status")
		enabled := true
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(520)).MaxW(el.Full).P(24).Gap(20).Items(el.Stretch).Child(
				kit.StatusMarker("已同步").Icon(kit.Icon(kit.IconCheck).Size(16)).Render(cx),
				kit.StatusMarker("今天").Variant(kit.StatusMarkerSeparator).Render(cx),
				kit.StatusMarker("3 条未读消息").Variant(kit.StatusMarkerBorder).Content(kit.Button("查看", func() { loading.SetText("已查看消息") })).Render(cx),
				kit.StatusMarker("正在载入…").Loading(true).Render(cx), loading.Render(cx),
				kit.StatusMarker("已送达").Alignment(el.End).Render(cx),
				kit.Button("切换加载状态", func() { enabled = !enabled; loading.Loading(enabled) }).Render(cx),
			)
		}))
	})
}
