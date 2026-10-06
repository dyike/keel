package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("group_box", "controls", func() core.Widget {
		hint := el.ViewFunc(func(*el.Context) el.Element {
			return el.Text(demoText("Settings apply to this device only; the footer sits outside the content box.", "设置只用于当前设备，底部说明位于内容框外。"))
		})
		boxes := []*kit.GroupBoxView{
			kit.GroupBox(demoText("Default · Account information", "默认 · 账户信息")).Description(demoText("Surface background and border", "Surface 背景和边框")).Child(kit.DescriptionList().Item(demoText("Name", "姓名"), demoText("Alex / Ada", "张三 / Ada")), kit.Tag(demoText("Verified", "已验证")).Tone(kit.ToneSuccess)).Footer(hint),
			kit.GroupBox(demoText("Filled · Notifications", "填充 · 通知")).Variant(kit.GroupBoxFill).Child(kit.Switch(demoText("Email reminders", "邮件提醒"), true)).Footer(hint),
			kit.GroupBox(demoText("Outline · Custom style", "描边 · 自定义样式")).Variant(kit.GroupBoxOutline).Child(kit.Input(demoText("Name", "名称"))).TitleStyle(func(e *el.TextEl) { e.TextColor(theme.PrimaryText).TextSize(theme.TextLg) }).ContentStyle(func(e *el.DivEl) { e.P(theme.Space2xl).Rounded(theme.RadiusLg).Border(2, theme.Primary) }).Footer(kit.Button(demoText("Save", "保存"), nil).Variant(kit.ButtonSecondary)),
			kit.GroupBox(demoText("No decoration", "无装饰")).Variant(kit.GroupBoxNormal).Child(kit.Checkbox(demoText("Receive updates", "接收更新"), true)),
		}
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			root := el.Div().P(theme.SpaceXl).Gap(theme.Space2xl).W(el.Full)
			for _, box := range boxes {
				root.Child(box.Render(cx))
			}
			return root
		}))
	})
}
