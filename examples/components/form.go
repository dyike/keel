package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"time"
)

func init() {
	registerSection("form", "inputs", func() core.Widget {
		name := kit.Input("").Placeholder("至少 2 个字")
		email := kit.Input("").Placeholder("name@example.com")
		role := kit.Select("", "管理员", "成员", "访客")
		msg := ""
		f := kit.Form().
			Field("姓名", name, func() string {
				if len([]rune(name.Value())) < 2 {
					return "姓名至少 2 个字"
				}
				return ""
			}).
			Field("邮箱", email, func() string { return kit.Required(email.Value(), "请填写邮箱") }).
			Field("角色", role, func() string { return kit.Required(role.Value(), "请选择角色") })
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(16).W(el.Dp(420)).Child(f.Render(cx), el.Div().Row().Gap(12).Items(el.Center).Child(
				kit.Button("提交", func() {
					token := f.BeginSubmit(cx)
					if token == 0 {
						return
					}
					submittedName := name.Value()
					msg = "正在校验并提交…"
					go func() {
						time.Sleep(600 * time.Millisecond)
						core.Update(func() {
							var errors []string
							if submittedName == "admin" {
								errors = []string{"此名称已被占用"}
							}
							if f.FinishSubmit(token, errors) {
								if len(errors) > 0 {
									msg = "请修正字段"
								} else {
									msg = "已提交 " + submittedName
								}
							}
						})
					}()
				}).Loading(f.Submitting()).Render(cx), kit.Button("取消提交", f.CancelSubmit).Variant(kit.ButtonGhost).Render(cx), el.Text(msg).TextColor(theme.Muted))))
		}))
	})
}
