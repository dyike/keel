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
		menu := kit.Menu().Item("复制订单号", "", nil).Item("查看操作记录", "", nil)
		menu.Trigger(kit.Button("订单操作", menu.Toggle))
		return el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Gap(8).Child(el.Text("订单 SO-1001").Bold(), el.Text("华东物流 · 待付款 · ¥300.00").TextColor(theme.Muted),
				kit.DescriptionList().Item("负责人", "张三").Item("创建时间", "2026-10-01 09:30").Render(cx), menu.Render(cx))
		})
	}
	g := &sheetGallery{right: kit.Sheet(el.Right, "订单详情").Body(body()), bottom: kit.Sheet(el.Bottom, "批量操作").Size(220).Body(body())}
	g.right.Footer(kit.Button("完成", func() { g.right.SetValue(false) }))
	g.bottom.Footer(kit.Button("关闭面板", func() { g.bottom.SetValue(false) }))
	return g
}

func (g *sheetGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Sheet：从窗口边缘滑入，Esc 或点遮罩关闭").Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button("右侧详情", func() { g.right.SetValue(true) }).Render(cx),
			kit.Button("底部面板", func() { g.bottom.SetValue(true) }).Variant(kit.ButtonSecondary).Render(cx),
		),
		g.right.Render(cx), g.bottom.Render(cx),
	)
}
