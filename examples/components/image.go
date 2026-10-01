package main

import (
	"image"
	"image/color"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("image", "controls", func() core.Widget {
		img := image.NewNRGBA(image.Rect(0, 0, 240, 72))
		colors := []color.NRGBA{{R: 224, G: 96, B: 86, A: 255}, {R: 89, G: 171, B: 137, A: 255}, {R: 87, G: 139, B: 209, A: 255}}
		for y := 0; y < 72; y++ {
			for x := 0; x < 240; x++ {
				img.SetNRGBA(x, y, colors[x/80])
			}
		}
		msg := "点击图片查看反馈"
		photo := kit.Image(img, "红绿蓝色块").Rounded(8).OnClick(func() { msg = "已点击图片" })
		loading := kit.Image(nil, "加载中的图片")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(photo.Render(cx), el.Text(msg).TextColor(theme.Muted), loading.Render(cx))
		}))
	})
}
