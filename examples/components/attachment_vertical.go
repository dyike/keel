package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
)

func init() {
	registerSection("attachment_vertical", "data", func() core.Widget {
		pixels := image.NewNRGBA(image.Rect(0, 0, 400, 200))
		for y := 0; y < 200; y++ {
			for x := 0; x < 400; x++ {
				pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(30 + x/2), G: uint8(60 + y/2), B: 180, A: 255})
			}
		}
		square := kit.Attachment("默认方形.png", 64000).Size(kit.AttachmentSizeSmall).Vertical(true).Media(kit.Image(pixels, "方形预览")).OnRemove(func() {})
		wide := kit.Attachment("宽屏预览.png", 64000).Size(kit.AttachmentSizeSmall).Vertical(true).MediaAspectRatio(2).Media(kit.Image(pixels, "宽屏预览")).OnCancel(func() {})
		wide.SetProgress(.6)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(el.Text("方形和自定义宽高比；操作区在右上角"), el.Div().Row().Wrap().Gap(12).Items(el.Start).Child(square.Render(cx), wide.Render(cx)))
		}))
	})
}
