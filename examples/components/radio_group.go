package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("radio_group", "inputs", func() core.Widget {
		pay := kit.RadioGroup("付款方式", "转账", "支票", "现金 Cash")
		pay.SetValue("转账")
		pay.SetOptionDisabled("支票", true)
		plans := kit.RadioGroup("套餐", "基础", "专业")
		size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Items(el.Start).Child(pay.Render(cx), size.Render(cx),
				el.Div().Role("radiogroup").Name("套餐").Gap(12).Child(
					el.Div().P(12).Child(plans.Item("基础").Render(cx), el.Text("适合个人使用")),
					el.Div().P(12).Child(plans.Item("专业").Render(cx), el.Text("支持团队协作"))))
		}))
	})
}
