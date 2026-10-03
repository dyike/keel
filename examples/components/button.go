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
}

func buttonGallery() core.Widget {
	v := &buttonDemo{}
	v.save = kit.Button("保存 Save 123", func() { v.count++ }).Icon(kit.IconPlus)
	v.toggle = kit.Button("切换加载状态", func() { v.loading = !v.loading; v.save.SetLoading(v.loading) }).Variant(kit.ButtonSecondary)
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
		return el.Div().Row().Gap(theme.SpaceSm).Child(el.Text("导出"), el.Text("CSV").Bold())
	})
	custom := func(a kit.ButtonAppearance) kit.ButtonAppearance {
		a.Background = color.NRGBA{R: 110, G: 60, B: 170, A: 255}
		a.Hover, a.Active = a.Background, a.Background
		a.Foreground, a.HoverForeground, a.ActiveForeground = color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		return a
	}
	disabled := kit.Button("不可操作", nil)
	disabled.SetDisabled(true)
	return el.Div().P(24).Gap(16).Child(
		el.Text("Button 按钮").TextSize(24).Bold(),
		el.Text("Tab 聚焦；Enter / Space 激活。加载和禁用时不响应。"),
		row(kit.Button("主要", nil), kit.Button("次要", nil).Variant(kit.ButtonSecondary), kit.Button("轻量", nil).Variant(kit.ButtonGhost), kit.Button("删除", nil).Variant(kit.ButtonDanger)),
		row(kit.Button("小 28", nil).Size(28), kit.Button("标准 32", nil), kit.Button("大 40", nil).Size(40)),
		row(kit.Button("链接", nil).Variant(kit.ButtonLink), kit.Button("文字", nil).Variant(kit.ButtonText), kit.Button("成功", nil).Variant(kit.ButtonSuccess), kit.Button("警告", nil).Variant(kit.ButtonWarning), kit.Button("信息", nil).Variant(kit.ButtonInfo)),
		row(kit.Button("描边", nil).Outline(true), kit.Button("危险描边", nil).Variant(kit.ButtonDanger).Outline(true), kit.Button("紧凑", nil).Compact(true), kit.Button("", nil).Name("添加").Icon(kit.IconPlus)),
		row(kit.Button("导出 CSV", nil).Content(rich), kit.Button("正在导出", nil).Content(rich).Loading(true), kit.Button("自定义配色", nil).Appearance(custom)),
		row(v.save, v.toggle),
		el.Text(fmt.Sprintf("保存次数：%d", v.count)).TextColor(theme.Muted),
		row(kit.Button("加载中", nil).Loading(true), kit.Button("加载图标", nil).Icon(kit.IconPlus).Loading(true), disabled),
		el.Div().W(el.Dp(140)).Child(kit.Button("窄容器中文 English 123", nil).Render(cx)),
	)
}
