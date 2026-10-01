package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("group-box", "controls", func() core.Widget {
		return el.Embed(kit.GroupBox("账户信息 Account 123").Child(kit.DescriptionList().Item("姓名", "张三 / Ada"), kit.Tag("已验证").Tone(kit.Success)).Description("账户资料与通知设置").Child(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("可在这里放任意 el 内容") })))
	})
}
