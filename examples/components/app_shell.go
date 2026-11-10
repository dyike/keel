package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("app_shell", "shell", func() core.Widget {
		return el.Root(kit.AppShell(demoText("Workspace", "工作区")).
			Page(kit.AppPage{ID: "projects", Title: demoText("Projects", "项目"), Description: demoText("Switch pages; your input is retained.", "切换页面后，输入内容仍保留。"), Icon: kit.IconHome, Create: func() el.View { return kit.Input(demoText("Project name", "项目名称")) }}).
			Page(kit.AppPage{ID: "settings", Title: demoText("Settings", "设置"), Icon: kit.IconSettings, Create: func() el.View { return kit.Switch(demoText("Notifications", "通知"), true) }}))
	})
}
