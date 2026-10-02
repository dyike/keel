package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() { registerSection("kbd", "controls", kbdGallery) }
func kbdGallery() core.Widget {
	core.Bind("demo.palette", "mod+shift+p")
	opened := 0
	return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Action("demo.palette", func() { opened++ })
		rebind := func(chord string) func() { return func() { core.Bind("demo.palette", chord) } }
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
			el.Text("动作绑定：按下键帽上的键触发动作，改键后键帽和处理器一起变").TextSize(16).Bold(),
			el.Div().Row().Items(el.Center).Gap(12).Child(
				el.Text(fmt.Sprintf("打开命令面板（已触发 %d 次）", opened)), kit.KbdFor("demo.palette").Render(cx),
				kit.Button("改成 mod+k", rebind("mod+k")).Variant(kit.ButtonSecondary).Render(cx),
				kit.Button("恢复 mod+shift+p", rebind("mod+shift+p")).Variant(kit.ButtonGhost).Render(cx)),
		)
	}))
}
