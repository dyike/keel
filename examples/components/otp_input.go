package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("otp_input", "inputs", func() core.Widget {
		msg := "输入 6 位验证码，可以直接粘贴"
		code := kit.OtpInput("验证码", 6)
		code.OnComplete(func(s string) {
			if s == "123456" {
				msg = "验证通过"
			} else {
				code.SetError("验证码错误，试试 123456")
			}
		})
		pin := kit.OtpInput("密码 PIN · 4 位", 4).Masked(true).Groups(1).Size(40)
		pin.SetValue("1234")
		grouped := kit.OtpInput("分组 Groups · 3–2–2", 7).Groups(3).Size(56)
		grouped.SetValue("1234567")
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(code.Render(cx), el.Text(msg).TextColor(theme.Muted), pin.Render(cx), grouped.Render(cx))
		}))
	})
}
