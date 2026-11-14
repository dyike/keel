package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
)

func init() {
	registerSection("rising_glow", "motion", func() core.Widget {
		glow := kit.RisingGlow().ParticleCount(80).Color(color.NRGBA{R: 0x13, G: 0x8d, B: 0xff, A: 255}).Height(100)
		enabled := true
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			width, _ := cx.ViewportSize()
			titleSize := float32(96)
			if width < 768 {
				titleSize = 60
			}
			return el.Div().WFull().HFull().Center().Bg(color.NRGBA{R: 9, G: 9, B: 11, A: 255}).P(24).Child(
				el.Text("keel").TextSize(titleSize).Bold().TextColor(color.NRGBA{R: 250, G: 250, B: 250, A: 255}),
				el.Div().WFull().MaxW(el.Dp(384)).Child(glow.Render(cx)),
				el.Div().Mt(32).Child(kit.Button(demoText("Toggle animation", "开关动画"), func() { enabled = !enabled; glow.Enabled(enabled) }).Render(cx)),
			)
		}))
	})
}
