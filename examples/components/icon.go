package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("icon", "controls", func() core.Widget { return el.Embed(iconGallery{}) })
}

type iconGallery struct{}

func (iconGallery) Render(cx *el.Context) el.Element {
	r := el.Div().Wrap().Gap(16)
	for n := kit.IconCheck; n <= kit.IconEdit; n++ {
		r.Child(kit.Icon(n).Size(24).Render(cx))
	}
	return r
}
