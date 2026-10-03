package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type GroupBoxVariant uint8

const (
	GroupBoxSurface GroupBoxVariant = iota // legacy surface and border
	GroupBoxNormal
	GroupBoxFill
	GroupBoxOutline
)

// GroupBoxView visually and semantically groups views under a title.
type GroupBoxView struct {
	title        string
	description  string
	children     []el.View
	variant      GroupBoxVariant
	footer       el.View
	titleStyle   func(*el.TextEl)
	contentStyle func(*el.DivEl)
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

// Variant changes the body surface; the title, description and footer stay outside.
func (v *GroupBoxView) Variant(variant GroupBoxVariant) *GroupBoxView {
	if variant <= GroupBoxOutline {
		v.variant = variant
	}
	return v
}

// Footer sets supporting content below the body; nil removes it.
func (v *GroupBoxView) Footer(view el.View) *GroupBoxView { v.footer = view; return v }

// TitleStyle refines the freshly built title each frame. Do not retain the element.
// Nil restores default styling.
func (v *GroupBoxView) TitleStyle(fn func(*el.TextEl)) *GroupBoxView { v.titleStyle = fn; return v }

// ContentStyle refines only the body, after the variant defaults. Do not retain it.
// Nil restores default styling; child view state is preserved.
func (v *GroupBoxView) ContentStyle(fn func(*el.DivEl)) *GroupBoxView { v.contentStyle = fn; return v }
func (v *GroupBoxView) Render(cx *el.Context) el.Element {
	box := el.Div().Role("group").Name(v.title).Gap(theme.SpaceMd)
	if v.title != "" {
		title := el.Text(v.title).Bold().TextColor(theme.Text)
		if v.titleStyle != nil {
			v.titleStyle(title)
		}
		box.Child(title.ID("title"))
	}
	if v.description != "" {
		box.Child(el.Text(v.description).ID("description").TextColor(theme.Muted))
	}
	content := el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).Rounded(theme.RadiusMd)
	switch v.variant {
	case GroupBoxSurface:
		content.Border(1, theme.Border).Bg(theme.Surface)
	case GroupBoxFill:
		content.Bg(theme.Subtle)
	case GroupBoxOutline:
		content.Border(1, theme.Border)
	}
	if v.contentStyle != nil {
		v.contentStyle(content)
	}
	content.ID("content")
	for _, child := range v.children {
		if child != nil {
			content.Child(child.Render(cx))
		}
	}
	box.Child(content)
	if v.footer != nil {
		box.Child(el.Div().ID("footer").TextSize(theme.TextMd).TextColor(theme.Muted).Child(v.footer.Render(cx)))
	}
	return box
}
