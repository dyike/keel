// Frameless draws its own title bar: the window has no system title bar, the
// app puts a search box and buttons in the bar, and its empty part moves the
// window. Run it and drag the bar, or use the window buttons.
package main

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

type framelessDemo struct {
	bar     *kit.TitleBarView
	nav     *kit.SidebarView
	palette *kit.CommandView
	search  *kit.InputView
}

func newFramelessDemo() *framelessDemo {
	a := &framelessDemo{
		nav: kit.Sidebar().Section("",
			kit.SidebarItem{ID: "inbox", Label: demoText("Inbox", "收件箱"), Icon: kit.IconInbox, Badge: 4},
			kit.SidebarItem{ID: "orders", Label: demoText("Orders", "订单"), Icon: kit.IconReceipt},
			kit.SidebarItem{ID: "settings", Label: demoText("Settings", "设置"), Icon: kit.IconSettings}),
		palette: kit.Command(kit.CommandItem{Title: demoText("New order", "新建订单"), Shortcut: "mod+n"}, kit.CommandItem{Title: demoText("Switch to dark mode", "切换深色模式"), Action: func() { theme.Apply(theme.Dark()) }}),
		search: kit.Input("").Placeholder(demoText("Search", "搜索")).Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(14).Color(theme.Muted).Render(cx)
		})),
	}
	a.nav.SetValue("inbox")
	a.bar = kit.TitleBar(demoText("Keel Frameless Window", "Keel 无边框窗口")).
		Leading(kit.Button("", func() { a.nav.SetCollapsed(!a.nav.Collapsed()) }).Name(demoText("Toggle sidebar", "切换侧栏")).Icon(kit.IconChevronLeft).Variant(kit.ButtonGhost).Size(26)).
		Trailing(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(200)).Child(a.search.Render(cx)) }),
			kit.Button(demoText("Commands", "命令"), a.palette.Toggle).Variant(kit.ButtonSecondary).Size(26))
	return a
}

func (a *framelessDemo) Render(cx *el.Context) el.Element {
	cx.Shortcut("mod+k", a.palette.Toggle)
	return el.Div().Items(el.Stretch).Child(
		a.bar.Render(cx),
		el.Div().Row().Grow().Items(el.Stretch).Child(a.nav.Render(cx),
			el.Div().Grow().P(24).Gap(8).Child(
				el.Text(demoText("Drag an empty area of the title bar to move the window", "拖动标题栏的空白处移动窗口")).TextSize(18).Bold(),
				el.Text(demoText("Window controls are on the left on macOS and on the right elsewhere. Search and buttons in the title bar remain interactive.", "macOS 上窗口按钮在左侧，其他平台在右侧；标题栏里的搜索框和按钮照常可用。")).TextColor(theme.Muted))),
		a.palette.Render(cx),
	)
}
