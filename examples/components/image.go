package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"net/url"

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
		msg := demoText("Click the image to preview; Esc closes it", "点击图片预览；Esc 关闭")
		photo := kit.Image(img, demoText("Red, green, and blue blocks", "红绿蓝色块")).Rounded(8).Preview().OnClick(func() { msg = demoText("Image clicked", "已点击图片") })
		loading := kit.Image(nil, demoText("Image that fails to load", "加载失败的图片")).Width(240)
		loading.SetError(demoText("Connection lost", "网络中断"))
		loading.OnRetry(func() { loading.SetImage(img) })
		var encoded bytes.Buffer
		_ = png.Encode(&encoded, img)
		source := "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
		sourced := kit.Image(nil, demoText("Async loading example", "异步加载示例")).Size(240, 72).LoadingContent(kit.Label(demoText("Loading…", "正在加载…"))).Fallback(kit.Label(demoText("Cannot read image", "无法读取图片"))).Source(source)
		crop := kit.Image(img, demoText("Center crop", "居中裁剪")).Size(120, 100).Fit(kit.ImageCover).Rounded(16).Preview()
		contain := kit.Image(img, demoText("Show in full", "完整显示")).Size(120, 100).Fit(kit.ImageContain).Rounded(16)
		vector := kit.Image(nil, demoText("SVG vector image", "SVG 矢量图")).Size(120, 100).Source("data:image/svg+xml," + url.PathEscape(demoSVG))
		animated := kit.Image(nil, demoText("Animated GIF", "GIF 动图")).Size(120, 100).Source("data:image/gif;base64," + base64.StdEncoding.EncodeToString(demoGIF()))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(photo.Render(cx), el.Text(msg).TextColor(theme.Muted), el.Div().Row().Gap(12).Child(crop.Render(cx), contain.Render(cx)), el.Div().Row().Gap(12).Child(vector.Render(cx), animated.Render(cx)), loading.Render(cx), sourced.Render(cx), kit.Button(demoText("Reload source image", "重新加载源图片"), sourced.Retry).Render(cx))
		}))
	})
}

const demoSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 100">
<circle cx="60" cy="50" r="40" fill="#578bd1"/><path d="M40 50 l14 14 l28 -28" stroke="#fff" stroke-width="8" fill="none" stroke-linecap="round"/></svg>`

// demoGIF builds a small looping animation: a dot moving along a track.
func demoGIF() []byte {
	pal := color.Palette{color.NRGBA{235, 238, 243, 255}, color.NRGBA{87, 139, 209, 255}}
	g := &gif.GIF{}
	for i := 0; i < 8; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, 120, 100), pal)
		cx := 20 + i*80/7
		for y := 40; y < 60; y++ {
			for x := cx - 10; x < cx+10; x++ {
				if (x-cx)*(x-cx)+(y-50)*(y-50) < 100 {
					frame.SetColorIndex(x, y, 1)
				}
			}
		}
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, 12)
	}
	var b bytes.Buffer
	_ = gif.EncodeAll(&b, g)
	return b.Bytes()
}
