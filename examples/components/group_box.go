package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("group-box", "controls", func() core.Widget {
		return el.Embed(kit.GroupBox("账户信息 Account 123", kit.DescriptionList(kit.Description{Label: "姓名", Text: "张三 / Ada"}), kit.Tag("已验证").Tone(kit.Success)))
	})
}
