package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func init() {
	registerSection("menu", "overlays", func() core.Widget { return el.Root(newMenuGallery()) })
}

type menuGallery struct {
	menu          *kit.MenuView
	long, context *kit.MenuView
	last          string
	native        *window.MenuBar
	preview       *window.Window
	animated      bool
}

func newMenuGallery() *menuGallery {
	g := &menuGallery{last: demoText("Nothing selected", "尚未选择")}
	if err := core.Bind("gallery.menu.inspect", "f7"); err != nil {
		panic(err)
	}
	if err := core.BindIn("menu-gallery", "gallery.menu.inspect", "f6"); err != nil {
		panic(err)
	}
	do := func(s string) func() { return func() { g.last = demoText("Executed: ", "已执行：") + s } }
	export := kit.Menu().Item(demoText("PDF document", "PDF 文档"), "", do(demoText("Export PDF", "导出 PDF"))).Item(demoText("CSV spreadsheet", "CSV 表格"), "", do(demoText("Export CSV", "导出 CSV")))
	g.menu = kit.Menu().Label(demoText("Editing actions", "编辑操作")).
		IconItem(demoText("Copy", "复制"), "mod+c", kit.IconCopy, do(demoText("Copy", "复制"))).
		Item(demoText("Paste", "粘贴"), "mod+v", do(demoText("Paste", "粘贴"))).
		Item(demoText("Undo 123", "撤销 123"), "mod+z", do(demoText("Undo", "撤销"))).
		Separator().
		Label(demoText("File actions", "文件操作")).Sub(demoText("Export", "导出"), export).
		Item(demoText("Delete", "删除"), "delete", do(demoText("Delete", "删除"))).
		CheckItem(demoText("Show details", "显示详情"), "", true, func(on bool) { g.last = fmt.Sprintf(demoText("Show details: %v", "显示详情：%v"), on) })
	g.menu.ContentItem(demoText("Project details", "项目详情"), "", el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(theme.SpaceXs).Child(el.Text(demoText("Project details", "项目详情")).Bold(), el.Text(demoText("View project status and recent activity", "查看项目状态与最近活动")).TextSize(theme.TextSm).TextColor(theme.Muted))
	}), do(demoText("Project details", "项目详情")))
	g.menu.ActionItem(demoText("Context command", "上下文命令"), "gallery.menu.inspect", do(demoText("Context command", "上下文命令")))
	g.menu.Link(demoText("Keel documentation", "Keel 文档"), "https://github.com/dyike/keel")
	g.menu.SetItemDisabled(demoText("Paste", "粘贴"), true)
	g.menu.Trigger(kit.Button(demoText("More actions", "更多操作"), g.menu.Toggle).Variant(kit.ButtonSecondary))
	g.long = kit.Menu().Scrollbars(el.ScrollbarAlways)
	for i := 1; i <= 60; i++ {
		label := fmt.Sprintf(demoText("Action %02d", "操作 %02d"), i)
		g.long.Item(label, "", do(label))
	}
	g.long.Trigger(kit.Button(demoText("Long menu: try End / Home", "长菜单：试试 End / Home"), g.long.Toggle))
	g.context = kit.Menu().Item(demoText("Copy selection", "复制选区"), "", do(demoText("Copy selection", "复制选区"))).Item(demoText("View details", "查看详情"), "", do(demoText("View details", "查看详情")))
	g.context.Trigger(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Role("button").Name(demoText("Context menu region", "右键菜单区域")).Focusable(true).P(16).Bg(theme.Subtle).OnContextMenu(g.context.Toggle).OnKey(func(e el.KeyEvent) bool {
			if e.Name == "F10" {
				if e.State == el.KeyPress {
					g.context.Toggle()
				}
				return true
			}
			return false
		}).Child(el.Text(demoText("Right-click here, or focus and press F10", "在这里右键，或聚焦后按 F10")))
	}))
	return g
}

