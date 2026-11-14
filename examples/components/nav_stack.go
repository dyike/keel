package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("nav_stack", "navigation", func() core.Widget {
		nav := kit.NavStack()
		nav.Page(kit.NavPage{ID: "home", Title: demoText("Workspace", "工作区"), Create: func() el.View {
			draft := kit.Input(demoText("Draft name", "草稿名称"))
			return el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Gap(16).Child(
					el.Text(demoText("Type a draft, open details, then go back. Your draft stays here.", "输入草稿，进入详情后返回，草稿仍然保留。")),
					draft.Render(cx), kit.Button(demoText("Open details", "打开详情"), func() { nav.Push("details") }).Size(44).Render(cx),
				)
			})
		}})
		nav.Page(kit.NavPage{ID: "details", Title: demoText("Project details", "项目详情"), Create: func() el.View {
			notes := kit.Input(demoText("Notes", "备注"))
			return el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Gap(16).Child(notes.Render(cx),
					kit.Button(demoText("Open settings", "打开设置"), func() { nav.Push("settings") }).Size(44).Render(cx),
					el.Text(demoText("Back keeps the parent page. Pushing a page again creates a fresh instance.", "返回会保留父页面。重新进入已退出的页面会创建新实例。")))
			})
		}})
		nav.Page(kit.NavPage{ID: "settings", Title: demoText("Settings", "设置"), Create: func() el.View { return kit.Switch(demoText("Notifications", "通知"), true) }})
		return el.Root(nav)
	})
}
