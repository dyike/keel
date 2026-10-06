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
			return el.Text(demoText("[INFO] Entry ", "[INFO] 第 ") + strconv.Itoa(i+1) + demoText(" log entry: request handled in 12ms", " 条日志 request handled in 12ms")).TextSize(13).TextColor(theme.Muted).MaxLines(1)
		}).Height(300)
		cards := kit.VirtualList(100000, 120, func(cx *el.Context, i int) el.Element {
			return el.Div().P(8).Border(1, theme.Border).Child(el.Text(demoText("Card ", "卡片 ") + strconv.Itoa(i+1)))
		}).Horizontal(true).Width(460).Height(72)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(520)).MaxW(el.Full).Gap(12).Child(kit.Button(demoText("Scroll horizontally to the last item", "横向定位末项"), func() { cards.ScrollToEnd(cx) }).Render(cx), cards.Render(cx), el.Div().Border(1, theme.Border).Rounded(6).Px(8).Items(el.Stretch).Child(logs.Render(cx)))
		}))
	})
}
