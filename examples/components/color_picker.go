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
		popover := kit.ColorPicker()
		var pop *kit.PopoverView
		pop = kit.Popover(popover).Trigger(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Items(el.Center).Gap(8).Child(
				el.Div().Size(el.Dp(28)).Rounded(6).Border(1, theme.Border).Bg(popover.Value()),
				kit.Button("选择主题色", pop.Toggle).Variant(kit.ButtonSecondary).Render(cx))
		}))
		disabled := false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(24).Items(el.Start).Child(
				picker.Render(cx), kit.Button("启用 / 禁用取色器", func() { disabled = !disabled; picker.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx),
				el.Text("放进 Popover：").TextColor(theme.Muted), pop.Render(cx))
		}))
	})
}
