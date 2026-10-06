package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("title_bar", "shell", func() core.Widget {
		bar := kit.TitleBar(demoText("Project · Keel", "项目 · Keel")).Trailing(kit.Button(demoText("Share", "分享"), nil).Variant(kit.ButtonSecondary).Size(26))
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Items(el.Stretch).Child(bar.Render(cx), el.Div().P(24).Child(
				el.Text(demoText("This window has a system title bar, so TitleBar appears as a page header. Run go run ./examples/frameless to see a frameless window and window controls.", "这个窗口有系统标题栏，所以 TitleBar 只显示为页头。运行 go run ./examples/frameless 查看无边框窗口和窗口按钮。")).TextColor(theme.Muted)))
		}))
	})
}
