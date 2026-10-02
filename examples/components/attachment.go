package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("attachment", "data", func() core.Widget {
		up := kit.Attachment("季度报表 Q3.xlsx", 2_400_000)
		up.SetProgress(0.6)
		up.OnCancel(func() {}).OnRetry(func() {})
		bad := kit.Attachment("合同扫描.pdf", 18_000_000)
		bad.SetError("网络中断，请重试")
		bad.OnRetry(func() {}).OnCancel(func() {})
		done := kit.Attachment("logo.png", 48_000)
		removed := false
		done.OnRemove(func() { removed = true })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			box := el.Div().P(24).Gap(10).Items(el.Start).Child(up.Render(cx), bad.Render(cx))
			if !removed {
				box.Child(done.Render(cx))
			}
			return box
		}))
	})
}
