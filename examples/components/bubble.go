package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("bubble", "data", func() core.Widget {
		say := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		incoming := kit.Bubble(say("可以，预计 10:00 前出库。Shipping by 10am."))
		outgoing := kit.Bubble(say("明天上午能发货吗？")).Mine()
		outline := kit.Bubble(say("确认发货时间后请点赞。")).Variant(kit.BubbleOutline).ReactionActions(kit.Button("赞同", func() {}).Variant(kit.ButtonGhost).Size(24), kit.Button("复制", func() {}).Variant(kit.ButtonGhost).Size(24))
		ghost := kit.Bubble(say("Ghost 使用整行宽度，适合长回复和自定义内容。")).Variant(kit.BubbleGhost)
		danger := kit.Bubble(say("发送失败，请检查网络后重试。")).Variant(kit.BubbleDestructive)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(10).W(el.Dp(480)).MaxW(el.Full).Child(
				outgoing.Render(cx), incoming.Render(cx), outline.Render(cx), ghost.Render(cx), danger.Render(cx))
		}))
	})
}
