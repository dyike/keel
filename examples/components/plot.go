package main

import (
	"math"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("plot", "data", func() core.Widget {
		var wave, damped []kit.PlotPoint
		for i := 0; i <= 200; i++ {
			x := float64(i) / 10
			wave = append(wave, kit.PlotPoint{X: x, Y: math.Sin(x)})
			damped = append(damped, kit.PlotPoint{X: x, Y: math.Exp(-x/8) * math.Cos(x*1.3)})
		}
		lines := kit.Plot(kit.PlotSeries{Name: "sin(x)", Points: wave}, kit.PlotSeries{Name: "衰减振荡", Points: damped}).
			Lines().Title("滚轮缩放 · 拖动平移 · 双击复位")
		var scatter []kit.PlotPoint
		for i := 0; i < 120; i++ {
			x := float64(i%40) + float64(i)/7
			scatter = append(scatter, kit.PlotPoint{X: x, Y: 30 + x*1.8 + 12*math.Sin(float64(i)*1.7)})
		}
		dots := kit.Plot(kit.PlotSeries{Name: "订单金额 vs 数量", Points: scatter}).Title("散点：悬停查看数值").Height(200)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(620)).Child(lines.Render(cx), dots.Render(cx))
		}))
	})
}
