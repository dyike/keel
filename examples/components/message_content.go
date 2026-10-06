package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("message_content", "data", func() core.Widget {
		mixed := kit.MessageContent(
			kit.Bubble(kit.Label(demoText("Export complete. The file is below.", "导出已完成，文件在下方。"))),
			kit.Attachment(demoText("order-summary.csv", "订单汇总.csv"), 2048),
			kit.Bubble(kit.Label(demoText("The download link expires in 24 hours.", "下载链接将在 24 小时后过期。"))).Variant(kit.BubbleGhost),
		)
		msg := kit.Message(demoText("Assistant", "助手"), mixed).Header(kit.Label(demoText("Assistant · Just now", "助手 · 刚刚"))).Footer(kit.Label(demoText("Delivered", "已送达")))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(msg.Render(cx))
		}))
	})
}
