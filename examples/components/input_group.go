package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"unicode/utf8"
)

func init() {
	registerSection("input_group", "inputs", func() core.Widget {
		query := kit.Input("").Placeholder(demoText("Customer or order number", "客户或订单号")).Clearable()
		status := demoText("Waiting for query", "等待查询")
		group := kit.InputGroup(demoText("Find orders", "查询订单"), query).Prefix(kit.Icon(kit.IconSearch)).Suffix(kit.Button(demoText("Search", "查询"), func() { status = demoText("Query: ", "查询：") + query.Value() }))
		amount := kit.InputGroup(demoText("Amount", "金额"), kit.Input("").Filter("0123456789.")).Prefix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("¥") })).Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("CNY") }))
		message := kit.TextArea("").Rows(3).Placeholder(demoText("Enter order notes…", "输入订单备注…"))
		composer := kit.InputGroup(demoText("Order notes", "订单备注"), message).
			Addon("heading", kit.InputGroupBlockStart, el.ViewFunc(func(*el.Context) el.Element {
				return el.Text(demoText("Notes are saved with the order", "备注将随订单保存"))
			})).
			Addon("footer", kit.InputGroupBlockEnd, el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().WFull().Row().Items(el.Center).Gap(theme.SpaceMd).Child(
					el.Text(fmt.Sprintf(demoText("%d characters", "%d 字"), utf8.RuneCountInString(message.Value()))).Grow(),
					kit.Button(demoText("Save notes", "保存备注"), func() { status = demoText("Saved: ", "已保存：") + message.Value() }).Size(28).Render(cx))
			}))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(440)).MaxW(el.Full).Gap(14).Child(group.Render(cx), el.Text(status), amount.Render(cx), composer.Render(cx))
		}))
	})
}
