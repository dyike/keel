package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("progress", "controls", func() core.Widget {
		done := kit.Progress("导入订单 Import")
		done.SetValue(0.42)
		busy := kit.Progress("同步中 · 细条").Height(4).Rounded(0)
		busy.SetIndeterminate(true)
		custom := kit.Progress("自定义颜色与边框").Height(16).Rounded(2)
		custom.SetValue(.65)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(360)).MaxW(el.Full).Child(done.Render(cx), busy.Render(cx), custom.Color(theme.Success).TrackStyle(func(e *el.DivEl) { e.Border(1, theme.Border).Bg(theme.Surface) }).Render(cx),
				el.Div().Row().Gap(8).Child(
					kit.Button("+10%", func() { done.SetValue(done.Value() + 0.1) }).Variant(kit.ButtonSecondary).Render(cx),
					kit.Button("重置", func() { done.SetValue(0) }).Variant(kit.ButtonGhost).Render(cx)))
		}))
	})
}
