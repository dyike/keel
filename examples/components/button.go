package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

func init() { registerSection("button", "controls", buttonGallery) }

type buttonDemo struct {
	save    *kit.ButtonView
	toggle  *kit.ButtonView
	count   int
	loading bool
	bold    *kit.ButtonView
}

func buttonGallery() core.Widget {
	v := &buttonDemo{}
	v.save = kit.Button(demoText("Save 123", "保存 Save 123"), func() { v.count++ }).Icon(kit.IconPlus)
	v.toggle = kit.Button(demoText("Toggle loading state", "切换加载状态"), func() { v.loading = !v.loading; v.save.SetLoading(v.loading) }).Variant(kit.ButtonSecondary)
	v.bold = kit.Button(demoText("Bold", "加粗"), nil).Variant(kit.ButtonGhost).Selected(true)
	return el.Embed(v)
}
func (v *buttonDemo) Render(cx *el.Context) el.Element {
	row := func(views ...el.View) el.Element {
		d := el.Div().Wrap().Gap(12).Items(el.Center)
		for _, view := range views {
			d.Child(view.Render(cx))
		}
		return d
	}
	rich := el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Row().Gap(theme.SpaceSm).Child(el.Text(demoText("Export", "导出")), el.Text("CSV").Bold())
	})
	custom := func(a kit.ButtonAppearance) kit.ButtonAppearance {
		a.Background = color.NRGBA{R: 110, G: 60, B: 170, A: 255}
		a.Hover, a.Active = a.Background, a.Background
		a.Foreground, a.HoverForeground, a.ActiveForeground = color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		return a
	}
	disabled := kit.Button(demoText("Disabled", "不可操作"), nil)
	disabled.SetDisabled(true)
	return el.Div().P(24).Gap(16).Child(
		el.Text(demoText("Button", "Button 按钮")).TextSize(24).Bold(),
		el.Text(demoText("Tab focuses; Enter / Space activates. Loading and disabled buttons do not respond.", "Tab 聚焦；Enter / Space 激活。加载和禁用时不响应。")),
		row(kit.Button(demoText("Primary", "主要"), nil), kit.Button(demoText("Secondary", "次要"), nil).Variant(kit.ButtonSecondary), kit.Button(demoText("Lightweight", "轻量"), nil).Variant(kit.ButtonGhost), kit.Button(demoText("Delete", "删除"), nil).Variant(kit.ButtonDanger)),
		row(kit.Button(demoText("Small 28", "小 28"), nil).Size(28), kit.Button(demoText("Standard 32", "标准 32"), nil), kit.Button(demoText("Large 40", "大 40"), nil).Size(40)),
		row(kit.Button(demoText("Link", "链接"), nil).Variant(kit.ButtonLink), kit.Button(demoText("Text", "文字"), nil).Variant(kit.ButtonText), kit.Button(demoText("Success", "成功"), nil).Variant(kit.ButtonSuccess), kit.Button(demoText("Warning", "警告"), nil).Variant(kit.ButtonWarning), kit.Button(demoText("Info", "信息"), nil).Variant(kit.ButtonInfo)),
		row(kit.Button(demoText("Outline", "描边"), nil).Outline(true), kit.Button(demoText("Danger outline", "危险描边"), nil).Variant(kit.ButtonDanger).Outline(true), kit.Button(demoText("Compact", "紧凑"), nil).Compact(true), kit.Button("", nil).Name(demoText("Add", "添加")).Icon(kit.IconPlus)),
		row(kit.Button(demoText("Export CSV", "导出 CSV"), nil).Content(rich), kit.Button(demoText("Exporting", "正在导出"), nil).Content(rich).Loading(true), kit.Button(demoText("Custom palette", "自定义配色"), nil).Appearance(custom)),
		row(v.save, v.toggle),
		el.Text(fmt.Sprintf(demoText("Save count: %d", "保存次数：%d"), v.count)).TextColor(theme.Muted),
		row(kit.Button(demoText("Loading", "加载中"), nil).Loading(true), kit.Button(demoText("Loading icon", "加载图标"), nil).Icon(kit.IconPlus).Loading(true), disabled),
		el.Div().W(el.Dp(140)).Child(kit.Button(demoText("Narrow container: English 123", "窄容器中文 English 123"), nil).Render(cx)),
		el.Text(demoText("Selected state", "选中状态")).TextSize(16).Bold(),
		row(kit.Button(demoText("Selected", "已选中"), nil).Selected(true), kit.Button(demoText("Secondary selected", "次要已选中"), nil).Variant(kit.ButtonSecondary).Selected(true),
			kit.Button(demoText("Outline selected", "描边已选中"), nil).Outline(true).Selected(true), v.bold),
	)
}
