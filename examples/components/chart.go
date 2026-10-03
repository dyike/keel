package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("chart", "data", func() core.Widget {
		months := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		sales := kit.AreaChart(months,
			kit.Series{Name: "华东", Values: []float64{120, 132, 101, 134, 190, 230, 210, 182, 191, 234, 290, 330}},
			kit.Series{Name: "华北", Values: []float64{220, 182, 191, 234, 290, 330, 310, 201, 154, 190, 330, 410}},
			kit.Series{Name: "华南 South", Values: []float64{150, 232, 201, 154, 190, 330, 410, 320, 332, 301, 334, 390}},
		).Title("月度销售额（万元）").YDomain(0, 500).YTickCount(6).XTickCount(6).GridColumns(5).GridDashed(true).Curve(kit.ChartCurveSmooth).ReferenceLines(kit.ChartReference{Value: 300, Label: "目标", Color: theme.Success})
		orders := kit.BarChart([]string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
			kit.Series{Name: "新订单", Values: []float64{32, 41, 38, 52, 61, 24, 18}},
			kit.Series{Name: "退款", Values: []float64{4, 6, 3, 8, 5, 2, 1}},
		).Title("本周订单").Height(180)
		stack := kit.BarChart([]string{"Q1", "Q2", "Q3", "Q4"},
			kit.Series{Name: "线上", Values: []float64{320, 410, 380, 520}},
			kit.Series{Name: "门店", Values: []float64{210, 190, 260, 300}},
			kit.Series{Name: "渠道", Values: []float64{80, 120, 90, 160}},
		).Stacked().Title("季度收入构成").Height(180)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(620)).MaxW(el.Full).Child(sales.Render(cx), orders.Render(cx), stack.Render(cx))
		}))
	})
}
