package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("dropdown_button", "overlays", func() core.Widget {
		msg := "还没有操作"
		set := func(s string) func() { return func() { msg = s } }
		export := kit.DropdownButton("导出", kit.Menu().Item("PDF", "", set("导出 PDF")).Item("CSV", "", set("导出 CSV"))).Variant(kit.ButtonSecondary)
		save := kit.DropdownButton("保存", kit.Menu().Item("另存为…", "mod+shift+s", set("另存为")).Item("保存全部", "", set("保存全部"))).Split(set("保存"))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(
				el.Text("DropdownButton：整体打开菜单，或分体（主操作 + 箭头）").Bold(),
				el.Div().Row().Gap(12).Child(export.Render(cx), save.Render(cx)),
				el.Text(msg).TextColor(theme.Muted),
			)
		}))
	})
}
