package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// DisclosureDuration is the duration of an expand/collapse transition.
const DisclosureDuration = 180 * time.Millisecond

type disclosureMotion struct{ valueMotion }

func (m *disclosureMotion) progress(cx *el.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return m.sample(cx, target, DisclosureDuration)
}

type disclosureStyle struct{ px, py, bottom, font, icon, gap float32 }

func defaultDisclosureStyle() disclosureStyle {
	return disclosureStyle{theme.SpaceMd + theme.SpaceSm, theme.SpaceLg, theme.SpaceMd + theme.SpaceSm, 0, 16, theme.SpaceMd}
}

func disclosureTrigger(cx *el.Context, id, label string, heading el.View, open, disabled bool, toggle func(), style disclosureStyle) *el.DivEl {
	state, icon := "collapsed", IconChevronRight
	if open {
		state = "expanded"
		icon = IconChevronDown
	}
	head := el.Div().ID(id).Role("disclosure").Name(label).Value(state).Disabled(disabled).Row().Items(el.Center).Gap(style.gap).Px(style.px).Py(style.py).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).OnClick(toggle)
	if style.font > 0 {
		head.TextSize(style.font)
	}
	if heading == nil {
		head.Child(el.Text(label).Bold().Grow())
	} else {
		head.Child(el.Div().Grow().MinW(el.Dp(0)).Child(heading.Render(cx)))
	}
	head.Child(Icon(icon).Size(style.icon).Color(theme.Muted).Render(cx))
	if !disabled {
		head.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	return head
}
func disclosureContent(cx *el.Context, id, trigger string, body el.View, open, disabled bool, motion *disclosureMotion, style disclosureStyle) el.Element {
	p := motion.progress(cx, open)
	if !open && cx.FocusWithin(id) {
		cx.Focus(trigger)
	}
	box := el.Div().ID(id).Items(el.Stretch).Reveal(p).Disabled(disabled || !open).Hidden(p == 0 && !open)
	if body != nil {
		content := el.Div().Px(style.px).Pb(style.bottom).Items(el.Stretch)
		if style.font > 0 {
			content.TextSize(style.font)
		}
		box.Child(content.Child(body.Render(cx)))
	}
	return box
}
