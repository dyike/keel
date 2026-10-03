package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("dialog", "overlays", func() core.Widget { return el.Root(newDialogGallery()) })
}

type dialogGallery struct {
	dlg, edit *kit.DialogView
	name, msg string
}

func newDialogGallery() *dialogGallery {
	g := &dialogGallery{dlg: kit.Dialog(""), msg: "还没有操作", name: "华东物流"}
	presets := kit.Menu().Item("华东物流", "", func() { g.name = "华东物流" }).Item("北方商贸", "", func() { g.name = "北方商贸" })
	presets.Trigger(kit.Button("客户预设", presets.Toggle))
	g.edit = kit.Dialog("编辑客户").CloseButton(true).OverlayClosable(false).Body(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(el.Text("客户名称").TextSize(13).TextColor(theme.Muted), el.Input().ID("dialog-name").Bind(&g.name), presets.Render(cx))
	})).Footer(
		kit.Button("取消", func() { g.edit.SetValue(false) }).Variant(kit.ButtonSecondary),
		kit.Button("保存", func() { g.edit.SetValue(false); g.msg = "已保存：" + g.name }),
	)
	return g
}

func (g *dialogGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Dialog：Esc 关闭，焦点限制在对话框内，关闭后回到按钮").Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button("确认", func() {
				g.dlg.Confirm("保存修改", "离开前保存 3 处修改吗？", func() { g.msg = "已确认" })
			}).Render(cx),
			kit.Button("删除", func() {
				g.dlg.ConfirmDanger("删除订单", "删除 SO-1001 后不能恢复。点遮罩不会关闭，Esc 等于取消。", "删除", func() { g.msg = "已删除" })
			}).Variant(kit.ButtonDanger).Render(cx),
			kit.Button("提示", func() { g.dlg.Alert("导出完成", "共 36 条记录 Exported.", nil) }).Variant(kit.ButtonSecondary).Render(cx),
			kit.Button("自定义", func() { g.edit.SetValue(true) }).Variant(kit.ButtonSecondary).Render(cx),
		),
		el.Text(g.msg).TextColor(theme.Muted),
		g.dlg.Render(cx), g.edit.Render(cx),
	)
}
