package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("progress", "controls", func() core.Widget {
		done := kit.Progress("导入订单 Import")
		done.SetValue(0.42)
		busy := kit.Progress("同步中")
		busy.SetIndeterminate(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(360)).MaxW(el.Full).Child(done.Render(cx), busy.Render(cx),
				el.Div().Row().Gap(8).Child(
					kit.Button("+10%", func() { done.SetValue(done.Value() + 0.1) }).Variant(kit.ButtonSecondary).Render(cx),
					kit.Button("重置", func() { done.SetValue(0) }).Variant(kit.ButtonGhost).Render(cx)))
		}))
	})
}
