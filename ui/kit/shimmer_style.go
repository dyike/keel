package kit

import "time"

// ShimmerStyle is reusable motion configuration for text and attachment titles.
// The zero value uses a two-second forward loop with a 0.3 half-width.
type ShimmerStyle struct {
	Duration time.Duration
	Spread   float32
	Reverse  bool
	Once     bool
}

func (s ShimmerStyle) normalized() ShimmerStyle {
	if s.Duration <= 0 {
		s.Duration = 2 * time.Second
	}
	if !(s.Spread > 0 && s.Spread <= 1) {
		s.Spread = .3
	}
	return s
}

// Style replaces all motion settings, leaving text and typography unchanged.
// An effective setting change restarts the sweep. Reapplying the same settings
// each frame preserves progress. Invalid duration/spread values use defaults.
func (v *ShimmerTextView) Style(style ShimmerStyle) *ShimmerTextView {
	style = style.normalized()
	old := ShimmerStyle{Duration: v.duration, Spread: v.spread, Reverse: v.reverse, Once: v.once}.normalized()
	if style != old {
		v.Restart()
	}
	v.duration, v.spread, v.reverse, v.once = style.Duration, style.Spread, style.Reverse, style.Once
	return v
}

// TitleShimmer configures the default attachment title's loading animation.
// ShimmerStyle{} restores the defaults. PartStatus still controls when it runs.
func (v *AttachmentView) TitleShimmer(style ShimmerStyle) *AttachmentView {
	v.titleStyle = style.normalized()
	return v
}
