package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("rating", "controls", func() core.Widget {
		mine := kit.Rating("你的评分", 5)
		avg := kit.Rating("平均 4/5", 5).ReadOnly()
		avg.SetValue(4)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).Items(el.Start).Child(mine.Render(cx), avg.Render(cx))
		}))
	})
}
