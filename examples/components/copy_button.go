package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("copy_button", "controls", func() core.Widget {
		cb := kit.CopyButton(func() string { return "SO-1001" })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(el.Text("订单号 SO-1001"), cb.Render(cx))
		}))
	})
}
