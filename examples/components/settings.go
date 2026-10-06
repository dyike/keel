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
		initialLanguage := locale.Current()
		lang := kit.Select("", "中文", "English").OnChange(func(s string) {
			if s == "English" {
				locale.Apply(locale.English())
			} else {
				locale.Apply(locale.Chinese())
			}
		})
		if initialLanguage.Lang == "en" {
			lang.SetValue("English")
		} else {
			lang.SetValue("中文")
		}
		dark := kit.Switch("", theme.Current().Bg == theme.Dark().Bg).OnChange(func(on bool) {
			if on {
				theme.Apply(theme.Dark())
			} else {
				theme.Apply(theme.Light())
			}
		})
		s := kit.Settings().Size(kit.SettingsSizeSmall).GroupNavigation(true).
			Page(kit.SettingPage{Title: demoText("General", "通用"), Icon: kit.IconSettings, Resettable: true, Groups: []kit.SettingGroup{
				{Title: demoText("Language", "语言"), Items: []kit.SettingItem{
					{Label: demoText("Language", "语言 Language"), Description: demoText("Framework text follows the selected language", "框架文字随之切换"), Keywords: []string{"locale", "language"}, Control: lang, Reset: func() {
						if initialLanguage.Lang == "en" {
							lang.SetValue("English")
						} else {
							lang.SetValue("中文")
						}
						locale.Apply(initialLanguage)
					}},
				}},
				{Title: demoText("Appearance", "外观"), Footer: el.ViewFunc(func(*el.Context) el.Element {
					return el.Text(demoText("Reset restores the initial language and light theme.", "重置此页会恢复初始语言和浅色主题。"))
				}), Items: []kit.SettingItem{
					{Label: demoText("Dark mode", "深色模式"), Description: demoText("Takes effect immediately; no restart needed", "立即生效，无需重新启动"), DescriptionContent: markdown.New(demoText("**Takes effect immediately**, without restarting.", "**立即生效**，无需重新启动。")), Keywords: []string{"theme", "dark"}, Control: dark, Reset: func() { dark.SetValue(false); theme.Apply(theme.Light()) }},
				}},
			}}).
			Section(demoText("Notifications", "通知"), kit.IconBell,
				kit.SettingItem{Label: demoText("Email reminders", "邮件提醒"), Description: demoText("Send email for new orders", "新订单时发送邮件"), Control: kit.Switch("", true)},
				kit.SettingItem{Label: demoText("Reminder interval", "提醒间隔"), Control: kit.NumberInput("").Range(1, 60)}).
			Section(demoText("Privacy", "隐私"), kit.IconLock,
				kit.SettingItem{Label: demoText("Usage analytics", "使用统计"), Description: demoText("Send anonymous crash reports", "匿名发送崩溃报告"), Control: kit.Checkbox("", false)})
		return el.Root(s)
	})
}
