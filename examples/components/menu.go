package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("menu", "overlays", func() core.Widget { return el.Root(newMenuGallery()) })
}

type menuGallery struct {
	menu          *kit.MenuView
	long, context *kit.MenuView
	last          string
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
	g.long = kit.Menu()
	for i := 1; i <= 60; i++ {
		label := fmt.Sprintf("操作 %02d", i)
		g.long.Item(label, "", do(label))
	}
	g.long.Trigger(kit.Button("长菜单：试试 End / Home", g.long.Toggle))
	g.context = kit.Menu().Item("复制选区", "", do("复制选区")).Item("查看详情", "", do("查看详情"))
	g.context.Trigger(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Role("button").Name("右键菜单区域").Focusable(true).P(16).Bg(theme.Subtle).OnContextMenu(g.context.Toggle).OnKey(func(e el.KeyEvent) bool {
			if e.Name == "F10" {
				if e.State == el.KeyPress {
					g.context.Toggle()
				}
				return true
			}
			return false
		}).Child(el.Text("在这里右键，或聚焦后按 F10"))
	}))
	return g
}

func (g *menuGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Menu：↑ ↓ 移动，→ 打开子菜单，← / Esc 返回").Bold(),
		g.menu.Render(cx),
		g.long.Render(cx), g.context.Render(cx),
		el.Text(g.last).TextColor(theme.Muted),
	)
}
