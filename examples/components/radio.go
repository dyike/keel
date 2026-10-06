package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("radio", "inputs", func() core.Widget {
		// Standalone radios placed freely, kept exclusive by the app.
		express, pickup := kit.Radio(demoText("Delivery", "快递配送")), kit.Radio(demoText("Store pickup", "门店自提"))
		express.SetValue(true)
		express.OnChange(func(bool) { pickup.SetValue(false) })
		pickup.OnChange(func(bool) { express.SetValue(false) })
		card := func(r *kit.RadioView, note string, cx *el.Context) el.Element {
			return el.Div().P(12).Rounded(theme.RadiusMd).Border(1, theme.Border).Gap(4).
				Child(r.Render(cx), el.Text(note).TextSize(theme.TextSm).TextColor(theme.Muted))
		}
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Items(el.Start).Child(
				el.Text(demoText("Radio", "Radio 单选按钮")).TextSize(24).Bold(),
				el.Text(demoText("Placed in separate cards; the application keeps them mutually exclusive in OnChange.", "分散在不同卡片里，由应用在 OnChange 中保持互斥。")).TextColor(theme.Muted),
				el.Div().Row().Gap(12).Child(card(express, demoText("Delivery in 1–2 days", "1–2 天送达"), cx), card(pickup, demoText("Same-day pickup", "当天可取"), cx)))
		}))
	})
}
