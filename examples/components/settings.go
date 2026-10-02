package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("settings", "shell", func() core.Widget {
		lang := kit.Select("", "中文", "English").OnChange(func(s string) {
			if s == "English" {
				locale.Apply(locale.English())
			} else {
				locale.Apply(locale.Chinese())
			}
		})
		lang.SetValue("中文")
		dark := kit.Switch("", theme.Current().Bg == theme.Dark().Bg).OnChange(func(on bool) {
			if on {
				theme.Apply(theme.Dark())
			} else {
				theme.Apply(theme.Light())
			}
		})
		s := kit.Settings().
			Section("通用", kit.IconUser,
				kit.SettingItem{Label: "语言 Language", Description: "框架文字随之切换", Control: lang},
				kit.SettingItem{Label: "深色模式", Description: "立即生效", Control: dark}).
			Section("通知", kit.IconInbox,
				kit.SettingItem{Label: "邮件提醒", Description: "新订单时发送邮件", Control: kit.Switch("", true)},
				kit.SettingItem{Label: "提醒间隔", Control: kit.NumberInput("").Range(1, 60)}).
			Section("隐私", kit.IconWarning,
				kit.SettingItem{Label: "使用统计", Description: "匿名发送崩溃报告", Control: kit.Checkbox("", false)})
		return el.Root(s)
	})
}
