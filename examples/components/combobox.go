package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("combobox", "inputs", func() core.Widget {
		customer := kit.Combobox("客户", "华东物流", "北京百货", "深圳电子", "成都餐饮", "杭州茶业", "上海文具").Placeholder("输入筛选")
		tag := kit.Combobox("标签 Tag", "紧急", "VIP", "待跟进").AllowCustom().Placeholder("可输入新标签")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).Child(customer.Render(cx), tag.Render(cx)))
		}))
	})
}
