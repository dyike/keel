package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
)

func init() {
	registerSection("badge", "controls", func() core.Widget {
		unread := kit.Badge(9)
		unread.Child(kit.Button(demoText("Notifications", "通知"), func() { unread.SetValue(0) }).Variant(kit.ButtonSecondary))
		online := kit.Badge(1).Dot().Tone(kit.ToneSuccess).Child(kit.Avatar(demoText("Alex Chen", "张三")))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Wrap().Gap(24).Items(el.Center).Child(
				unread.Render(cx),
				kit.Button(demoText("Add unread message", "增加未读"), func() { unread.SetValue(unread.Value() + 1) }).Variant(kit.ButtonGhost).Render(cx),
				kit.Badge(120).Render(cx), online.Render(cx),
				kit.Badge(0).Icon(kit.IconCheck).Name(demoText("Verified", "已验证")).Tone(kit.ToneSuccess).Child(kit.Avatar("Ada")).Render(cx),
				kit.Badge(0).Icon(kit.IconStar).Name(demoText("Star", "收藏")).Size(24).Color(color.NRGBA{R: 255, G: 210, A: 255}).Child(kit.Avatar(demoText("Lin", "林"))).Render(cx),
				kit.Badge(3).Size(12).Render(cx), kit.Badge(12).Size(18).Render(cx), kit.Badge(999).Size(24).Max(999).Render(cx))
		}))
	})
}
