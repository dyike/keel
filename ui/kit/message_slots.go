package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Avatar replaces the default initials with arbitrary content. Nil hides the
// avatar. Explicit avatars also appear on user messages, on the trailing side.
func (v *MessageView) Avatar(view el.View) *MessageView {
	v.avatar, v.avatarSet = view, true
	return v
}

// DefaultAvatar restores initials for incoming messages and no avatar for User.
func (v *MessageView) DefaultAvatar() *MessageView {
	v.avatar, v.avatarSet = nil, false
	return v
}

// Header supplies metadata above the body. Nil removes this slot.
func (v *MessageView) Header(view el.View) *MessageView { v.header = view; return v }

// Footer supplies content below delivery state, actions and reactions. It may
// contain interactive controls. Nil removes this slot.
func (v *MessageView) Footer(view el.View) *MessageView { v.footer = view; return v }

// Content replaces the body independently of metadata and delivery state.
func (v *MessageView) Content(view el.View) *MessageView { v.content = view; return v }

func (v *MessageView) metadata(cx *el.Context, id string, view el.View) el.Element {
	row := el.Div().ID(id).Row().Wrap().Gap(theme.SpaceXs).TextSize(theme.TextXs).TextColor(theme.Muted)
	if v.user {
		row.Justify(el.End)
	}
	return row.Child(view.Render(cx))
}
