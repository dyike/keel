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
func (v *MessageView) Content(view el.View) *MessageView {
	v.content, v.surface = view, nil
	return v
}

func (v *MessageView) metadata(cx *el.Context, id string, view el.View, override *bool) *el.DivEl {
	row := el.Div().ID(id).Row().Wrap().Gap(theme.SpaceXs).TextSize(theme.TextXs).TextColor(theme.Muted)
	inset := v.user || v.surface != nil
	if v.surface != nil && v.surface.variant == BubbleGhost {
		inset = false
	}
	if mixed, ok := v.content.(*MessageContentView); ok && mixed != nil {
		inset = !mixed.hasGhost()
	}
	if override != nil {
		inset = *override
	}
	if inset {
		row.Px(theme.SpaceLg)
	}
	if v.isEnd() {
		row.Justify(el.End)
	}
	return row.Child(view.Render(cx))
}

// Bubble installs a typed surface. Its alignment follows the message and its
// Ghost variant removes automatic header/footer insets. The source is not
// mutated; reuse it to update presentation. Nil clears the body.
func (v *MessageView) Bubble(surface *BubbleView) *MessageView {
	v.surface = surface
	v.content = nil
	if surface != nil {
		v.content = surface
	}
	return v
}

// HeaderInset overrides automatic surface padding for the header.
func (v *MessageView) HeaderInset(on bool) *MessageView { v.headerInset = &on; return v }

// FooterInset overrides automatic surface padding for the footer.
func (v *MessageView) FooterInset(on bool) *MessageView { v.footerInset = &on; return v }

// ResetContentInsets restores automatic header and footer padding.
func (v *MessageView) ResetContentInsets() *MessageView {
	v.headerInset, v.footerInset = nil, nil
	return v
}

// Alignment sets placement independently of User's default surface and avatar.
// Only Start and End are accepted; invalid values leave the current setting.
func (v *MessageView) Alignment(align el.Align) *MessageView {
	if align == el.Start || align == el.End {
		end := align == el.End
		v.alignEnd = &end
	}
	return v
}

// ResetAlignment restores Start for incoming messages and End for User.
func (v *MessageView) ResetAlignment() *MessageView { v.alignEnd = nil; return v }
func (v *MessageView) isEnd() bool {
	if v.alignEnd != nil {
		return *v.alignEnd
	}
	return v.user
}
