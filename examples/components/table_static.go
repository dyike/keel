package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("table_static", "data", func() core.Widget {
		note := kit.Input("").Placeholder(demoText("Notes", "备注"))
		status := demoText("No action yet", "尚未操作")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			cell := func(text string) *el.DivEl { return kit.TableDataCell().Child(el.Text(text)) }
			return el.Div().P(24).Gap(12).Items(el.Stretch).Child(
				kit.StaticTable().Name(demoText("Recent orders", "近期订单")).Child(
					kit.TableHeader().Child(kit.TableRow().Child(kit.TableHead().Child(el.Text(demoText("Orders", "订单"))), kit.TableHead().Child(el.Text(demoText("Status", "状态"))), kit.TableHead().Items(el.End).Child(el.Text(demoText("Amount", "金额"))))),
					kit.TableBody().Child(
						kit.TableRow().ID("first").Child(cell("SO-001"), kit.TableDataCell().Child(kit.Tag(demoText("Paid", "已支付")).Render(cx)), cell("¥250.00").Items(el.End)),
						kit.TableRow().ID("second").Child(cell("SO-002"), kit.TableDataCell().Child(kit.Button(demoText("Confirm", "确认"), func() { status = demoText("SO-002 confirmed", "已确认 SO-002") }).Render(cx)), cell("¥100.00").Items(el.End)),
						kit.TableRow().ID("note").Child(kit.TableDataCell().Child(note.Render(cx))),
					),
					kit.TableFooter().Child(kit.TableRow().Decorate(nil).Child(cell(demoText("Total", "合计")), cell(""), cell("¥350.00").Items(el.End))),
					kit.TableCaption().Child(el.Text(demoText("Two orders; the notes row spans the full width.", "两笔订单；备注行使用整行宽度。"))),
				), el.Text(status),
			)
		}))
	})
}
