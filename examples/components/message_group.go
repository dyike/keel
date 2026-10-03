package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("message_group", "data", func() core.Widget {
		say := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		received := kit.MessageGroup(
			kit.Message("客服", say("订单已确认。")).Header(kit.Label("客服 · 10:24")),
			kit.Message("客服", say("明天上午安排发货，运单号会在出库后发送。")).Footer(kit.Label("已送达")),
		).Name("客服连续消息")
		sent := kit.MessageGroup(
			kit.Message("我", say("收到，谢谢。 ")).User(),
			kit.Message("我", say("请帮我备注放在门卫。 ")).User(),
		).Name("我的连续消息")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(24).Child(received.Render(cx), sent.Render(cx))
		}))
	})
}
