package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("toggle_group", "controls", func() core.Widget {
		align := kit.ToggleGroup(demoText("Align left", "左对齐"), demoText("Center", "居中"), demoText("Align right", "右对齐")).Variant(kit.ToggleOutline).Size(kit.ToggleSizeLarge).Segmented(true)
		align.SetValue(demoText("Center", "居中"))
		separated := kit.ToggleGroup(demoText("Day", "日"), demoText("Week", "周"), demoText("Month", "月")).Segmented(true).Gap(8).Variant(kit.ToggleOutline)
		separated.SetValue(demoText("Week", "周"))
		icons := kit.ToggleGroup(demoText("Star", "收藏"), demoText("Notifications", "通知"), demoText("Inbox", "收件箱")).Multiple().Segmented(true).Variant(kit.ToggleOutline).
			Item(demoText("Star", "收藏"), kit.Toggle("", false).Icon(kit.IconStar)).
			Item(demoText("Notifications", "通知"), kit.Toggle("", false).Icon(kit.IconBell)).
			Item(demoText("Inbox", "收件箱"), kit.Toggle(demoText("Inbox", "收件箱"), false).Icon(kit.IconInbox))
		icons.SetValue(demoText("Star", "收藏"))
		style := kit.ToggleGroup("B", "I", "U").Multiple().Variant(kit.ToggleGhost).Size(kit.ToggleSizeSmall)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(align.Render(cx), style.Render(cx), separated.Render(cx), icons.Render(cx))
		}))
	})
}
