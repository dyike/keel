package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() { registerSection("kbd", "controls", kbdGallery) }
func kbdGallery() core.Widget {
	return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		row := func(size float32) el.Element {
			return el.Div().Wrap().Items(el.Center).Gap(12).TextSize(size).Child(
				el.Text("命令面板 123"), kit.Kbd("mod+shift+p").Render(cx),
				el.Text("确认"), kit.Kbd("enter").Render(cx), kit.Kbd("mod+s").Plain().Render(cx))
		}
		return el.Div().P(24).Gap(16).Child(
			el.Text("Kbd 快捷键键帽").TextSize(24).Bold(),
			el.Text("随父元素字号变化；仅展示，不注册快捷键。"),
			row(12), row(16), row(24),
			el.Div().Wrap().Gap(12).Child(kit.Kbd("esc").Render(cx), kit.Kbd("alt+backspace").Render(cx), kit.Kbd("ctrl+up").Render(cx)),
			el.Div().W(el.Dp(140)).Child(kit.Kbd("未解析的中文 English 123 长文案").Render(cx)),
		)
	}))
}
