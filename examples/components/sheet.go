package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("sheet", "overlays", func() core.Widget { return el.Root(newSheetGallery()) })
}

type sheetGallery struct{ right, bottom *kit.SheetView }

func newSheetGallery() *sheetGallery {
	body := func() el.View {
		menu := kit.Menu().Item(demoText("Copy order number", "复制订单号"), "", nil).Item(demoText("View action log", "查看操作记录"), "", nil)
		menu.Trigger(kit.Button(demoText("Order actions", "订单操作"), menu.Toggle))
		return el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Gap(8).Child(el.Text(demoText("Order SO-1001", "订单 SO-1001")).Bold(), el.Text(demoText("East China Logistics · Awaiting payment · ¥300.00", "华东物流 · 待付款 · ¥300.00")).TextColor(theme.Muted),
				kit.DescriptionList().Item(demoText("Owner", "负责人"), demoText("Alex Chen", "张三")).Item(demoText("Created at", "创建时间"), "2026-10-01 09:30").Render(cx), menu.Render(cx))
		})
	}
	g := &sheetGallery{right: kit.Sheet(el.Right, demoText("Order details", "订单详情")).MarginTop(32).Body(body()), bottom: kit.Sheet(el.Bottom, demoText("Bulk actions", "批量操作")).Size(220).Body(body())}
	g.right.PanelStyle(func(e *el.DivEl) { e.P(24).Gap(theme.SpaceLg).Border(1, theme.Border) })
	g.right.Footer(kit.Button(demoText("Done", "完成"), func() { g.right.SetValue(false) }))
	g.bottom.Footer(kit.Button(demoText("Close panel", "关闭面板"), func() { g.bottom.SetValue(false) }))
	return g
}

func (g *sheetGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("Sheet: drag the inner edge to resize. Esc or clicking the backdrop closes it.", "Sheet：拖动内侧边缘调整尺寸，Esc 或点遮罩关闭")).Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button(demoText("Details on the right", "右侧详情"), func() { g.right.SetValue(true) }).Render(cx),
			kit.Button(demoText("Bottom panel", "底部面板"), func() { g.bottom.SetValue(true) }).Variant(kit.ButtonSecondary).Render(cx),
		),
		g.right.Render(cx), g.bottom.Render(cx),
	)
}
