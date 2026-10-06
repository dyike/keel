package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("chart", "data", func() core.Widget {
		months := []string{demoText("January", "1月"), demoText("February", "2月"), demoText("March", "3月"), demoText("April", "4月"), demoText("May", "5月"), demoText("June", "6月"), demoText("July", "7月"), demoText("August", "8月"), demoText("September", "9月"), demoText("October", "10月"), demoText("November", "11月"), demoText("December", "12月")}
		sales := kit.AreaChart(months,
			kit.Series{Name: demoText("East China", "华东"), Values: []float64{120, 132, 101, 134, 190, 230, 210, 182, 191, 234, 290, 330}},
			kit.Series{Name: demoText("North China", "华北"), Values: []float64{220, 182, 191, 234, 290, 330, 310, 201, 154, 190, 330, 410}},
			kit.Series{Name: demoText("South China", "华南 South"), Values: []float64{150, 232, 201, 154, 190, 330, 410, 320, 332, 301, 334, 390}},
		).Title(demoText("Monthly sales (CNY 10,000) · 14 fixed points", "月度销售额（万元）· 固定 14 个点位")).PointCount(14).YDomain(0, 500).YTickCount(6).XTickCount(6).GridColumns(5).GridDashed(true).Curve(kit.ChartCurveSmooth).ReferenceLines(kit.ChartReference{Value: 300, Label: demoText("Target", "目标"), Color: theme.Success})
		orders := kit.BarChart([]string{demoText("Monday", "周一"), demoText("Tuesday", "周二"), demoText("Wednesday", "周三"), demoText("Thursday", "周四"), demoText("Friday", "周五"), demoText("Saturday", "周六"), demoText("Sunday", "周日")},
			kit.Series{Name: demoText("New order", "新订单"), Values: []float64{32, 41, 38, 52, 61, 24, 18}},
			kit.Series{Name: demoText("Refund", "退款"), Values: []float64{4, 6, 3, 8, 5, 2, 1}},
		).Title(demoText("This week's orders · Labels inside the axis", "本周订单 · 轴内标签")).Height(180).YLabelsInside(true).Gutter(kit.ChartGutter{Left: 8, Right: 16, Top: 12, Bottom: 24})
		orders.BarGradient(func(d kit.ChartBarDatum, r kit.ChartBarRange) []kit.ChartColorStop {
			faded := d.Color
			faded.A = 80
			return []kit.ChartColorStop{{Position: r.ChartToBar(r.Min), Color: faded}, {Position: r.ChartToBar((r.Min + r.Max) / 2), Color: d.Color}, {Position: r.ChartToBar(r.Max), Color: theme.Success}}
		})
		stack := kit.BarChart([]string{"Q1", "Q2", "Q3", "Q4"},
			kit.Series{Name: demoText("Online", "线上"), Values: []float64{320, 410, 380, 520}},
			kit.Series{Name: demoText("Store", "门店"), Values: []float64{210, 190, 260, 300}},
			kit.Series{Name: demoText("Channel", "渠道"), Values: []float64{80, 120, 90, 160}},
		).Stacked().BarAlignment(kit.BarAlignmentLeft).Title(demoText("Quarterly revenue breakdown · Horizontal", "季度收入构成 · 横向")).Height(180)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(620)).MaxW(el.Full).Child(sales.Render(cx), orders.Render(cx), stack.Render(cx))
		}))
	})
}
