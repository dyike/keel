package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("badge", "controls", badgeGallery) }
func badgeGallery() core.Widget {
	badge := widget.Badge(3)
	badge.Child(widget.Button("通知", func() { badge.SetCount(0) }).Secondary())
	return layout.Card(widget.Heading("数字、圆点和图标角标"), layout.Row(badge, widget.Button("增加未读", func() { badge.SetCount(badge.Count() + 1) }).Secondary(), widget.Badge(120).Size(widget.Large)), layout.Row(widget.Badge(1).Dot().Child(widget.Text("在线")), widget.Badge(1).Icon(widget.Icon(widget.IconCheck)).Color(theme.Primary, theme.OnColor).Child(widget.Text("已验证"))))
}
