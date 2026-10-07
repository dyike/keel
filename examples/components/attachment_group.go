package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("attachment_group", "data", func() core.Widget {
		first := kit.Attachment(demoText("report.pdf", "报告.pdf"), 2048)
		second := kit.Attachment(demoText("image.png", "图片.png"), 4096)
		pending := kit.Attachment(demoText("pending-upload.zip", "待上传.zip"), 8192)
		pending.SetStatus(kit.AttachmentStatusPending)
		processing := kit.Attachment(demoText("processing.mp4", "处理中.mp4"), 16000)
		processing.SetStatus(kit.AttachmentStatusProcessing)
		processing.OnCancel(func() {}).OnRetry(func() {})
		group := kit.AttachmentGroup(first, second, pending, processing).Name(demoText("Attachments", "附件")).Gap(12)
		second.OnRemove(func() { group.SetItems(first, pending, processing) })
		sizes := []*kit.AttachmentView{
			kit.Attachment("XSmall.pdf", 1024).Size(kit.AttachmentSizeXSmall),
			kit.Attachment("Small.pdf", 1024).Size(kit.AttachmentSizeSmall),
			kit.Attachment("Medium.pdf", 1024).Size(kit.AttachmentSizeMedium),
			kit.Attachment("Large.pdf", 1024).Size(kit.AttachmentSizeLarge),
		}
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			group.EdgeFade(theme.Bg)
			paging := el.Div().Wrap().Gap(8).Child(
				kit.Button(demoText("Previous screen", "前一屏"), func() { offset, width, _ := group.ScrollState(cx); group.ScrollTo(offset - width) }).Render(cx),
				kit.Button(demoText("Next screen", "后一屏"), func() { offset, width, _ := group.ScrollState(cx); group.ScrollTo(offset + width) }).Render(cx),
			)
			box := el.Div().P(24).Gap(12).Child(el.Text(demoText("Scroll horizontally; remove buttons update the attachment group", "横向滚动查看；移除按钮更新附件组")), paging, group.Render(cx))
			for _, a := range sizes {
				box.Child(a.Render(cx))
			}
			return box
		}))
	})
}
