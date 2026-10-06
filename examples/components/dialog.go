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
	g := &dialogGallery{dlg: kit.Dialog(""), msg: demoText("No action yet", "还没有操作"), name: demoText("East China Logistics", "华东物流")}
	presets := kit.Menu().Item(demoText("East China Logistics", "华东物流"), "", func() { g.name = demoText("East China Logistics", "华东物流") }).Item(demoText("Northern Trading", "北方商贸"), "", func() { g.name = demoText("Northern Trading", "北方商贸") })
	presets.Trigger(kit.Button(demoText("Customer presets", "客户预设"), presets.Toggle))
	g.edit = kit.Dialog(demoText("Edit customer", "编辑客户")).CloseButton(true).OverlayClosable(false).Body(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(el.Text(demoText("Customer name", "客户名称")).TextSize(13).TextColor(theme.Muted), el.Input().ID("dialog-name").Bind(&g.name), presets.Render(cx))
	})).Footer(
		kit.Button(demoText("Cancel", "取消"), func() { g.edit.SetValue(false) }).Variant(kit.ButtonSecondary),
		kit.Button(demoText("Save", "保存"), func() { g.edit.SetValue(false); g.msg = demoText("Saved: ", "已保存：") + g.name }),
	)
	return g
}

func (g *dialogGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("Dialog: Esc closes it. Focus stays inside the dialog and returns to the trigger when closed.", "Dialog：Esc 关闭，焦点限制在对话框内，关闭后回到按钮")).Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button(demoText("Confirm", "确认"), func() {
				g.dlg.Icon(kit.IconNone).Confirm(demoText("Save changes", "保存修改"), demoText("Save 3 changes before leaving?", "离开前保存 3 处修改吗？"), func() { g.msg = demoText("Confirmed", "已确认") })
			}).Render(cx),
			kit.Button(demoText("Delete", "删除"), func() {
				g.dlg.Icon(kit.IconWarning).ConfirmDanger(demoText("Delete order", "删除订单"), demoText("Deleting SO-1001 cannot be undone. Clicking the backdrop does not close it; Esc cancels.", "删除 SO-1001 后不能恢复。点遮罩不会关闭，Esc 等于取消。"), demoText("Delete", "删除"), func() { g.msg = demoText("Deleted", "已删除") })
			}).Variant(kit.ButtonDanger).Render(cx),
			kit.Button(demoText("Hint", "提示"), func() {
				g.dlg.Icon(kit.IconCheck).IconTone(kit.ToneSuccess).Alert(demoText("Export complete", "导出完成"), demoText("36 records exported.", "共 36 条记录 Exported."), nil)
			}).Variant(kit.ButtonSecondary).Render(cx),
			kit.Button(demoText("Custom", "自定义"), func() { g.edit.SetValue(true) }).Variant(kit.ButtonSecondary).Render(cx),
		),
		el.Text(g.msg).TextColor(theme.Muted),
		g.dlg.Render(cx), g.edit.Render(cx),
	)
}
