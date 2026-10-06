package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("dropdown_button", "overlays", func() core.Widget {
		msg := demoText("No action yet", "还没有操作")
		set := func(s string) func() { return func() { msg = s } }
		export := kit.DropdownButton(demoText("Export", "导出"), kit.Menu().Item("PDF", "", set(demoText("Export PDF", "导出 PDF"))).Item("CSV", "", set(demoText("Export CSV", "导出 CSV")))).Variant(kit.ButtonSecondary)
		save := kit.DropdownButton(demoText("Save", "保存"), kit.Menu().Item(demoText("Save as…", "另存为…"), "mod+shift+s", set(demoText("Save as", "另存为"))).Item(demoText("Save all", "保存全部"), "", set(demoText("Save all", "保存全部")))).Split(set(demoText("Save", "保存")))
		busy := kit.DropdownButton(demoText("Saving", "保存中"), kit.Menu().Item(demoText("View records", "查看记录"), "", set(demoText("View records", "查看记录")))).Button(kit.Button(demoText("Saving", "保存中"), nil).Icon(kit.IconDone).Loading(true).Variant(kit.ButtonSecondary).Size(40))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(
				el.Text(demoText("DropdownButton: use the whole button as a menu trigger or split it into a primary action and an arrow.", "DropdownButton：整体打开菜单，或分体（主操作 + 箭头）")).Bold(),
				el.Div().Row().Gap(12).Child(export.Render(cx), save.Render(cx)),
				busy.Render(cx),
				el.Text(msg).TextColor(theme.Muted),
			)
		}))
	})
}
