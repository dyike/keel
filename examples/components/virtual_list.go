package main

import (
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("virtual_list", "data", func() core.Widget {
		logs := kit.VirtualList(100000, 24, func(cx *el.Context, i int) el.Element {
			return el.Text("[INFO] 第 " + strconv.Itoa(i+1) + " 条日志 request handled in 12ms").TextSize(13).TextColor(theme.Muted).MaxLines(1)
		}).Height(300)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(520)).Child(el.Div().Border(1, theme.Border).Rounded(6).Px(8).Items(el.Stretch).Child(logs.Render(cx)))
		}))
	})
}
