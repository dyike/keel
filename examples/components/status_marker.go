package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("status_marker", "controls", func() core.Widget {
		loading := kit.StatusMarker(demoText("Generating answer…", "正在生成回答…")).Loading(true).LoadingStyle(kit.StatusMarkerLoadingStyleShimmer).Role("status")
		synced := kit.StatusMarker(demoText("Synced", "已同步")).Icon(kit.Icon(kit.IconCheck).Size(16))
		today := kit.StatusMarker(demoText("Today", "今天")).Variant(kit.StatusMarkerSeparator)
		unread := kit.StatusMarker(demoText("3 unread messages", "3 条未读消息")).Variant(kit.StatusMarkerBorder).Content(kit.Button(demoText("View", "查看"), func() { loading.SetText(demoText("Message viewed", "已查看消息")) }))
		spinner := kit.StatusMarker(demoText("Loading…", "正在载入…")).Loading(true)
		delivered := kit.StatusMarker(demoText("Delivered", "已送达")).Alignment(el.End)
		enabled := true
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(520)).MaxW(el.Full).P(24).Gap(20).Items(el.Stretch).Child(
				synced.Render(cx), today.Render(cx), unread.Render(cx),
				spinner.Render(cx), loading.Render(cx), delivered.Render(cx),
				kit.Button(demoText("Toggle loading state", "切换加载状态"), func() { enabled = !enabled; loading.Loading(enabled) }).Render(cx),
			)
		}))
	})
}
