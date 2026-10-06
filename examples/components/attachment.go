package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
)

func init() {
	registerSection("attachment", "data", func() core.Widget {
		up := kit.Attachment(demoText("quarterly-report-Q3.xlsx", "季度报表 Q3.xlsx"), 2_400_000)
		up.SetProgress(0.6)
		up.OnCancel(func() {}).OnRetry(func() {})
		bad := kit.Attachment(demoText("scanned-contract.pdf", "合同扫描.pdf"), 18_000_000)
		bad.SetError(demoText("Connection lost; try again", "网络中断，请重试"))
		bad.OnRetry(func() {}).OnCancel(func() {})
		preview := image.NewNRGBA(image.Rect(0, 0, 240, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 240; x++ {
				preview.SetNRGBA(x, y, color.NRGBA{R: uint8(40 + x/2), G: uint8(80 + y), B: 190, A: 255})
			}
		}
		done := kit.Attachment("logo.png", 48_000).Vertical(true).Media(kit.Image(preview, demoText("Image preview", "图片预览")).Size(240, 100).Fit(kit.ImageCover).Rounded(8))
		pending := kit.Attachment(demoText("pending-upload.txt", "待上传.txt"), 512)
		pending.SetStatus(kit.AttachmentStatusPending)
		rejected := kit.Attachment(demoText("unsupported-file.exe", "不支持的文件.exe"), 2048)
		rejected.SetError(demoText("Unsupported file type", "不支持此文件类型"))
		mediaUpload := kit.Attachment(demoText("landscape-upload.png", "风景上传.png"), 48000).Media(kit.Image(preview, demoText("Upload preview", "上传预览")).Size(64, 64))
		mediaUpload.SetProgress(.6)
		mediaProcessing := kit.Attachment(demoText("processing-photo.png", "照片处理.png"), 48000).Media(kit.Image(preview, demoText("Processing preview", "处理预览")).Size(64, 64))
		mediaProcessing.SetStatus(kit.AttachmentStatusProcessing)
		mediaFailed := kit.Attachment(demoText("failed-photo.png", "照片失败.png"), 48000).Media(kit.Image(preview, demoText("Failed preview", "失败预览")).Size(64, 64)).OnRetry(func() {})
		mediaFailed.SetError(demoText("Connection lost", "网络中断"))
		custom := kit.Attachment(demoText("video-preview.mp4", "视频预览.mp4"), 96000).Media(kit.Image(preview, demoText("Video thumbnail", "视频缩略图")).Size(80, 80)).
			MediaOverlay(kit.Button(demoText("Play", "播放"), func() {}).Size(24)).OnOpen(func() {})
		previous := kit.Attachment(demoText("previous-version.pdf", "保留历史版本.pdf"), 1024).Description(demoText("Previous version uploaded", "上一版本已上传")).
			PartStatus(kit.AttachmentPartDescription, kit.AttachmentStatusComplete).OnRetry(func() {})
		previous.SetError(demoText("Current version failed to upload", "当前版本上传失败"))
		removed := false
		done.OnRemove(func() { removed = true })
		done.Actions(kit.Button(demoText("View versions", "查看版本"), func() {}).Variant(kit.ButtonGhost)).PartStyle(kit.AttachmentPartRoot, func(e *el.DivEl) { e.P(12) })
		group := kit.AttachmentGroup(up, bad, done).Name(demoText("Attachments", "附件")).Gap(12)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			if removed {
				group.SetItems(up, bad)
			}
			return el.Div().P(24).Gap(10).Child(el.Text(demoText("Scroll horizontally for attachments; removing one preserves the others' upload states", "横向滚动查看附件；移除后保留其他上传状态")), group.Render(cx), pending.Render(cx), rejected.Render(cx), mediaUpload.Render(cx), mediaProcessing.Render(cx), mediaFailed.Render(cx), custom.Render(cx), previous.Render(cx))
		}))
	})
}
