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
		removed := false
		done.OnRemove(func() { removed = true })
		group := kit.AttachmentGroup(up, bad, done).Name("附件").Gap(12)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			if removed {
				group.SetItems(up, bad)
			}
			return el.Div().P(24).Gap(10).Child(el.Text("横向滚动查看附件；移除后保留其他上传状态"), group.Render(cx))
		}))
	})
}
