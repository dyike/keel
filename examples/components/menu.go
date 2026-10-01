package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("menu", "overlays", func() core.Widget { return el.Root(newMenuGallery()) })
}

type menuGallery struct {
	menu *kit.MenuView
	last string
}

func newMenuGallery() *menuGallery {
	g := &menuGallery{last: "尚未选择"}
	do := func(s string) func() { return func() { g.last = "已执行：" + s } }
	export := kit.Menu().Item("PDF 文档", "", do("导出 PDF")).Item("CSV 表格", "", do("导出 CSV"))
	g.menu = kit.Menu().
		Item("复制", "mod+c", do("复制")).
		Item("粘贴", "mod+v", do("粘贴")).
		Item("撤销 123", "mod+z", do("撤销")).
		Separator().
		Sub("导出", export).
		Item("删除", "delete", do("删除"))
	g.menu.SetItemDisabled("粘贴", true)
	g.menu.Trigger(kit.Button("更多操作", g.menu.Toggle).Variant(kit.ButtonSecondary))
	return g
}

func (g *menuGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Menu：↑ ↓ 移动，→ 打开子菜单，← / Esc 返回").Bold(),
		g.menu.Render(cx),
		el.Text(g.last).TextColor(theme.Muted),
	)
}
