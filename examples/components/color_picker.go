package main

import (
	"image/color"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("color_picker", "inputs", func() core.Widget {
		swatches := []color.NRGBA{}
		for _, c := range theme.Chart {
			swatches = append(swatches, c)
		}
		picker := kit.ColorPicker().Alpha().Swatches(swatches...)
		narrow := kit.ColorPicker().Alpha().Swatches(swatches...)
		narrow.SetValue(swatches[0])
		pop := kit.ColorPicker().Popup(true).Label("主题色").Size(kit.ColorPickerSizeSmall)
		iconPop := kit.ColorPicker().Popup(true).Label("图标触发器").Icon(kit.IconCheck).Size(kit.ColorPickerSizeLarge)
		rgbPop := kit.ColorPicker().Popup(true).Label("RGB 格式，向右弹出").Format(kit.ColorRGB).Placement(el.Right, el.Start)
		hslPop := kit.ColorPicker().Popup(true).Alpha().Label("HSL 带透明度").Format(kit.ColorHSL)
		disabled := false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			// Taller than a small window: scroll rather than squeeze every row.
			return el.Div().P(24).Gap(24).Items(el.Start).Grow().ScrollY().Child(
				picker.Render(cx), kit.Button("启用 / 禁用取色器", func() { disabled = !disabled; picker.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx),
				el.Text("内置弹层：").TextColor(theme.Muted), pop.Render(cx), iconPop.Render(cx),
				el.Text("颜色格式与弹出位置（面板里也可以切换 HEX / RGB / HSL）：").TextColor(theme.Muted), rgbPop.Render(cx), hslPop.Render(cx),
				el.Text("窄容器（160dp）：").TextColor(theme.Muted), el.Div().W(el.Dp(160)).Child(narrow.Render(cx)))
		}))
	})
}
