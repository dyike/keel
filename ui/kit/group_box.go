package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// GroupBoxView visually and semantically groups views under a title.
type GroupBoxView struct {
	title       string
	description string
	children    []el.View
}

func GroupBox(title string) *GroupBoxView                  { return &GroupBoxView{title: title} }
func (v *GroupBoxView) Description(s string) *GroupBoxView { v.description = s; return v }
func (v *GroupBoxView) Child(views ...el.View) *GroupBoxView {
	v.children = append(v.children, views...)
	return v
}
func (v *GroupBoxView) SetTitle(s string) { v.title = s }
func (v *GroupBoxView) SetChildren(children ...el.View) {
	v.children = append([]el.View(nil), children...)
}
func (v *GroupBoxView) Render(cx *el.Context) el.Element {
	box := el.Div().Role("group").Name(v.title).Gap(theme.SpaceMd)
	if v.title != "" {
		box.Child(el.Text(v.title).Bold().TextColor(theme.Text))
	}
	if v.description != "" {
		box.Child(el.Text(v.description).TextColor(theme.Muted))
	}
	content := el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(theme.Surface)
	for _, child := range v.children {
		if child != nil {
			content.Child(child.Render(cx))
		}
	}
	return box.Child(content)
}
