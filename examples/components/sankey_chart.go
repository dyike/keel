package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("sankey_chart", "data", func() core.Widget {
		chart := kit.SankeyChart([]kit.SankeyNode{{Name: "营收"}, {Name: "成本"}, {Name: "毛利"}, {Name: "运营"}, {Name: "净利"}}, []kit.SankeyLink{{Source: 0, Target: 1, Value: 55}, {Source: 0, Target: 2, Value: 45}, {Source: 2, Target: 3, Value: 20}, {Source: 2, Target: 4, Value: 25}}).Title("收入流向").NodeRadius(2).MinLinkWidth(3).Height(350)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(24).Items(el.Stretch).Child(chart.Render(cx)) }))
	})
}
