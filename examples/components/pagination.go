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
		compact := kit.Pagination(195, 10).Compact(true).Size(24)
		wide := kit.Pagination(1000, 10).VisiblePages(9).Size(36)
		wide.SetValue(50)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			s, e := p.Bounds()
			return el.Div().P(24).Gap(8).Items(el.Start).Child(p.Render(cx), el.Text(fmt.Sprintf("显示第 %d–%d 条", s+1, e)).TextColor(theme.Muted), el.Text("紧凑模式 · 24dp"), compact.Render(cx), el.Text("九个页码 · 36dp"), wide.Render(cx))
		}))
	})
}
