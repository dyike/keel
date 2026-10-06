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
			Section(demoText("Workspace", "工作台"), kit.SidebarItem{ID: "inbox", Label: demoText("Inbox", "收件箱"), Icon: kit.IconInbox, Badge: 12},
				kit.SidebarItem{ID: "orders", Label: demoText("Orders", "订单 Orders"), Icon: kit.IconReceipt, Children: []kit.SidebarItem{
					{ID: "pending", Label: demoText("Pending orders", "待处理订单"), Icon: kit.IconClock, Badge: 6},
					{ID: "completed", Label: demoText("Completed orders", "已完成订单"), Icon: kit.IconCheck},
					{ID: "archived", Label: demoText("Archive order", "归档订单"), Icon: kit.IconArchive, Disabled: true},
				}},
				kit.SidebarItem{ID: "calendar", Label: demoText("Schedule", "日程"), Icon: kit.IconCalendar}).
			Section(demoText("Team", "团队"), kit.SidebarItem{ID: "members", Label: demoText("Members", "成员"), Icon: kit.IconUser},
				kit.SidebarItem{ID: "saved", Label: demoText("Starred", "已收藏"), Icon: kit.IconStarOutline}).
			Section("", kit.SidebarItem{ID: "settings", Label: demoText("Preferences", "偏好设置"), Icon: kit.IconSettings})
		status, right := "", false
		nav.SetSuffix("inbox", kit.Button(demoText("Refresh", "刷新"), func() { status = demoText("Inbox refreshed", "已刷新收件箱") }).Variant(kit.ButtonGhost).Size(24))
		nav.SetContextMenu("inbox", kit.Menu().
			Item(demoText("Mark all read", "标记全部已读"), "", func() {
				nav.SetBadge("inbox", 0)
				status = demoText("All inbox messages read", "收件箱已全部读完")
			}).
			Separator().Item(demoText("Open inbox", "打开收件箱"), "", func() { nav.SetValue("inbox") }))
		nav.SetValue("inbox")
		nav.SetExpanded("orders", true)
		nav.Header(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Items(el.Center).Gap(10).Py(8).Child(kit.Avatar("Keel").Size(28).Render(cx))
			if !nav.Collapsed() {
				row.Child(el.Text(demoText("Keel workspace", "Keel 工作台")).TextSize(15).Bold())
			}
			return row
		}))
		nav.Footer(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Items(el.Center).Gap(10).Py(8).Child(kit.Avatar("Yike").Size(28).Render(cx))
			if !nav.Collapsed() {
				row.Child(el.Div().Gap(2).Child(el.Text("Yike").TextSize(13), el.Text(demoText("Personal workspace", "个人工作区")).TextSize(11).TextColor(theme.Muted)))
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
			titles := map[string]string{"pending": demoText("Pending orders", "待处理订单"), "completed": demoText("Completed orders", "已完成订单"), "archived": demoText("Archive order", "归档订单"), "inbox": demoText("Inbox", "收件箱"), "orders": demoText("Orders", "订单"), "calendar": demoText("Schedule", "日程"), "members": demoText("Team members", "团队成员"), "saved": demoText("Starred", "已收藏"), "settings": demoText("Preferences", "偏好设置")}
			body := el.Div().Grow().W(el.Dp(0)).ScrollY().Bg(theme.Surface).P(28).Gap(12).Items(el.Stretch)
			body.Child(el.Text(demoText("Workspace / ", "工作台 / ")+titles[nav.Value()]).TextSize(12).TextColor(theme.Muted),
				el.Text(titles[nav.Value()]).TextSize(24).Bold(),
				el.Text(demoText("View team activity and tasks in one place.", "集中查看团队动态与待办事项。")).TextSize(13).TextColor(theme.Muted),
				el.Div().H(el.Dp(1)).My(12).Bg(theme.Border))
			if nav.Value() == "inbox" {
				for _, item := range []struct{ title, detail string }{
					{demoText("Order awaiting confirmation", "订单待确认"), demoText("East China customer · Just now", "华东客户 · 刚刚")}, {demoText("This week's deliveries", "本周交付安排"), demoText("Project team · 20 minutes ago", "项目组 · 20 分钟前")}, {demoText("Design review updated", "设计评审已更新"), demoText("Product team · 1 hour ago", "产品团队 · 1 小时前")},
				} {
					body.Child(el.Div().Py(12).Gap(6).Child(el.Text(item.title).TextSize(14), el.Text(item.detail).TextSize(12).TextColor(theme.Muted)))
				}
			} else {
				body.Child(kit.Empty(demoText("No pending items", "暂无待处理事项")).Render(cx))
			}
			body.Child(kit.Button(demoText("Toggle sidebar position", "切换侧栏位置"), func() { right = !right }).Render(cx), el.Text(status))
			root := el.Div().Row().Items(el.Stretch)
			if right {
				nav.Side(el.Right)
				return root.Child(body, nav.Render(cx))
			}
			nav.Side(el.Left)
			return root.Child(nav.Render(cx), body)
		}))
	})
}
