package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("radar_chart", "data", func() core.Widget {
		chart := kit.RadarChart([]string{demoText("Speed", "速度"), demoText("Reliability", "可靠性"), demoText("Features", "功能"), demoText("Usability", "易用性"), demoText("Cost", "成本")}, kit.Series{Name: demoText("Option A", "方案 A"), Values: []float64{90, 85, 70, 95, 60}}, kit.Series{Name: demoText("Option B", "方案 B"), Values: []float64{70, 95, 90, 65, 85}}).Title(demoText("Compare option capabilities", "方案能力对比")).RadarMax(100).GridLevels(5).SeriesStyle(0, kit.ChartSeriesStyle{Dots: true})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(24).Items(el.Stretch).Child(chart.Render(cx)) }))
	})
}
