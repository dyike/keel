package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("stepper", "controls", func() core.Widget {
		steps := kit.Stepper(demoText("Order details", "填写订单"), demoText("Confirm payment", "确认付款"), demoText("Ship", "发货 Ship")).TextCenter(true).Navigation(kit.StepperNavigationAll)
		steps.SetValue(1)
		vertical := kit.Stepper().Vertical().Size(32).Navigable()
		vertical.SetEntries(
			kit.StepperItem{Label: demoText("Order details", "订单详情"), Icon: kit.IconReceipt},
			kit.StepperItem{Label: demoText("Payment history", "付款记录"), Icon: kit.IconLock, Disabled: true},
			kit.StepperItem{Label: demoText("Shipping", "发货"), Icon: kit.IconInbox, Content: el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Text(demoText("Schedule shipping", "安排发货")), el.Text(demoText("Ship after confirming the address", "确认地址后寄出")))
			})},
		)
		vertical.SetValue(2)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).Items(el.Start).Child(steps.Render(cx), vertical.Render(cx), el.Div().Wrap().Gap(8).Child(
				kit.Button(demoText("Previous step", "上一步"), func() { steps.SetValue(steps.Value() - 1) }).Variant(kit.ButtonSecondary).Render(cx),
				kit.Button(demoText("Next step", "下一步"), func() { steps.SetValue(steps.Value() + 1) }).Render(cx),
			))
		}))
	})
}
