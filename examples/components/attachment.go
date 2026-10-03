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
		up := kit.Attachment("季度报表 Q3.xlsx", 2_400_000)
		up.SetProgress(0.6)
		up.OnCancel(func() {}).OnRetry(func() {})
		bad := kit.Attachment("合同扫描.pdf", 18_000_000)
		bad.SetError("网络中断，请重试")
		bad.OnRetry(func() {}).OnCancel(func() {})
		preview := image.NewNRGBA(image.Rect(0, 0, 240, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 240; x++ {
				preview.SetNRGBA(x, y, color.NRGBA{R: uint8(40 + x/2), G: uint8(80 + y), B: 190, A: 255})
			}
		}
		done := kit.Attachment("logo.png", 48_000).Vertical(true).Media(kit.Image(preview, "图片预览").Size(240, 100).Fit(kit.ImageCover).Rounded(8))
		pending := kit.Attachment("待上传.txt", 512)
		pending.SetStatus(kit.AttachmentStatusPending)
		rejected := kit.Attachment("不支持的文件.exe", 2048)
		rejected.SetError("不支持此文件类型")
		mediaUpload := kit.Attachment("风景上传.png", 48000).Media(kit.Image(preview, "上传预览").Size(64, 64))
		mediaUpload.SetProgress(.6)
		mediaProcessing := kit.Attachment("照片处理.png", 48000).Media(kit.Image(preview, "处理预览").Size(64, 64))
		mediaProcessing.SetStatus(kit.AttachmentStatusProcessing)
		mediaFailed := kit.Attachment("照片失败.png", 48000).Media(kit.Image(preview, "失败预览").Size(64, 64)).OnRetry(func() {})
		mediaFailed.SetError("网络中断")
		removed := false
		done.OnRemove(func() { removed = true })
		done.Actions(kit.Button("查看版本", func() {}).Variant(kit.ButtonGhost)).PartStyle(kit.AttachmentPartRoot, func(e *el.DivEl) { e.P(12) })
		group := kit.AttachmentGroup(up, bad, done).Name("附件").Gap(12)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			if removed {
				group.SetItems(up, bad)
			}
			return el.Div().P(24).Gap(10).Child(el.Text("横向滚动查看附件；移除后保留其他上传状态"), group.Render(cx), pending.Render(cx), rejected.Render(cx), mediaUpload.Render(cx), mediaProcessing.Render(cx), mediaFailed.Render(cx))
		}))
	})
}
