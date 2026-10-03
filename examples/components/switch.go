package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("switch", "controls", func() core.Widget {
		on, off := kit.Switch("接收通知 Notifications", true), kit.Switch("不可用", false)
		off.SetDisabled(true)
		small := kit.Switch("小尺寸 · 左侧标签", true).Size(kit.SwitchSmall).LabelSide(el.Left)
		custom := kit.Switch("自定义选中颜色", true)
		dimmed := kit.Switch("禁用时保留自定义颜色", true)
		dimmed.SetDisabled(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			custom.Color(theme.Success)
			dimmed.Color(theme.Success)
			return el.Div().P(24).Gap(10).Items(el.Start).Child(on.Render(cx), off.Render(cx), small.Render(cx), custom.Render(cx), dimmed.Render(cx))
		}))
	})
}
