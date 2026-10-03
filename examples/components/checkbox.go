package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("checkbox", "controls", func() core.Widget {
		all, a, b := kit.Checkbox("全选 Select all", false), kit.Checkbox("订单 SO-1001", true), kit.Checkbox("订单 SO-1002", false)
		sync := func() {
			n := 0
			for _, c := range []*kit.CheckboxView{a, b} {
				if c.Value() {
					n++
				}
			}
			all.SetValue(n == 2)
			all.SetMixed(n == 1)
		}
		a.OnChange(func(bool) { sync() })
		b.OnChange(func(bool) { sync() })
		all.OnChange(func(on bool) { a.SetValue(on); b.SetValue(on) })
		sync()
		off := kit.Checkbox("不可用", true)
		off.SetDisabled(true)
		small := kit.Checkbox("小尺寸", true).Size(14).TextSize(12)
		large := kit.Checkbox("大尺寸 · 半选", false).Size(28).TextSize(20)
		large.SetMixed(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(10).Items(el.Start).Child(all.Render(cx), el.Div().Pl(24).Gap(8).Child(a.Render(cx), b.Render(cx)), off.Render(cx), small.Render(cx), large.Render(cx))
		}))
	})
}
