package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("input_group", "inputs", func() core.Widget {
		query := kit.Input("").Placeholder("客户或订单号").Clearable()
		status := "等待查询"
		group := kit.InputGroup("查询订单", query).Prefix(kit.Icon(kit.IconSearch)).Suffix(kit.Button("查询", func() { status = "查询：" + query.Value() }))
		amount := kit.InputGroup("金额", kit.Input("").Filter("0123456789.")).Prefix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("¥") })).Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("CNY") }))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(440)).Gap(14).Child(group.Render(cx), el.Text(status), amount.Render(cx))
		}))
	})
}
