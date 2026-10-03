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
		pending := kit.Attachment("待上传.zip", 8192)
		pending.SetStatus(kit.AttachmentStatusPending)
		processing := kit.Attachment("处理中.mp4", 16000)
		processing.SetStatus(kit.AttachmentStatusProcessing)
		processing.OnCancel(func() {}).OnRetry(func() {})
		group := kit.AttachmentGroup(first, second, pending, processing).Name("附件").Gap(12)
		second.OnRemove(func() { group.SetItems(first, pending, processing) })
		sizes := []*kit.AttachmentView{
			kit.Attachment("XSmall.pdf", 1024).Size(kit.AttachmentSizeXSmall),
			kit.Attachment("Small.pdf", 1024).Size(kit.AttachmentSizeSmall),
			kit.Attachment("Medium.pdf", 1024).Size(kit.AttachmentSizeMedium),
			kit.Attachment("Large.pdf", 1024).Size(kit.AttachmentSizeLarge),
		}
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			paging := el.Div().Row().Gap(8).Child(
				kit.Button("前一屏", func() { offset, width, _ := group.ScrollState(cx); group.ScrollTo(offset - width) }).Render(cx),
				kit.Button("后一屏", func() { offset, width, _ := group.ScrollState(cx); group.ScrollTo(offset + width) }).Render(cx),
			)
			box := el.Div().P(24).Gap(12).Child(el.Text("横向滚动查看；移除按钮更新附件组"), paging, group.Render(cx))
			for _, a := range sizes {
				box.Child(a.Render(cx))
			}
			return box
		}))
	})
}
