package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
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
				el.Text(demoText("Command palette 123", "命令面板 123")), kit.Kbd("mod+shift+p").Render(cx),
				el.Text(demoText("Confirm", "确认")), kit.Kbd("enter").Render(cx), kit.Kbd("mod+s").Plain().Render(cx))
		}
		return el.Div().P(24).Gap(16).Child(
			el.Text(demoText("Kbd · Keyboard shortcut", "Kbd 快捷键键帽")).TextSize(24).Bold(),
			el.Text(demoText("Follows the parent font size; display only, without registering a shortcut.", "随父元素字号变化；仅展示，不注册快捷键。")),
			row(12), row(16), row(24),
			el.Div().Wrap().Gap(12).Items(el.Center).Child(kit.Kbd("mod+s").Size(12).Render(cx), kit.Kbd("mod+k").Size(24).Style(func(e *el.TextEl) { e.Bg(theme.Subtle).TextColor(theme.PrimaryText) }).Render(cx)),
			el.Div().Wrap().Gap(12).Child(kit.Kbd("esc").Render(cx), kit.Kbd("alt+backspace").Render(cx), kit.Kbd("ctrl+up").Render(cx)),
			el.Div().W(el.Dp(140)).Child(kit.Kbd(demoText("Unparsed text: English 123, with a long label", "未解析的中文 English 123 长文案")).Render(cx)),
			el.Text(demoText("Action binding: press the displayed shortcut to run the action. Changing the shortcut updates both the keycap and handler.", "动作绑定：按下键帽上的键触发动作，改键后键帽和处理器一起变")).TextSize(16).Bold(),
			el.Div().Row().Items(el.Center).Gap(12).Child(
				el.Text(fmt.Sprintf(demoText("Open command palette (triggered %d times)", "打开命令面板（已触发 %d 次）"), opened)), kit.KbdFor("demo.palette").Render(cx),
				kit.Button(demoText("Change to mod+k", "改成 mod+k"), rebind("mod+k")).Variant(kit.ButtonSecondary).Render(cx),
				kit.Button(demoText("Restore mod+shift+p", "恢复 mod+shift+p"), rebind("mod+shift+p")).Variant(kit.ButtonGhost).Render(cx)),
		)
	}))
}
