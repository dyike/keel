package main

import (
	"embed"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

// Lucide 0.468.0, ISC license in assets/lucide/LICENSE.
//
//go:embed assets/lucide/*.svg
var gooeyIcons embed.FS

func init() {
	registerSection("gooey_dock", "navigation", func() core.Widget {
		names := []string{"Home", "Messages", "Search", "Music", "Notifications", "Profile", "Settings"}
		files := []string{"house", "message-circle", "search", "music", "bell", "user", "settings"}
		items := make([]kit.GooeyDockItem, len(names))
		for i, name := range names {
			data, err := gooeyIcons.ReadFile("assets/lucide/" + files[i] + ".svg")
			if err != nil {
				panic(err)
			}
			icon, err := kit.SVGIcon(data)
			if err != nil {
				panic(err)
			}
			items[i] = kit.GooeyDockItem{ID: files[i], Label: name, IconView: icon, Active: i == 0}
		}
		dock := kit.GooeyDock(items...)
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Child(el.Div().WFull().H(el.Dp(200)).Justify(el.End).Items(el.Center).Pb(32).Child(dock.Render(cx)))
		}))
	})
}
