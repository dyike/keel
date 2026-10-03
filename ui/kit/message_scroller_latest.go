package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// LatestButton enables the built-in control, which is enabled by default.
func (v *MessageScrollerView) LatestButton(on bool) *MessageScrollerView {
	v.latestHidden = !on
	return v
}

// LatestLabel overrides the control's label and accessible name. Empty restores
// the current locale's label.
func (v *MessageScrollerView) LatestLabel(label string) *MessageScrollerView {
	v.latestLabel = label
	return v
}

// LatestRenderer refines a fresh default button each frame. Nil restores the
// default renderer; returning nil falls back to the default button. The scroller
// always owns its identity and scroll action, including for a replacement button.
func (v *MessageScrollerView) LatestRenderer(fn func(*ButtonView) *ButtonView) *MessageScrollerView {
	v.latestRenderer = fn
	return v
}

// LatestTransition sets the enter/leave fade duration. Zero disables animation;
// negative durations are ignored. Reduced motion always displays the final state.
func (v *MessageScrollerView) LatestTransition(d time.Duration) *MessageScrollerView {
	if d >= 0 {
		v.latestDuration = d
	}
	return v
}

func (v *MessageScrollerView) renderLatest(cx *el.Context, wrap *el.DivEl, scrolledUp bool) {
	visible := scrolledUp && !v.latestHidden
	target := float32(0)
	if visible {
		target = 1
	}
	if v.latestDuration == 0 {
		v.latestMotion = valueMotion{}
	}
	alpha := v.latestMotion.sample(cx, target, max(v.latestDuration, time.Nanosecond))
	if !visible && alpha == 0 {
		return
	}
	label := v.latestLabel
	if label == "" {
		label = locale.Current().Latest
	}
	button := Button(label, v.ScrollToEnd).Icon(IconChevronDown).Variant(ButtonSecondary).Size(28)
	if v.latestRenderer != nil {
		defaults := *button
		if custom := v.latestRenderer(&defaults); custom != nil {
			copy := *custom
			button = &copy
		}
	}
	button.id = v.list.ID() + "/latest"
	button.onClick = v.ScrollToEnd
	// A leaving control cannot steal focus or accept a stale click.
	wrap.Child(el.Div().Absolute().Right(20).Bottom(16).Opacity(alpha).Disabled(!visible).Child(button.Render(cx)))
}
