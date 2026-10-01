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
		size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Items(el.Start).Child(pay.Render(cx), size.Render(cx))
		}))
	})
}
