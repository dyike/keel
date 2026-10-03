package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("popover", "overlays", func() core.Widget { return el.Root(newPopoverGallery()) })
}

type popoverGallery struct {
	filters *kit.PopoverView
	query   string
}

func newPopoverGallery() *popoverGallery {
	g := &popoverGallery{}
	menu := kit.Menu().Item("最近七天", "", func() { g.query = "最近七天" }).Item("待处理订单", "", func() { g.query = "待处理" })
	menu.Trigger(kit.Button("选择预设", menu.Toggle))
	g.filters = kit.Popover(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(
			el.Text("筛选条件").Bold(),
			menu.Render(cx),
			el.Input().ID("popover-query").Placeholder("客户名称或单号 123").Bind(&g.query),
			el.Text("点外部或按 Esc 关闭，下面的按钮照常响应。").TextSize(12).TextColor(theme.Muted),
		)
	})).Width(280).Offset(12)
	g.filters.Trigger(kit.Button("筛选", g.filters.Toggle).Variant(kit.ButtonSecondary).Icon(kit.IconSearch))
	return g
}

func (g *popoverGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Popover").TextSize(20).Bold(),
		g.filters.Render(cx),
	)
}
