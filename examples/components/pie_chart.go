package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("pie_chart", "data", func() core.Widget {
		data := []kit.PieSlice{{Name: demoText("East China", "华东 East"), Value: 48}, {Name: demoText("North China", "华北"), Value: 32}, {Name: demoText("South China", "华南"), Value: 20}}
		pie := kit.PieChart(data...).Title(demoText("Sales share by region", "销售地区占比")).TooltipContent(func(cx *el.Context, data kit.PieChartTooltip) el.Element {
			return el.Div().Gap(4).Child(el.Text(data.Slice.Name).Bold(), el.Text(demoText("Sales: ", "销售额：")+data.FormattedValue), el.Text(demoText("Visible region shares: ", "可见地区占比：")+data.FormattedShare).TextColor(data.Color))
		})
		donut := kit.PieChart(data...).Donut(.6).Title(demoText("Donut chart · Click the legend to toggle", "环形图 · 点击图例切换"))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(600)).MaxW(el.Full).Child(pie.Render(cx), donut.Render(cx))
		}))
	})
}
