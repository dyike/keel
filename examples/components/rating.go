package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("rating", "controls", func() core.Widget {
		mine := kit.Rating(demoText("Your rating · Click a filled star again to reduce the rating", "你的评分 · 再点已填星可减分"), 5).Size(32)
		avg := kit.Rating(demoText("Average 3.7/5 · Small", "平均 3.7/5 · 小尺寸"), 5).Size(16).ReadOnly()
		avg.SetScore(3.7)
		half := kit.Rating(demoText("Average 4.5/5", "平均 4.5/5"), 5).ReadOnly()
		half.SetScore(4.5)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).Items(el.Start).Child(mine.Render(cx), avg.Render(cx), half.Color(theme.PrimaryText).Render(cx))
		}))
	})
}
