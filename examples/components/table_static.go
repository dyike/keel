package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("table_static", "data", func() core.Widget {
		note := kit.Input("").Placeholder("备注")
		status := "尚未操作"
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			cell := func(text string) *el.DivEl { return kit.TableDataCell().Child(el.Text(text)) }
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(
				kit.StaticTable().Name("近期订单").Child(
					kit.TableHeader().Child(kit.TableRow().Child(kit.TableHead().Child(el.Text("订单")), kit.TableHead().Child(el.Text("状态")), kit.TableHead().Items(el.End).Child(el.Text("金额")))),
					kit.TableBody().Child(
						kit.TableRow().ID("first").Child(cell("SO-001"), kit.TableDataCell().Child(kit.Tag("已支付").Render(cx)), cell("¥250.00").Items(el.End)),
						kit.TableRow().ID("second").Child(cell("SO-002"), kit.TableDataCell().Child(kit.Button("确认", func() { status = "已确认 SO-002" }).Render(cx)), cell("¥100.00").Items(el.End)),
						kit.TableRow().ID("note").Child(kit.TableDataCell().Child(note.Render(cx))),
					),
					kit.TableFooter().Child(kit.TableRow().Decorate(nil).Child(cell("合计"), cell(""), cell("¥350.00").Items(el.End))),
					kit.TableCaption().Child(el.Text("两笔订单；备注行使用整行宽度。")),
				), el.Text(status),
			)
		}))
	})
}
