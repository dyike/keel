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
				kit.SidebarItem{ID: "orders", Label: "订单 Orders", Icon: kit.IconCopy, Children: []kit.SidebarItem{
					{ID: "pending", Label: "待处理订单", Icon: kit.IconClock, Badge: 6},
					{ID: "completed", Label: "已完成订单", Icon: kit.IconCheck},
					{ID: "archived", Label: "归档订单", Icon: kit.IconInbox, Disabled: true},
				}},
				kit.SidebarItem{ID: "calendar", Label: "日程", Icon: kit.IconCalendar}).
			Section("团队", kit.SidebarItem{ID: "members", Label: "成员", Icon: kit.IconUser},
				kit.SidebarItem{ID: "saved", Label: "已收藏", Icon: kit.IconStarOutline}).
			Section("", kit.SidebarItem{ID: "settings", Label: "偏好设置", Icon: kit.IconInfo})
		nav.SetValue("inbox")
		nav.SetExpanded("orders", true)
		nav.Header(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Items(el.Center).Gap(10).Py(8).Child(kit.Avatar("Keel").Size(28).Render(cx))
			if !nav.Collapsed() {
				row.Child(el.Text("Keel 工作台").TextSize(15).Bold())
			}
			return row
		}))
		nav.Footer(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Items(el.Center).Gap(10).Py(8).Child(kit.Avatar("Yike").Size(28).Render(cx))
			if !nav.Collapsed() {
				row.Child(el.Div().Gap(2).Child(el.Text("Yike").TextSize(13), el.Text("个人工作区").TextSize(11).TextColor(theme.Muted)))
			}
			return row
		}))
		initialized, wasNarrow := false, false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			width, height := cx.ViewportSize()
			narrow := width < 600
			if !initialized || narrow != wasNarrow {
				nav.SetCollapsed(narrow)
				initialized, wasNarrow = true, narrow
			}
			nav.Height(height)
			titles := map[string]string{"pending": "待处理订单", "completed": "已完成订单", "archived": "归档订单", "inbox": "收件箱", "orders": "订单", "calendar": "日程", "members": "团队成员", "saved": "已收藏", "settings": "偏好设置"}
			body := el.Div().Grow().W(el.Dp(0)).ScrollY().Bg(theme.Surface).P(28).Gap(12).Items(el.Stretch)
			body.Child(el.Text("工作台 / "+titles[nav.Value()]).TextSize(12).TextColor(theme.Muted),
				el.Text(titles[nav.Value()]).TextSize(24).Bold(),
				el.Text("集中查看团队动态与待办事项。").TextSize(13).TextColor(theme.Muted),
				el.Div().H(el.Dp(1)).My(12).Bg(theme.Border))
			if nav.Value() == "inbox" {
				for _, item := range []struct{ title, detail string }{
					{"订单待确认", "华东客户 · 刚刚"}, {"本周交付安排", "项目组 · 20 分钟前"}, {"设计评审已更新", "产品团队 · 1 小时前"},
				} {
					body.Child(el.Div().Py(12).Gap(6).Child(el.Text(item.title).TextSize(14), el.Text(item.detail).TextSize(12).TextColor(theme.Muted)))
				}
			} else {
				body.Child(kit.Empty("暂无待处理事项").Render(cx))
			}
			return el.Div().Row().Items(el.Stretch).Child(nav.Render(cx), body)
		}))
	})
}
