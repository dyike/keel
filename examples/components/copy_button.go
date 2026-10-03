package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("copy_button", "controls", func() core.Widget {
		cb := kit.CopyButton(func() string { return "SO-1001" })
		last := "尚未复制"
		custom := kit.CopyButton(func() string { return "SO-1001" }).OnCopied(func(value string) { last = "最近复制：" + value })
		custom.Content(el.ViewFunc(func(cx *el.Context) el.Element {
			label, icon := "复制订单号", kit.IconCopy
			if custom.Copied() {
				label, icon = "订单号已复制", kit.IconCheck
			}
			return el.Div().Row().Gap(8).Items(el.Center).Child(kit.Icon(icon).Render(cx), el.Text(label))
		}))
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(8).Items(el.Start).Child(el.Text("订单号 SO-1001"), cb.Render(cx), custom.Render(cx), el.Text(last))
		}))
	})
}
