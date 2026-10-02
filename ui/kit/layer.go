package kit

import (
	"fmt"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// autoID gives a view a stable element ID for anchors, focus and hover
// queries, unique within the process for as long as the view lives.
func autoID(kind string, v any) string { return fmt.Sprintf("kit-%s-%p", kind, v) }

// surface is the panel of cards that sit in the page. Colors are read on
// every call, so theme.Apply takes effect on the next frame.
func surface() *el.DivEl {
	return el.Div().Bg(theme.Surface).Border(1, theme.Border).Rounded(theme.RadiusLg)
}

// floating is the panel of a layer above the page (menus, popovers,
// dropdowns, dialogs): a surface lifted by a shadow, so it reads as on top.
func floating(level theme.Elevation) *el.DivEl { return surface().Shadow(level) }

// anchor wraps a trigger in a non-interactive box that layers can anchor to.
// It adds no Tab stop and no click handler of its own.
func anchor(id string, cx *el.Context, v el.View) el.Element {
	box := el.Div().ID(id).Items(el.Start)
	if v != nil {
		box.Child(v.Render(cx))
	}
	return box
}
