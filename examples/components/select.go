package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("select", "inputs", func() core.Widget {
		status := kit.Select("状态", "待付款", "已付款", "已发货", "已完成")
		city := kit.Select("城市 City", "北京", "上海", "广州", "深圳", "杭州", "成都", "Chicago").Searchable()
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).Child(status.Render(cx), city.Render(cx)))
		}))
	})
}
