package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// GroupBoxView visually and semantically groups views under a title.
type GroupBoxView struct {
	title       string
	description string
	elements    []el.Element
	children    []el.View
}

func GroupBox(title string, children ...el.View) *GroupBoxView {
	v := &GroupBoxView{title: title}
	v.SetChildren(children...)
	return v
}
func (v *GroupBoxView) Description(s string) *GroupBoxView { v.description = s; return v }
func (v *GroupBoxView) Child(els ...el.Element) *GroupBoxView {
	v.elements = append(v.elements, els...)
	return v
}
func (v *GroupBoxView) SetTitle(s string) { v.title = s }
func (v *GroupBoxView) SetChildren(children ...el.View) {
	v.children = append([]el.View(nil), children...)
}
func (v *GroupBoxView) Render(cx *el.Context) el.Element {
	box := el.Div().Role("group").Name(v.title).Gap(8)
	if v.title != "" {
		box.Child(el.Text(v.title).Bold().TextColor(theme.Text))
	}
	if v.description != "" {
		box.Child(el.Text(v.description).TextColor(theme.Muted))
	}
	content := el.Div().P(16).Gap(12).Rounded(6).Border(1, theme.Border).Bg(theme.Surface)
	for _, e := range v.elements {
		if e != nil {
			content.Child(e)
		}
	}
	for _, child := range v.children {
		if child != nil {
			content.Child(child.Render(cx))
		}
	}
	return box.Child(content)
}
