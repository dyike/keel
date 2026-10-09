package main

import (
	"github.com/dyike/keel/third_party/gio/io/pointer"
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
	context *kit.PopoverView
	query   string
}

func newPopoverGallery() *popoverGallery {
	g := &popoverGallery{}
	menu := kit.Menu().Item(demoText("Last seven days", "最近七天"), "", func() { g.query = demoText("Last seven days", "最近七天") }).Item(demoText("Pending orders", "待处理订单"), "", func() { g.query = demoText("Pending", "待处理") })
	menu.Trigger(kit.Button(demoText("Choose preset", "选择预设"), menu.Toggle))
	g.filters = kit.Popover(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(
			el.Text(demoText("Filter conditions", "筛选条件")).Bold(),
			menu.Render(cx),
			el.Input().ID("popover-query").Placeholder(demoText("Customer name or order number 123", "客户名称或单号 123")).Bind(&g.query),
			el.Text(demoText("Click outside or press Esc to close. Buttons underneath still respond.", "点外部或按 Esc 关闭，下面的按钮照常响应。")).TextSize(12).TextColor(theme.Muted),
		)
	})).Width(280).Offset(12).Arrow(true).PanelStyle(func(panel *el.DivEl) {
		panel.Rounded(theme.RadiusMd).Border(1, theme.Primary)
	})
	g.filters.Trigger(kit.Button(demoText("Filters", "筛选"), g.filters.Toggle).Variant(kit.ButtonSecondary).Icon(kit.IconSearch))
	g.context = kit.Popover(el.ViewFunc(func(*el.Context) el.Element {
		return el.Text(demoText("Opened by right-click. Esc or clicking outside closes it.", "右键打开的面板，按 Esc 或点外部关闭。"))
	})).MouseButton(pointer.ButtonSecondary).Width(260)
	g.context.Trigger(kit.Button(demoText("Right-click for details", "右键查看详情"), g.context.Toggle).Variant(kit.ButtonSecondary))
	return g
}

func (g *popoverGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Popover").TextSize(20).Bold(),
		g.filters.Render(cx),
		g.context.Render(cx),
	)
}
