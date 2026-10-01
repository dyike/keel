package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("tag", "controls", func() core.Widget {
		return el.Embed(&tagGallery{interactive: kit.Tag("可选择和移除").Selectable()})
	})
}

type tagGallery struct {
	interactive *kit.TagView
	removed     bool
}

func (v *tagGallery) Render(cx *el.Context) el.Element {
	v.interactive.OnRemove(func() { v.removed = true })
	root := el.Div().Gap(16).Items(el.Start).Child(
		el.Div().Row().Gap(8).Child(kit.Tag("默认 123").Render(cx), kit.Tag("主要").Color(kit.TagPrimary).Render(cx), kit.Tag("完成").Tone(kit.Success).Render(cx), kit.Tag("待检查").Tone(kit.Warning).Render(cx)),
	)
	if !v.removed {
		root.Child(v.interactive.Render(cx))
	}
	return root
}
