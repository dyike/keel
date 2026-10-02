package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("stepper", "controls", func() core.Widget {
		steps := kit.Stepper("填写订单", "确认付款", "发货 Ship").Navigable()
		steps.SetValue(1)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).Items(el.Start).Child(steps.Render(cx), el.Div().Wrap().Gap(8).Child(
				kit.Button("上一步", func() { steps.SetValue(steps.Value() - 1) }).Variant(kit.ButtonSecondary).Render(cx),
				kit.Button("下一步", func() { steps.SetValue(steps.Value() + 1) }).Render(cx),
			))
		}))
	})
}
