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
	target   el.View
	text     string
	shown    bool // the hover delay has elapsed
	disabled bool
	muted    bool // dismissed with Esc until the pointer and focus leave
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
		cx.AfterEnabled(id, tooltipKey{id}, TooltipDelay, func() { v.shown = true })
	}
	if !hovered && !focused {
		v.muted = false
	}
	if (v.shown || focused) && !v.muted && !v.disabled && v.text != "" {
		w, _ := cx.ViewportSize()
		tip := el.Div().Role("tooltip").Name(v.text).Px(theme.SpaceMd).Py(theme.SpaceXs).Rounded(theme.RadiusMd).Shadow(theme.ElevationSm).MaxW(el.Dp(min(280, max(0, w-16)))).
			Bg(theme.Text).TextColor(theme.Surface).TextSize(theme.TextSm).Child(el.Text(v.text))
		cx.Overlay(id, el.Anchored(id, tip).Placement(el.Top, el.Center).OnDismiss(func() { v.muted = true }))
	}
	return el.Div().Disabled(v.disabled).Child(anchor(id, cx, v.target))
}

type tooltipKey struct{ id string }

func (v *TooltipView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.shown = false
		v.muted = false
	}
}
