package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("badge", "controls", func() core.Widget {
		unread := kit.Badge(9)
		unread.Child(kit.Button("通知", func() { unread.SetValue(0) }).Variant(kit.ButtonSecondary))
		online := kit.Badge(1).Dot().Tone(kit.ToneSuccess).Child(kit.Avatar("张三"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Wrap().Gap(24).Items(el.Center).Child(
				unread.Render(cx),
				kit.Button("增加未读", func() { unread.SetValue(unread.Value() + 1) }).Variant(kit.ButtonGhost).Render(cx),
				kit.Badge(120).Render(cx), online.Render(cx))
		}))
	})
}
