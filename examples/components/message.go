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
		failed := kit.Message("我", say("请优先处理这笔订单")).User()
		failed.SetState(kit.MessageFailed, "网络中断")
		failed.OnRetry(func() { failed.SetState(kit.MessageReady, "") })
		reply := kit.Message("AI 助手", say(answer)).
			Avatar(kit.Avatar("AI").Size(32)).
			Header(say("AI 助手 · 刚刚")).
			Footer(say("订单数据已同步")).Actions(copy).
			Reactions(kit.MessageReaction{Name: "有帮助", Count: 2}).OnReaction(func(int, bool) {})
		notice := kit.Message("系统", nil).Avatar(nil).
			Bubble(kit.Bubble(say("这段对话已保存到订单记录。")).Variant(kit.BubbleGhost)).
			Header(kit.Label("系统消息")).Footer(kit.Label("仅当前团队可见"))
		positioned := kit.Message("我", say("用户消息也可以独立靠左排版。")).User().Alignment(el.Start).Header(kit.Label("自定义消息位置"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(520)).MaxW(el.Full).Child(
				kit.Message("我", say("SO-1021 什么时候到？")).User().Render(cx),
				reply.Render(cx), failed.Render(cx), notice.Render(cx), positioned.Render(cx))
		}))
	})
}
