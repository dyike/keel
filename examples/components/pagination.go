package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("pagination", "data", func() core.Widget {
		p := kit.Pagination(195, 10)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			s, e := p.Bounds()
			return el.Div().P(24).Gap(8).Items(el.Start).Child(p.Render(cx), el.Text(fmt.Sprintf("显示第 %d–%d 条", s+1, e)).TextColor(theme.Muted))
		}))
	})
}
