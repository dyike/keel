package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("marker", "controls", func() core.Widget { return el.Embed(markerGallery{}) })
}

type markerGallery struct{}

func (markerGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Dp(200)).Gap(16).Child(kit.Marker("在线 Online 123").Tone(kit.Success).Render(cx), kit.Marker("等待连接，请稍后重试").Tone(kit.Warning).Render(cx), kit.Marker("有新消息").Render(cx))
}
