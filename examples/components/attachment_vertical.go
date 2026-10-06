package main

import (
	"bytes"
	"encoding/base64"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
	"image/png"
	"time"
)

func init() {
	registerSection("attachment_vertical", "data", func() core.Widget {
		pixels := image.NewNRGBA(image.Rect(0, 0, 400, 200))
		for y := 0; y < 200; y++ {
			for x := 0; x < 400; x++ {
				pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(30 + x/2), G: uint8(60 + y/2), B: 180, A: 255})
			}
		}
		square := kit.Attachment(demoText("default-square.png", "默认方形.png"), 64000).Size(kit.AttachmentSizeSmall).Vertical(true).Media(kit.Image(pixels, demoText("Square preview", "方形预览"))).OnRemove(func() {})
		wide := kit.Attachment(demoText("wide-preview.png", "宽屏预览.png"), 64000).Size(kit.AttachmentSizeSmall).Vertical(true).MediaAspectRatio(2).Media(kit.Image(pixels, demoText("Wide preview", "宽屏预览"))).OnCancel(func() {})
		wide.TitleShimmer(kit.ShimmerStyle{Duration: 3 * time.Second, Spread: .45, Reverse: true})
		wide.SetProgress(.6)
		tile := kit.Attachment(demoText("Image only", "纯图片"), 64000).RemoveOnHover(false).Size(kit.AttachmentSizeSmall).Vertical(true).ShowContent(false).Media(kit.Image(pixels, demoText("Image-only preview", "纯图片预览"))).OnOpen(func() {}).OnRemove(func() {})
		metadata := kit.Attachment(demoText("metadata-only.txt", "只有元信息.txt"), 1024).Size(kit.AttachmentSizeSmall).ShowMedia(false)
		actions := kit.Attachment(demoText("Actions only", "只有操作"), 0).Size(kit.AttachmentSizeSmall).ShowMedia(false).ShowContent(false).Actions(kit.Button(demoText("Choose file", "选择文件"), func() {}).Size(24))
		var encoded bytes.Buffer
		_ = png.Encode(&encoded, pixels)
		source := kit.Attachment(demoText("Image URL", "URL 图片"), 64000).Size(kit.AttachmentSizeSmall).Vertical(true).ShowContent(false).
			MediaSource("data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes()))
		failed := kit.Attachment(demoText("Image failed to load", "图片加载失败"), 0).Size(kit.AttachmentSizeSmall).MediaSource("data:image/png;base64,invalid")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(el.Text(demoText("Square and wide previews; images have a separate remove button in the top right", "方形和宽屏预览；纯图片右上角为独立移除角标")), el.Div().Row().Wrap().Gap(12).Items(el.Start).Child(square.Render(cx), wide.Render(cx), tile.Render(cx), metadata.Render(cx), actions.Render(cx), source.Render(cx), failed.Render(cx)))
		}))
	})
}
