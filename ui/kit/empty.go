package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// EmptyView explains why a region contains no results.
type EmptyView struct{ title, description string }

func Empty(title, description string) *EmptyView { return &EmptyView{title, description} }
func (v *EmptyView) SetTitle(s string)           { v.title = s }
func (v *EmptyView) SetDescription(s string)     { v.description = s }
func (v *EmptyView) Render(*el.Context) el.Element {
	box := el.Div().Role("empty").Name(v.title).P(24).Gap(8).Items(el.Center).Bg(theme.Surface)
	if v.title != "" {
		box.Child(el.Text(v.title).TextColor(theme.Text))
	}
	if v.description != "" {
		box.Child(el.Text(v.description).TextColor(theme.Muted))
	}
	return box
}
