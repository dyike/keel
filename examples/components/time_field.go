package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("time_field", "inputs", func() core.Widget {
		start := kit.TimeField("开始时间 · 小").Segmented().Size(kit.TimeFieldSizeSmall)
		seconds := kit.TimeField("精确到秒 · 分段快捷键").Seconds().Hour12(false).SegmentKeys(true)
		seconds.SetValue(23*time.Hour + 59*time.Minute + 58*time.Second)
		ampm := kit.TimeField("12 小时制 · 大").Hour12(true).SegmentKeys(true).Size(kit.TimeFieldSizeLarge)
		ampm.SetValue(13*time.Hour + 30*time.Minute)
		start.SetValue(9*time.Hour + 30*time.Minute)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(start.Render(cx), seconds.Render(cx), ampm.Render(cx), el.Text("分段快捷键模式：左右切段，两位数字自动跳段；a/p 切换时段，删除重置。普通模式仍按 Enter 或移焦提交。"))
		}))
	})
}
