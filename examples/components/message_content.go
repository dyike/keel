package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("message_content", "data", func() core.Widget {
		mixed := kit.MessageContent(
			kit.Bubble(kit.Label("导出已完成，文件在下方。")),
			kit.Attachment("订单汇总.csv", 2048),
			kit.Bubble(kit.Label("下载链接将在 24 小时后过期。")).Variant(kit.BubbleGhost),
		)
		msg := kit.Message("助手", mixed).Header(kit.Label("助手 · 刚刚")).Footer(kit.Label("已送达"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(msg.Render(cx))
		}))
	})
}
