package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("toggle_group", "controls", func() core.Widget {
		align := kit.ToggleGroup("左对齐", "居中", "右对齐").Variant(kit.ToggleOutline).Size(kit.ToggleSizeLarge)
		align.SetValue("居中")
		style := kit.ToggleGroup("B", "I", "U").Multiple().Variant(kit.ToggleGhost).Size(kit.ToggleSizeSmall)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(align.Render(cx), style.Render(cx))
		}))
	})
}
