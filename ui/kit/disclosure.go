package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"time"
)

// DisclosureDuration is the duration of an expand/collapse transition.
const DisclosureDuration = 180 * time.Millisecond

type disclosureMotion struct {
	initialized     bool
	value, from, to float32
	started         time.Time
}

func (m *disclosureMotion) progress(cx *el.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	if !m.initialized || el.ReducedMotion() {
		m.initialized = true
		m.value = target
		m.from = target
		m.to = target
		m.started = cx.Now()
		return target
	}
	p := max(0, min(1, float32(cx.Now().Sub(m.started))/float32(DisclosureDuration)))
	p = p * p * (3 - 2*p)
	m.value = m.from + (m.to-m.from)*p
	if target != m.to {
		m.from = m.value
		m.to = target
		m.started = cx.Now()
	}
	if m.value != m.to {
		cx.Animating()
	}
	return m.value
}
func disclosureTrigger(cx *el.Context, id, label string, heading el.View, open, disabled bool, toggle func()) *el.DivEl {
	state, icon := "collapsed", IconChevronRight
	if open {
		state = "expanded"
		icon = IconChevronDown
	}
	head := el.Div().ID(id).Role("disclosure").Name(label).Value(state).Disabled(disabled).Row().Items(el.Center).Gap(8).Px(14).Py(12).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).OnClick(toggle)
	if heading == nil {
		head.Child(el.Text(label).Bold().Grow())
	} else {
		head.Child(el.Div().Grow().MinW(el.Dp(0)).Child(heading.Render(cx)))
	}
	head.Child(Icon(icon).Size(16).Color(theme.Muted).Render(cx))
	if !disabled {
		head.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	return head
}
func disclosureContent(cx *el.Context, id, trigger string, body el.View, open, disabled bool, motion *disclosureMotion) el.Element {
	p := motion.progress(cx, open)
	if !open && cx.FocusWithin(id) {
		cx.Focus(trigger)
	}
	box := el.Div().ID(id).Items(el.Stretch).Reveal(p).Disabled(disabled || !open).Hidden(p == 0 && !open)
	if body != nil {
		box.Child(el.Div().Px(14).Pb(14).Items(el.Stretch).Child(body.Render(cx)))
	}
	return box
}
