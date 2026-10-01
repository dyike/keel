package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("time_field", "inputs", func() core.Widget {
		start := kit.TimeField("开始时间")
		start.SetValue(9*time.Hour + 30*time.Minute)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(start.Render(cx), el.Text("可以输入 930、0930 或 9:30，回车或离开时规整为 HH:MM"))
		}))
	})
}
