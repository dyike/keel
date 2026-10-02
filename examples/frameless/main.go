// Frameless draws its own title bar: the window has no system title bar, the
// app puts a search box and buttons in the bar, and its empty part moves the
// window. Run it and drag the bar, or use the window buttons.
package main

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type app struct {
	bar     *kit.TitleBarView
	nav     *kit.SidebarView
	palette *kit.CommandView
	search  *kit.InputView
}

func newApp() *app {
	a := &app{
		nav: kit.Sidebar().Section("",
			kit.SidebarItem{ID: "inbox", Label: "收件箱", Icon: kit.IconInbox, Badge: 4},
			kit.SidebarItem{ID: "orders", Label: "订单", Icon: kit.IconReceipt},
			kit.SidebarItem{ID: "settings", Label: "设置", Icon: kit.IconSettings}),
		palette: kit.Command(kit.CommandItem{Title: "新建订单", Shortcut: "mod+n"}, kit.CommandItem{Title: "切换深色模式", Action: func() { theme.Apply(theme.Dark()) }}),
		search: kit.Input("").Placeholder("搜索 ⌘K").Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(14).Color(theme.Muted).Render(cx)
		})),
	}
	a.nav.SetValue("inbox")
	a.bar = kit.TitleBar("Keel 无边框窗口").
		Leading(kit.Button("", func() { a.nav.SetCollapsed(!a.nav.Collapsed()) }).Name("切换侧栏").Icon(kit.IconChevronLeft).Variant(kit.ButtonGhost).Size(26)).
		Trailing(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(200)).Child(a.search.Render(cx)) }),
			kit.Button("命令", a.palette.Toggle).Variant(kit.ButtonSecondary).Size(26))
	return a
}

func (a *app) Render(cx *el.Context) el.Element {
	cx.Shortcut("mod+k", a.palette.Toggle)
	return el.Div().Items(el.Stretch).Child(
		a.bar.Render(cx),
		el.Div().Row().Grow().Items(el.Stretch).Child(a.nav.Render(cx),
			el.Div().Grow().P(24).Gap(8).Child(
				el.Text("拖动标题栏的空白处移动窗口").TextSize(18).Bold(),
				el.Text("macOS 上窗口按钮在左侧，其他平台在右侧；标题栏里的搜索框和按钮照常可用。").TextColor(theme.Muted))),
		a.palette.Render(cx),
	)
}

func main() {
	window.Open(window.Options{Title: "Keel 无边框窗口", Width: 820, Height: 520, Frameless: true, Content: el.Root(newApp())})
	window.Main()
}
