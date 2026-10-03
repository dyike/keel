package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("attachment_group", "data", func() core.Widget {
		first := kit.Attachment("报告.pdf", 2048)
		second := kit.Attachment("图片.png", 4096)
		group := kit.AttachmentGroup(first, second).Name("附件").Gap(12)
		second.OnRemove(func() { group.SetItems(first) })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Child(el.Text("横向滚动查看；移除按钮更新附件组"), group.Render(cx))
		}))
	})
}
