package main

import (
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("list", "data", func() core.Widget {
		var items []string
		for i := range 300 {
			items = append(items, "联系人 Contact "+strconv.Itoa(i+1))
		}
		msg := "点击或用 ↑ ↓ 选择，回车打开"
		var l *kit.ListView
		l = kit.List(items...).Height(240).OnActivate(func(i int) { msg = "打开 " + l.Items()[i] })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).W(el.Dp(320)).Child(l.Render(cx), el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
