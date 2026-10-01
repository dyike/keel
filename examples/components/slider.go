package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("slider", "controls", func() core.Widget {
		volume := kit.Slider("音量", 0, 100).Step(5)
		volume.SetValue(40)
		temp := kit.Slider("温度 °C", -10, 40).Step(0.5)
		off := kit.Slider("不可用", 0, 1)
		off.SetDisabled(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(320)).Child(volume.Render(cx),
				el.Text(fmt.Sprintf("当前音量 %g", volume.Value())).TextColor(theme.Muted), temp.Render(cx), off.Render(cx))
		}))
	})
}
