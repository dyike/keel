package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("number_input", "inputs", func() core.Widget {
		qty := kit.NumberInput("数量（1–99）").Range(1, 99)
		qty.SetValue(1)
		price := kit.NumberInput("单价").Range(0, 1e6).Step(0.5).Decimals(2)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).W(el.Dp(240)).Child(qty.Render(cx), price.Render(cx))
		}))
	})
}
