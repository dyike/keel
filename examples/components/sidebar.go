package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("sidebar", "shell", func() core.Widget {
		nav := kit.Sidebar().
			Section("工作台", kit.SidebarItem{ID: "inbox", Label: "收件箱", Icon: kit.IconInbox, Badge: 12},
				kit.SidebarItem{ID: "orders", Label: "订单 Orders", Icon: kit.IconCopy},
				kit.SidebarItem{ID: "calendar", Label: "日程", Icon: kit.IconCalendar}).
			Section("", kit.SidebarItem{ID: "settings", Label: "设置", Icon: kit.IconUser})
		nav.SetValue("inbox")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Items(el.Stretch).Child(nav.Render(cx),
				el.Div().Grow().P(24).Child(el.Text("当前页面："+nav.Value()).TextColor(theme.Muted)))
		}))
	})
}
