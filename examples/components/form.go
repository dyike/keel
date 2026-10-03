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
		f := kit.Form().VerticalLabels(true).
			Field("姓名", name, func() string {
				if len([]rune(name.Value())) < 2 {
					return "姓名至少 2 个字"
				}
				return ""
			}).
			Field("邮箱", email, func() string { return kit.Required(email.Value(), "请填写邮箱") }).
			Field("角色", role, func() string { return kit.Required(role.Value(), "请选择角色") })
		f.SetFieldOptions(0, kit.FormFieldOptions{Required: true, Description: "公开显示的姓名"})
		f.SetFieldOptions(1, kit.FormFieldOptions{Required: true, Description: "用于接收通知"})
		f.SetFieldOptions(2, kit.FormFieldOptions{Required: true, ColSpan: 2})
		var submit func(cx *el.Context)
		f.Actions(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Button("提交", func() { submit(cx) }).Loading(f.Submitting()).Render(cx)
		}), el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Button("取消提交", f.CancelSubmit).Variant(kit.ButtonSecondary).Render(cx)
		}), el.ViewFunc(func(*el.Context) el.Element { return el.Text(msg).TextColor(theme.Muted) }))
		submit = func(cx *el.Context) {
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
		}
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			width, _ := cx.ViewportSize()
			columns := 1
			if width >= 640 {
				columns = 2
			}
			f.Columns(columns)
			return el.Div().P(24).Items(el.Start).Child(el.Div().W(el.Dp(640)).MaxW(el.Full).Child(f.Render(cx)))
		}))
	})
}
