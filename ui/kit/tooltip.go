package kit

import (
	"runtime"
	"time"

	"github.com/dyike/keel/ui/core"
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
	content  el.View
	action   string
	side     el.Side
	align    el.Align
	offset   float32
	shown    bool // the hover delay has elapsed
	disabled bool
	muted    bool // dismissed with Esc until the pointer and focus leave
}

// WithTooltip wraps target with a hint.
func WithTooltip(target el.View, text string) *TooltipView {
	return &TooltipView{target: target, text: text, side: el.Top, align: el.Center, offset: 4}
}
func (v *TooltipView) SetText(s string) { v.text = s }

// Content replaces the displayed text with a rich, noninteractive view. The
// text supplied to WithTooltip/SetText remains the accessible name. Nil restores text.
func (v *TooltipView) Content(content el.View) *TooltipView { v.content = content; return v }

// Action displays the first current key binding for an action; it does not
// register or execute the action. Unbound actions show no key label.
func (v *TooltipView) Action(action string) *TooltipView { v.action = action; return v }
func (v *TooltipView) Placement(side el.Side, align el.Align) *TooltipView {
	v.side, v.align = side, align
	return v
}
func (v *TooltipView) Offset(dp float32) *TooltipView {
	if finiteNumber(float64(dp)) {
		v.offset = dp
	}
	return v
}

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
			Bg(theme.Text).TextColor(theme.Surface).TextSize(theme.TextSm).Gap(theme.SpaceSm).
			Disabled(true).DisabledStyle(func(s *el.Style) { s.TextColor(theme.Surface) })
		cx.Overlay(id, el.Anchored(id, tip).Placement(v.side, v.align).Offset(v.offset).OnDismiss(func() { v.muted = true }))
		if v.content != nil {
			tip.Child(v.content.Render(cx))
		} else {
			tip.Child(el.Text(v.text))
		}
		if bindings := core.Bindings(v.action); v.action != "" && len(bindings) > 0 {
			tip.Child(el.Text(core.ShortcutLabel(bindings[0], runtime.GOOS)).Name(bindings[0]).TextColor(theme.Surface))
		}
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
