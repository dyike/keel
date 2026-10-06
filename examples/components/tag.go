package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

func init() {
	registerSection("tag", "controls", func() core.Widget {
		return el.Embed(&tagGallery{interactive: kit.Tag(demoText("Selectable and removable", "可选择和移除")).Selectable()})
	})
}

type tagGallery struct {
	interactive *kit.TagView
	removed     bool
}

func (v *tagGallery) Render(cx *el.Context) el.Element {
	v.interactive.OnRemove(func() { v.removed = true })
	custom := func(a kit.TagAppearance) kit.TagAppearance {
		a.Background = color.NRGBA{R: 110, G: 60, B: 170, A: 255}
		a.Foreground = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		a.Border = color.NRGBA{}
		return a
	}
	rich := el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Row().Gap(theme.SpaceXs).Child(kit.Icon(kit.IconCheck).Size(14).Color(theme.Success).Render(cx), el.Text(demoText("Verified", "已验证")).Bold())
	})
	root := el.Div().P(theme.SpaceXl).Gap(16).Items(el.Start).Child(
		el.Div().Row().Gap(8).Child(kit.Tag(demoText("Default 123", "默认 123")).Render(cx), kit.Tag(demoText("Primary", "主要")).Tone(kit.ToneInfo).Render(cx), kit.Tag(demoText("Done", "完成")).Tone(kit.ToneSuccess).Render(cx), kit.Tag(demoText("To check", "待检查")).Tone(kit.ToneWarning).Render(cx)),
	)
	root.Child(el.Div().Wrap().Gap(theme.SpaceSm).Child(kit.Tag(demoText("Outline", "描边")).Tone(kit.ToneInfo).Outline(true).Render(cx), kit.Tag(demoText("Square corners", "直角")).Rounded(0).Render(cx), kit.Tag(demoText("Custom colors", "自定义颜色")).Appearance(custom).Rounded(4).Render(cx)),
		el.Div().Wrap().Gap(theme.SpaceSm).Items(el.Center).Child(kit.Tag(demoText("Compact 20", "紧凑 20")).Size(20).Render(cx), kit.Tag(demoText("Standard 28", "标准 28")).Size(28).Render(cx), kit.Tag(demoText("Large 32", "大号 32")).Size(32).Render(cx)), kit.Tag(demoText("Verified", "已验证")).Tone(kit.ToneSuccess).Content(rich).Outline(true).Render(cx))
	if !v.removed {
		root.Child(v.interactive.Render(cx))
	}
	return root
}
