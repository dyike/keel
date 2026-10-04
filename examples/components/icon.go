package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("icon", "controls", func() core.Widget {
		svg, err := kit.SVGIcon([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 24"><path fill="currentColor" d="M2 2L30 12L2 22L8 12Z"/></svg>`))
		if err != nil {
			panic(err)
		}
		return el.Embed(iconGallery{svg: svg.Size(32).Rotate(30)})
	})
}

type iconGallery struct{ svg *kit.IconView }

func (v iconGallery) Render(cx *el.Context) el.Element {
	r := el.Div().Wrap().Gap(16)
	for n := kit.IconCheck; n <= kit.IconEdit; n++ {
		r.Child(kit.Icon(n).Size(24).Render(cx))
	}
	for _, angle := range []float32{0, 45, 90, 180, 270} {
		r.Child(kit.Icon(kit.IconChevronRight).Size(32).Rotate(angle).Render(cx))
	}
	if v.svg != nil {
		r.Child(v.svg.Render(cx))
	}
	return r
}
