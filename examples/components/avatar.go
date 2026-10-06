package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("avatar", "controls", func() core.Widget {
		img := image.NewNRGBA(image.Rect(0, 0, 80, 40))
		for y := 0; y < 40; y++ {
			for x := 0; x < 80; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 3), G: 140, B: uint8(y * 6), A: 255})
			}
		}
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, img); err != nil {
			panic(err)
		}
		return el.Embed(&avatarGallery{
			photo:  kit.Avatar(demoText("Sample image", "示例图片")).Image(img),
			source: kit.Avatar(demoText("Async image", "异步图片")).Source("data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())),
		})
	})
}

type avatarGallery struct{ photo, source *kit.AvatarView }

func (v *avatarGallery) Render(cx *el.Context) el.Element {
	return el.Div().Row().Gap(16).Items(el.Center).Child(kit.Avatar(demoText("Alex Chen", "张三")).Status(kit.AvatarOnline).Size(32).Render(cx), kit.Avatar("Ada Lovelace").Status(kit.AvatarBusy).Render(cx), kit.Avatar("").Size(56).Render(cx), v.photo.Render(cx), v.source.Render(cx),
		kit.Avatar("Keel Team").Size(48).Rounded(theme.RadiusLg).Colors(theme.Primary, theme.OnColor).Render(cx),
		kit.Avatar("Bob").Size(48).Border(2, theme.Primary).Render(cx),
		kit.Avatar("").Size(48).Placeholder(kit.IconSettings).Rounded(theme.RadiusMd).Render(cx))
}
