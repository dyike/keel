package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("input", "inputs", func() core.Widget {
		search := kit.Input("搜索").Placeholder("客户名称或单号 123").Clearable().Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(16).Color(theme.Muted).Render(cx)
		}))
		price := kit.Input("单价").Placeholder("0.00").Filter("0123456789.").Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("元") }))
		pass := kit.Input("密码").Password()
		bad := kit.Input("邮箱 Email")
		bad.SetValue("not-an-email")
		bad.SetError("邮箱格式不正确")
		note := kit.TextArea("备注").Placeholder("多行文字，回车换行").Rows(3)
		off := kit.Input("只读")
		off.SetValue("SO-1001")
		off.SetReadOnly(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).W(el.Dp(360)).Child(search.Render(cx), price.Render(cx), pass.Render(cx), bad.Render(cx), note.Render(cx), off.Render(cx))
		}))
	})
}
