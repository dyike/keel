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
			kit.Message(demoText("Support", "客服"), say(demoText("Order confirmed.", "订单已确认。"))).Header(kit.Label(demoText("Support · 10:24", "客服 · 10:24"))),
			kit.Message(demoText("Support", "客服"), say(demoText("Shipping is scheduled for tomorrow morning. The tracking number will be sent after dispatch.", "明天上午安排发货，运单号会在出库后发送。"))).Footer(kit.Label(demoText("Delivered", "已送达"))),
		).Name(demoText("Consecutive support messages", "客服连续消息"))
		sent := kit.MessageGroup(
			kit.Message(demoText("Me", "我"), say(demoText("Got it, thanks. ", "收到，谢谢。 "))).User(),
			kit.Message(demoText("Me", "我"), say(demoText("Please leave it at the reception desk. ", "请帮我备注放在门卫。 "))).User(),
		).Name(demoText("My consecutive messages", "我的连续消息"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(24).Child(received.Render(cx), sent.Render(cx))
		}))
	})
}
