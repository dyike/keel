package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("tabs", "controls", func() core.Widget {
		tabs := kit.Tabs().
			Add("基本", kit.Input("名称").Placeholder("切换后内容保留")).
			Add("通知", kit.Switch("邮件提醒", true)).
			Add("关于 About", el.ViewFunc(func(*el.Context) el.Element { return el.Text("Keel 组件示例 1.0") }))
		tabs.Closable(tabs.Remove).Reorderable(nil).Trailing(kit.Button("新增", func() { tabs.Add("新标签", kit.Input("内容")) }).Variant(kit.ButtonGhost))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(420)).MaxW(el.Full).Child(tabs.Render(cx))
		}))
	})
}
