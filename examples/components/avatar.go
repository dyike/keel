package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
)

func init() {
	registerSection("avatar", "controls", func() core.Widget {
		img := image.NewNRGBA(image.Rect(0, 0, 80, 40))
		for y := 0; y < 40; y++ {
			for x := 0; x < 80; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 3), G: 140, B: uint8(y * 6), A: 255})
			}
		}
		return el.Embed(&avatarGallery{photo: kit.Avatar("示例图片").Image(img)})
	})
}

type avatarGallery struct{ photo *kit.AvatarView }

func (v *avatarGallery) Render(cx *el.Context) el.Element {
	return el.Div().Row().Gap(16).Items(el.Center).Child(kit.Avatar("张三").Size(32).Render(cx), kit.Avatar("Ada Lovelace").Render(cx), kit.Avatar("").Size(56).Render(cx), v.photo.Render(cx))
}
