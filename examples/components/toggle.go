package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("toggle", "controls", func() core.Widget {
		pin, star := kit.Toggle(demoText("Pin", "固定 Pin"), false).Variant(kit.ToggleGhost).Size(kit.ToggleSizeSmall), kit.Toggle(demoText("Star", "收藏"), true).Icon(kit.IconStar).Variant(kit.ToggleOutline).Size(kit.ToggleSizeLarge)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Row().Gap(8).Child(pin.Render(cx), star.Render(cx))
		}))
	})
}
