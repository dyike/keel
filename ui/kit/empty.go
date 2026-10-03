package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// EmptyView explains why a region contains no results.
type EmptyView struct {
	title, description string
	icon               IconName
	action             el.View
	media              el.View
}

func Empty(title string) *EmptyView                  { return &EmptyView{title: title, icon: IconInbox} }
func (v *EmptyView) Description(s string) *EmptyView { v.description = s; return v }
func (v *EmptyView) Icon(i IconName) *EmptyView      { v.icon = i; return v }
func (v *EmptyView) Action(e el.View) *EmptyView     { v.action = e; return v }

// Media replaces the icon with arbitrary content, preserving its size and semantics.
// Nil restores the configured icon; IconNone hides that fallback.
func (v *EmptyView) Media(view el.View) *EmptyView { v.media = view; return v }
func (v *EmptyView) SetTitle(s string)             { v.title = s }
func (v *EmptyView) SetDescription(s string)       { v.description = s }
func (v *EmptyView) Render(cx *el.Context) el.Element {
	box := el.Div().P(theme.Space2xl).Gap(theme.SpaceMd).Items(el.Center).Bg(theme.Surface)
	if v.media != nil {
		box.Child(el.Div().ID("media").Items(el.Center).MaxW(el.Full).Child(v.media.Render(cx)))
	} else if v.icon != IconNone {
		box.Child(el.Div().ID("media").Child(Icon(v.icon).Size(40).Color(theme.Muted).Render(cx)))
	}
	if v.title != "" {
		box.Child(el.Text(v.title).ID("title").TextColor(theme.Text))
	}
	if v.description != "" {
		box.Child(el.Text(v.description).ID("description").TextColor(theme.Muted))
	}
	if v.action != nil {
		box.Child(el.Div().ID("action").Items(el.Center).MaxW(el.Full).Child(v.action.Render(cx)))
	}
	return box
}
