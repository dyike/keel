package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("marker", "controls", func() core.Widget { return el.Embed(markerGallery{}) })
}

type markerGallery struct{}

func (markerGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(16).Child(el.Div().Row().Gap(8).Items(el.Center).Child(kit.Marker(kit.MarkerDot).Color(theme.Success).Render(cx), el.Text("在线")), el.Div().Row().Gap(8).Items(el.Center).Child(kit.Marker(kit.MarkerSquare).Color(theme.Warning).Render(cx), el.Text("待处理")), el.Div().Row().Gap(8).Items(el.Center).Child(kit.Marker(kit.MarkerDiamond).Size(12).Color(theme.Info).Render(cx), el.Text("里程碑")))
}
