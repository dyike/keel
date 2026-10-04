package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/markdown"
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
		s := kit.Settings().Size(kit.SettingsSizeSmall).GroupNavigation(true).
			Page(kit.SettingPage{Title: "通用", Icon: kit.IconSettings, Resettable: true, Groups: []kit.SettingGroup{
				{Title: "语言", Items: []kit.SettingItem{
					{Label: "语言 Language", Description: "框架文字随之切换", Keywords: []string{"locale", "language"}, Control: lang, Reset: func() { lang.SetValue("中文"); locale.Apply(locale.Chinese()) }},
				}},
				{Title: "外观", Footer: el.ViewFunc(func(*el.Context) el.Element { return el.Text("重置此页会恢复中文和浅色主题。") }), Items: []kit.SettingItem{
					{Label: "深色模式", Description: "立即生效，无需重新启动", DescriptionContent: markdown.New("**立即生效**，无需重新启动。"), Keywords: []string{"theme", "dark"}, Control: dark, Reset: func() { dark.SetValue(false); theme.Apply(theme.Light()) }},
				}},
			}}).
			Section("通知", kit.IconBell,
				kit.SettingItem{Label: "邮件提醒", Description: "新订单时发送邮件", Control: kit.Switch("", true)},
				kit.SettingItem{Label: "提醒间隔", Control: kit.NumberInput("").Range(1, 60)}).
			Section("隐私", kit.IconLock,
				kit.SettingItem{Label: "使用统计", Description: "匿名发送崩溃报告", Control: kit.Checkbox("", false)})
		return el.Root(s)
	})
}
