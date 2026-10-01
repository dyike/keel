package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("spinner", "controls", func() core.Widget { return el.Embed(spinnerGallery{}) })
}

type spinnerGallery struct{}

func (spinnerGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(16).Child(kit.Spinner().Label("正在加载订单 123").Render(cx), kit.Spinner().Size(32).Label("").Render(cx), el.Div().OnClick(func() { theme.SetReducedMotion(!theme.ReducedMotion) }).P(8).Child(el.Text("切换减少动画")))
}
