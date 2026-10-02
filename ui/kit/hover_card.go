package kit

import (
	"time"

	"github.com/dyike/keel/ui/theme"

	"github.com/dyike/keel/ui/el"
)

const (
	HoverCardOpenDelay  = 700 * time.Millisecond
	HoverCardCloseDelay = 300 * time.Millisecond
)

// HoverCardView previews richer content when the pointer rests on a view.
// Moving from the view onto the card keeps it open, so the card may contain
// links and buttons. Esc or a click outside closes it.
type HoverCardView struct {
	target, content el.View
	width           float32
	open            bool
	muted, disabled bool
}

func HoverCard(target, content el.View) *HoverCardView {
	return &HoverCardView{target: target, content: content, width: 300}
}

// Width sets the card width in dp, 300 by default.
func (v *HoverCardView) Width(dp float32) *HoverCardView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}

func (v *HoverCardView) Render(cx *el.Context) el.Element {
	id := autoID("hovercard", v)
	card := id + "/card"
	hovered := cx.Hovered(id) || (v.open && cx.Hovered(card))
	focused := cx.FocusWithin(id) || (v.open && cx.FocusWithin(card))
	if !hovered && !focused {
		v.muted = false
	}
	if focused && !v.muted && !v.disabled {
		v.open = true
	}
	switch {
	case hovered && !v.open && !v.muted && !v.disabled:
		cx.AfterEnabled(id, hoverCardKey{id, true}, HoverCardOpenDelay, func() { v.open = true })
	case !hovered && !focused && v.open:
		cx.AfterEnabled(card, hoverCardKey{id, false}, HoverCardCloseDelay, func() { v.open = false })
	}
	if v.open {
		w, h := cx.ViewportSize()
		panel := floating(theme.ElevationMd).ID(card).Role("dialog").W(el.Dp(v.width)).MaxW(el.Dp(max(0, w-16))).MaxH(el.Dp(max(0, h-16))).ScrollY().ScrollX().P(theme.SpaceXl)
		cx.Overlay(id, el.Anchored(id, panel).OnDismiss(func() { v.open = false; v.muted = true }))
		if v.content != nil {
			panel.Child(v.content.Render(cx))
		}
	}
	return el.Div().Disabled(v.disabled).Child(anchor(id, cx, v.target))
}

type hoverCardKey struct {
	id   string
	open bool
}

func (v *HoverCardView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.open = false
		v.muted = false
	}
}
