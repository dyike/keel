package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("time_field", "inputs", func() core.Widget {
		start := kit.TimeField("开始时间").Segmented()
		seconds := kit.TimeField("精确到秒").Seconds().Hour12(false)
		seconds.SetValue(23*time.Hour + 59*time.Minute + 58*time.Second)
		ampm := kit.TimeField("12 小时制").Hour12(true)
		ampm.SetValue(13*time.Hour + 30*time.Minute)
		start.SetValue(9*time.Hour + 30*time.Minute)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(start.Render(cx), seconds.Render(cx), ampm.Render(cx), el.Text("Tab 切换时、分、秒；↑ ↓ 调整当前段，PageUp / PageDown 调整十单位。"))
		}))
	})
}
