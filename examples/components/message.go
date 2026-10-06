package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("message", "data", func() core.Widget {
		say := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		answer := demoText("Order SO-1021 is paid and should arrive within 3 days. Should I ask the warehouse to prioritize it?", "订单 SO-1021 已付款，预计 3 天内送达。需要我通知仓库优先处理吗？")
		copy := kit.CopyButton(func() string { return answer })
		failed := kit.Message(demoText("Me", "我"), say(demoText("Please prioritize this order", "请优先处理这笔订单"))).User()
		failed.SetState(kit.MessageFailed, demoText("Connection lost", "网络中断"))
		failed.OnRetry(func() { failed.SetState(kit.MessageReady, "") })
		reply := kit.Message(demoText("AI assistant", "AI 助手"), say(answer)).
			Avatar(kit.Avatar("AI").Size(32)).
			Header(say(demoText("AI assistant · Just now", "AI 助手 · 刚刚"))).
			Footer(say(demoText("Order data synced", "订单数据已同步"))).Actions(copy).
			Reactions(kit.MessageReaction{Name: demoText("Helpful", "有帮助"), Count: 2}).OnReaction(func(int, bool) {})
		notice := kit.Message(demoText("System", "系统"), nil).Avatar(nil).
			Bubble(kit.Bubble(say(demoText("This conversation has been saved to the order record.", "这段对话已保存到订单记录。"))).Variant(kit.BubbleGhost)).
			Header(kit.Label(demoText("System message", "系统消息"))).Footer(kit.Label(demoText("Visible to this team only", "仅当前团队可见"))).
			PartStyle(kit.MessagePartRoot, func(e *el.DivEl) { e.P(theme.SpaceLg).Bg(theme.Subtle).Rounded(theme.RadiusMd) }).
			PartStyle(kit.MessagePartStack, func(e *el.DivEl) { e.Gap(theme.SpaceSm) })
		positioned := kit.Message(demoText("Me", "我"), say(demoText("User messages can also be independently left-aligned.", "用户消息也可以独立靠左排版。"))).User().Alignment(el.Start).Header(kit.Label(demoText("Custom message placement", "自定义消息位置")))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(520)).MaxW(el.Full).Child(
				kit.Message(demoText("Me", "我"), say(demoText("When will SO-1021 arrive?", "SO-1021 什么时候到？"))).User().Render(cx),
				reply.Render(cx), failed.Render(cx), notice.Render(cx), positioned.Render(cx))
		}))
	})
}
