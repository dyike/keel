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
			Section("团队", kit.SidebarItem{ID: "members", Label: "成员", Icon: kit.IconUser},
				kit.SidebarItem{ID: "saved", Label: "已收藏", Icon: kit.IconStarOutline}).
			Section("", kit.SidebarItem{ID: "settings", Label: "偏好设置", Icon: kit.IconInfo})
		nav.SetValue("inbox")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			titles := map[string]string{"inbox": "收件箱", "orders": "订单", "calendar": "日程", "members": "团队成员", "saved": "已收藏", "settings": "偏好设置"}
			body := el.Div().Grow().Bg(theme.Surface).P(28).Gap(12).Items(el.Stretch)
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
