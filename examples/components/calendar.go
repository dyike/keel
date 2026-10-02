package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("calendar", "inputs", func() core.Widget {
		one := kit.Calendar().DisableDates(func(t time.Time) bool { return t.Weekday() == time.Sunday })
		span := kit.Calendar().Range().Months(2).DisableDates(func(t time.Time) bool { return t.Weekday() == time.Sunday })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(24).Items(el.Start).Child(el.Text("点击月份标题快速切换年月；方向键跳过周日"), one.Render(cx), el.Text("双月范围：不能跨过禁用日期，Esc 或取消放弃草稿"), span.Render(cx))
		}))
	})
}
