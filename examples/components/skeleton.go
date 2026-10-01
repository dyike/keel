package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("skeleton", "controls", func() core.Widget { return el.Embed(skeletonGallery{}) })
}

type skeletonGallery struct{}

func (skeletonGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(16).Child(kit.Skeleton().W(el.Dp(48)).H(el.Dp(48)).Circle().Render(cx), kit.Skeleton().W(el.Dp(240)).Render(cx), kit.Skeleton().W(el.Dp(320)).H(el.Dp(24)).Shimmer().Render(cx), el.Div().OnClick(func() { theme.SetReducedMotion(!theme.ReducedMotion) }).P(8).Child(el.Text("切换减少动画")))
}