func (g *menuGallery) Render(cx *el.Context) el.Element {
	cx.ActionAt("menu-gallery", "gallery.menu.inspect", func() { g.last = demoText("Executed: context command", "已执行：上下文命令") })
	return el.Div().ID("menu-gallery").KeyContext("menu-gallery").P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("Menu: ↑ ↓ moves; → opens a submenu; ← or Esc goes back.", "Menu：↑ ↓ 移动，→ 打开子菜单，← / Esc 返回")).Bold(),
		el.Text(demoText("Context commands in this region use F6; the global default is F7", "本区域的上下文命令使用 F6；全局默认是 F7")).TextColor(theme.Muted),
		g.menu.Render(cx),
		g.long.Render(cx), g.context.Render(cx),
		el.Text(g.last).TextColor(theme.Muted),
		el.Text(demoText("Application menus: icons and platform animation controls", "应用菜单：图标与平台动画控制")).Bold(),
		kit.Button(demoText("Install native menu with icons", "安装带图标的原生菜单"), g.installNativeMenu).Render(cx),
		kit.Button(demoText("Preview animated window menu", "预览窗口菜单动画"), g.previewMenu).Render(cx),
		kit.Button(demoText("Toggle window menu animation", "切换窗口菜单动画"), func() {
			g.installNativeMenu()
			g.animated = !g.animated
			animation := window.MenuAnimation{}
			if g.animated {
				animation = window.MenuAnimation{Duration: 180 * time.Millisecond, Offset: 6}
			}
			if err := g.native.SetWindowAnimation(animation); err != nil {
				panic(err)
			}
			g.last = fmt.Sprintf(demoText("Window menu animation: %v", "窗口菜单动画：%v"), g.animated)
		}).Render(cx),
		el.Text(demoText("Native opening follows the OS; window animation respects reduced motion.", "原生展开动画跟随系统；窗口菜单动画遵守减弱动画设置。")).TextColor(theme.Muted),
	)
}

func (g *menuGallery) installNativeMenu() {
	if g.native == nil {
		art := image.NewNRGBA(image.Rect(0, 0, 32, 32))
		for y := 6; y < 26; y++ {
			for x := 6; x < 26; x++ {
				if x < 10 || y < 10 || y > 21 {
					art.SetNRGBA(x, y, color.NRGBA{R: 75, G: 130, B: 240, A: 255})
				}
			}
		}
		mask, err := window.NewMenuIcon(art, true)
		if err != nil {
			panic(err)
		}
		colored, err := window.NewMenuIcon(art, false)
		if err != nil {
			panic(err)
		}
		g.native, err = window.NewMenuBar(window.MenuItem{ID: "gallery-native", Title: demoText("Keel Menu", "Keel 菜单"), Children: []window.MenuItem{
			{ID: "native-action", Title: demoText("Action with icon", "带图标的操作"), Icon: mask, OnSelect: func() { g.last = demoText("Executed: native icon action", "已执行：原生图标操作") }},
			{ID: "native-check", Title: demoText("Checked icon", "勾选图标"), Checked: true, Icon: colored, OnSelect: func() {
				if err := g.native.UpdateItem("native-check", func(item *window.MenuItem) { item.Checked = !item.Checked }); err != nil {
					panic(err)
				}
			}},
			{Title: demoText("Unavailable icon", "禁用图标"), Disabled: true, Icon: mask},
			{Separator: true},
			{Title: demoText("Nested icons", "嵌套图标"), Icon: colored, Children: []window.MenuItem{{Title: demoText("Child action", "子菜单操作"), Icon: mask, OnSelect: func() { g.last = demoText("Executed: child icon action", "已执行：子菜单图标操作") }}}},
		}})
		if err != nil {
			panic(err)
		}
		g.animated = true
		if err := g.native.SetWindowAnimation(window.MenuAnimation{Duration: 180 * time.Millisecond, Offset: 6}); err != nil {
			panic(err)
		}
	}
	if err := g.native.Install(); err != nil {
		panic(err)
	}
}

func (g *menuGallery) previewMenu() {
	g.installNativeMenu()
	if g.preview != nil && !g.preview.Closed() {
		g.preview.Raise()
		return
	}
	g.preview = window.Open(window.Options{
		Title: demoText("Menu icons and animation", "菜单图标与动画"), Width: 520, Height: 300, MenuDisplay: window.MenuDisplayWindow,
		Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Gap(16).Child(
				el.Text(demoText("Open Keel Menu or press F10 to inspect the icons.", "打开 Keel 菜单或按 F10 查看图标。")),
				el.Text(g.last),
				kit.Button(demoText("Cancel menu tracking", "取消菜单跟踪"), func() { g.native.CancelTracking(false) }).Render(cx),
			)
		})),
		OnClose: func() { g.preview = nil },
	})
}
