package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("bubble", "data", func() core.Widget {
		say := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(10).W(el.Dp(480)).MaxW(el.Full).Child(
				kit.Bubble(say("明天上午能发货吗？")).Mine().Render(cx),
				kit.Bubble(say("可以，预计 10:00 前出库。Shipping by 10am.")).Render(cx))
		}))
	})
}
