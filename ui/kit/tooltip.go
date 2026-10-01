package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TooltipDelay is how long the pointer rests on a view before its tooltip shows.
const TooltipDelay = 500 * time.Millisecond

// TooltipView shows a short hint next to a view: after TooltipDelay of
// hovering, or at once when keyboard focus is inside the view. Moving away or
// pressing Esc hides it. The hint never takes focus or clicks.
type TooltipView struct {
	target el.View
	text   string
	shown  bool // the hover delay has elapsed
	muted  bool // dismissed with Esc until the pointer and focus leave
}

// WithTooltip wraps target with a hint.
func WithTooltip(target el.View, text string) *TooltipView {
	return &TooltipView{target: target, text: text}
}
func (v *TooltipView) SetText(s string) { v.text = s }

func (v *TooltipView) Render(cx *el.Context) el.Element {
	id := autoID("tooltip", v)
	hovered, focused := cx.Hovered(id), cx.FocusWithin(id)
	if !hovered {
		v.shown = false
	} else if !v.shown {
		cx.After(tooltipKey{id}, TooltipDelay, func() { v.shown = true })
	}
	if !hovered && !focused {
		v.muted = false
	}
	if (v.shown || focused) && !v.muted && v.text != "" {
		tip := el.Div().Role("tooltip").Name(v.text).Px(8).Py(4).Rounded(4).MaxW(el.Dp(280)).
			Bg(theme.Text).TextColor(theme.Surface).TextSize(12).Child(el.Text(v.text))
		cx.Overlay(id, el.Anchored(id, tip).Placement(el.Top, el.Center).OnDismiss(func() { v.muted = true }))
	}
	return anchor(id, cx, v.target)
}

type tooltipKey struct{ id string }
