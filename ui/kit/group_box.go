package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// GroupBoxView visually and semantically groups views under a title.
type GroupBoxView struct {
	title    string
	children []el.View
}

func GroupBox(title string, children ...el.View) *GroupBoxView {
	v := &GroupBoxView{title: title}
	v.SetChildren(children...)
	return v
}
func (v *GroupBoxView) SetTitle(s string) { v.title = s }
func (v *GroupBoxView) SetChildren(children ...el.View) {
	v.children = append([]el.View(nil), children...)
}
func (v *GroupBoxView) Render(cx *el.Context) el.Element {
	box := el.Div().Role("group").Name(v.title).P(16).Gap(12).Rounded(6).Border(1, theme.Border).Bg(theme.Surface)
	if v.title != "" {
		box.Child(el.Text(v.title).Bold().TextColor(theme.Text))
	}
	for _, child := range v.children {
		if child != nil {
			box.Child(child.Render(cx))
		}
	}
	return box
}
