package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() { registerSection("tag", "controls", func() core.Widget { return el.Embed(tagGallery{}) }) }

type tagGallery struct{}

func (tagGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(16).Items(el.Start).Child(
		el.Div().Row().Gap(8).Child(kit.Tag("默认 123").Render(cx), kit.Tag("提示").Tone(kit.Info).Render(cx), kit.Tag("完成").Tone(kit.Success).Render(cx), kit.Tag("待检查").Tone(kit.Warning).Render(cx)),
		el.Div().W(el.Dp(100)).Child(kit.Tag("Release 123 中文长标签需要换行").Render(cx)),
	)
}
