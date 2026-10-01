package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("link", "controls", func() core.Widget {
		msg := "点击链接，或 Tab 聚焦后按回车"
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(
				kit.Link("查看订单详情 Details", func() { msg = "已打开详情" }).Render(cx),
				el.Text(msg).TextColor(theme.Muted))
		}))
	})
}
