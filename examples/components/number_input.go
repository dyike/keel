package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("number_input", "inputs", func() core.Widget {
		qty := kit.NumberInput("数量（1–99）").Range(1, 99).Size(28)
		qty.SetValue(1)
		price := kit.NumberInput("单价").Range(0, 1e6).Decimals(2).
			StepBy(func(value float64, action kit.NumberStepAction) float64 {
				if value < 1 || value == 1 && action == kit.NumberStepActionDecrement {
					return 0.1
				}
				return 0.5
			}).Prefix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("¥") })).
			Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("元") }))
		packs := kit.NumberInput("包装数量（12 件/箱）").Range(0, 1200).Size(48).Appearance(false)
		packs.OnStep(func(e kit.NumberStepEvent) {
			delta := float64(e.Count) * 12
			if e.Action == kit.NumberStepActionDecrement {
				delta = -delta
			}
			packs.SetValue(e.Value + delta)
		})
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).W(el.Dp(240)).MaxW(el.Full).Child(qty.Render(cx), price.Render(cx), packs.Render(cx), el.Text(fmt.Sprintf("实际值：%g", price.Value())), kit.Button("填入 1.236（保留两位）", func() { price.SetValue(1.236) }).Render(cx))
		}))
	})
}
