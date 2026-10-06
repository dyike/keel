package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("radio_group", "inputs", func() core.Widget {
		pay := kit.RadioGroup(demoText("Payment method", "付款方式"), demoText("Bank transfer", "转账"), demoText("Check", "支票"), demoText("Cash", "现金 Cash"))
		pay.SetValue(demoText("Bank transfer", "转账"))
		pay.SetOptionDisabled(demoText("Check", "支票"), true)
		plans := kit.RadioGroup(demoText("Plan", "套餐"), demoText("Basic", "基础"), demoText("Pro", "专业")).Size(28).TextSize(20)
		for _, plan := range []struct{ name, description string }{{demoText("Basic", "基础"), demoText("For personal use", "适合个人使用")}, {demoText("Pro", "专业"), demoText("Includes team collaboration", "支持团队协作")}} {
			plans.Content(plan.name, el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Gap(theme.SpaceXs).Child(el.Text(plan.name).Bold(),
					el.Text(plan.description).TextSize(theme.TextSm).TextColor(theme.Muted))
			}))
		}
		size := kit.RadioGroup(demoText("Size", "尺寸"), "S", "M", "L").Horizontal().Size(14).TextSize(12).ItemSize("M", 18, 14).ItemSize("L", 28, 20)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Items(el.Start).Child(pay.Render(cx), size.Render(cx),
				el.Div().Role("radiogroup").Name(demoText("Plan", "套餐")).Gap(12).Child(
					el.Div().P(12).Child(plans.Item(demoText("Basic", "基础")).Render(cx)),
					el.Div().P(12).Child(plans.Item(demoText("Pro", "专业")).Render(cx))))
		}))
	})
}
