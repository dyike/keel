package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
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
		d := el.Div().Row().Gap(12).Items(el.Center)
		for _, view := range views {
			d.Child(view.Render(cx))
		}
		return d
	}
	disabled := kit.Button("不可操作", nil)
	disabled.SetDisabled(true)
	return el.Div().P(24).Gap(16).Child(
		el.Text("Button 按钮").TextSize(24).Bold(),
		el.Text("Tab 聚焦；Enter / Space 激活。加载和禁用时不响应。"),
		row(kit.Button("主要", nil), kit.Button("次要", nil).Variant(kit.ButtonSecondary), kit.Button("轻量", nil).Variant(kit.ButtonGhost), kit.Button("删除", nil).Variant(kit.ButtonDanger)),
		row(kit.Button("小 28", nil).Size(28), kit.Button("标准 32", nil), kit.Button("大 40", nil).Size(40)),
		row(v.save, v.toggle),
		el.Text(fmt.Sprintf("保存次数：%d", v.count)).TextColor(theme.Muted),
		row(kit.Button("加载中", nil).Loading(true), kit.Button("加载图标", nil).Icon(kit.IconPlus).Loading(true), disabled),
		el.Div().W(el.Dp(140)).Child(kit.Button("窄容器中文 English 123", nil).Render(cx)),
	)
}
