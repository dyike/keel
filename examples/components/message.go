package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("message", "data", func() core.Widget {
		say := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		answer := "订单 SO-1021 已付款，预计 3 天内送达。需要我通知仓库优先处理吗？"
		copy := kit.CopyButton(func() string { return answer })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(520)).Child(
				kit.Message("我", say("SO-1021 什么时候到？")).User().Render(cx),
				kit.Message("AI 助手", say(answer)).Actions(copy).Render(cx))
		}))
	})
}
