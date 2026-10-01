package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("switch", "controls", func() core.Widget {
		on, off := kit.Switch("接收通知 Notifications", true), kit.Switch("不可用", false)
		off.SetDisabled(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(10).Items(el.Start).Child(on.Render(cx), off.Render(cx))
		}))
	})
}
