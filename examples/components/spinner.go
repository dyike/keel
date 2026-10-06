package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"time"
)

func init() {
	registerSection("spinner", "controls", func() core.Widget { return el.Embed(spinnerGallery{}) })
}

type spinnerGallery struct{}

func (spinnerGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(16).Child(kit.Spinner().Label(demoText("Loading orders 123", "正在加载订单 123")).Render(cx), kit.Spinner().Size(32).Label("").Render(cx), kit.Spinner().Icon(kit.IconSettings).Size(32).Color(theme.Primary).Label(demoText("Custom gear icon", "自定义齿轮图标")).Render(cx), kit.Spinner().Icon(kit.IconClock).Period(2*time.Second).Size(24).Color(theme.Muted).Label(demoText("Waiting · One cycle every two seconds", "等待处理 · 两秒一周")).Render(cx), el.Div().OnClick(func() { theme.SetReducedMotion(!theme.ReducedMotion) }).P(8).Child(el.Text(demoText("Toggle reduced motion", "切换减少动画"))))
}
