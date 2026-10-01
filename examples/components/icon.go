package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("icon", "controls", iconGallery) }
func iconGallery() core.Widget {
	return layout.Card(widget.Heading("矢量图标"), layout.Row(widget.Icon(widget.IconCheck), widget.Icon(widget.IconClose), widget.Icon(widget.IconPlus), widget.Icon(widget.IconSearch), widget.Icon(widget.IconCopy), widget.Icon(widget.IconChevronDown), widget.Icon(widget.IconChevronRight)), layout.Row(widget.Icon(widget.IconSearch).Size(12), widget.Icon(widget.IconSearch).Size(24).Color(theme.Primary), widget.Icon(widget.IconSearch).Size(36).Color(theme.Danger)))
}
