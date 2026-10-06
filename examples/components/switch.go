package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("switch", "controls", func() core.Widget {
		on, off := kit.Switch(demoText("Receive notifications", "接收通知 Notifications"), true), kit.Switch(demoText("Unavailable", "不可用"), false)
		off.SetDisabled(true)
		small := kit.Switch(demoText("Small · Left label", "小尺寸 · 左侧标签"), true).Size(kit.SwitchSmall).LabelSide(el.Left)
		custom := kit.Switch(demoText("Custom selection color", "自定义选中颜色"), true)
		quiet := kit.Switch(demoText("Hide focus ring · Keep keyboard interaction", "隐藏焦点环 · 保留键盘操作"), false).FocusRing(false)
		on.TabIndex(1)
		small.TabIndex(2)
		dimmed := kit.Switch(demoText("Keep custom colors when disabled", "禁用时保留自定义颜色"), true)
		dimmed.SetDisabled(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			custom.Color(theme.Success)
			dimmed.Color(theme.Success)
			return el.Div().P(24).Gap(10).Items(el.Start).Child(on.Render(cx), off.Render(cx), small.Render(cx), custom.Render(cx), dimmed.Render(cx), quiet.Render(cx))
		}))
	})
}
