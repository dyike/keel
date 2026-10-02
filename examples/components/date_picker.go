package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("date_picker", "inputs", func() core.Widget {
		due := kit.DatePicker("交货日期").Placeholder("选择日期")
		trip := kit.DatePicker("出差日期 Range").Range().Months(2).Placeholder("开始 – 结束")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).MaxW(el.Full).Child(due.Render(cx), trip.Render(cx)))
		}))
	})
}
