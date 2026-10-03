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
		price := kit.RangeSlider("价格区间", 0, 100).Step(5)
		price.SetValues(20, 80)
		level := kit.RangeSlider("竖向区间", 0, 100).Vertical(180).Step(5)
		level.SetValues(25, 75)
		committed := "尚未提交"
		log := kit.Slider("频率 Hz · 对数刻度", 20, 20000).Scale(kit.SliderLogarithmic).
			OnRelease(func(x float64) { committed = fmt.Sprintf("已提交频率 %.1f Hz", x) })
		log.SetValue(1000)
		off := kit.Slider("不可用", 0, 1)
		off.SetDisabled(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(320)).MaxW(el.Full).Child(volume.Render(cx),
				el.Text(fmt.Sprintf("当前音量 %g", volume.Value())).TextColor(theme.Muted), temp.Render(cx), log.Render(cx), el.Text(committed), price.Render(cx), level.Render(cx), off.Render(cx))
		}))
	})
}
