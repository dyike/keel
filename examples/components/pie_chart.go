package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("pie_chart", "data", func() core.Widget {
		data := []kit.PieSlice{{Name: "华东 East", Value: 48}, {Name: "华北", Value: 32}, {Name: "华南", Value: 20}}
		pie := kit.PieChart(data...).Title("销售地区占比")
		donut := kit.PieChart(data...).Donut(.6).Title("环形图 · 点击图例切换")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(600)).Child(pie.Render(cx), donut.Render(cx))
		}))
	})
}
