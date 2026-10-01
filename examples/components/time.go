package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"time"
)

func init() {
	registerSection("time", "controls", func() core.Widget { return el.Embed(&timeGallery{}) })
}

type timeGallery struct {
	visible bool
	fired   int
}

func (v *timeGallery) Render(cx *el.Context) el.Element {
	root := el.Div().Gap(16).Child(el.Div().OnClick(func() { v.visible = true }).P(8).Child(el.Text("显示 3 秒提示")), el.Div().OnClick(func() { v.visible = false }).P(8).Child(el.Text("提前取消")))
	if v.visible {
		s := cx.Scope("notice")
		s.After(3*time.Second, func() { v.visible = false; v.fired++ })
		root.Child(el.Div().ID("notice").Child(kit.Alert("即将关闭").Description("提前取消会移除节点及其定时器").Render(cx)))
	}
	return root
}
