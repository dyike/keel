package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("status-bar", "controls", func() core.Widget { return el.Embed(statusBarGallery{}) })
}

type statusBarGallery struct{}

func (statusBarGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Full).Gap(16).Child(kit.StatusBar().Left(el.Text("连接正常 Connected")).Right(el.Text("共 123 条记录")).Render(cx), el.Div().W(el.Dp(180)).Child(kit.StatusBar().Left(el.Text("离线，等待重新连接")).Right(el.Text("3 个任务待发送")).Render(cx)))
}
