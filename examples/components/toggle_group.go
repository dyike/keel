package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("toggle_group", "controls", func() core.Widget {
		align := kit.ToggleGroup("左对齐", "居中", "右对齐").Variant(kit.ToggleOutline).Size(kit.ToggleSizeLarge).Segmented(true)
		align.SetValue("居中")
		separated := kit.ToggleGroup("日", "周", "月").Segmented(true).Gap(8).Variant(kit.ToggleOutline)
		separated.SetValue("周")
		icons := kit.ToggleGroup("收藏", "通知", "收件箱").Multiple().Segmented(true).Variant(kit.ToggleOutline).
			Item("收藏", kit.Toggle("", false).Icon(kit.IconStar)).
			Item("通知", kit.Toggle("", false).Icon(kit.IconBell)).
			Item("收件箱", kit.Toggle("收件箱", false).Icon(kit.IconInbox))
		icons.SetValue("收藏")
		style := kit.ToggleGroup("B", "I", "U").Multiple().Variant(kit.ToggleGhost).Size(kit.ToggleSizeSmall)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(align.Render(cx), style.Render(cx), separated.Render(cx), icons.Render(cx))
		}))
	})
}
