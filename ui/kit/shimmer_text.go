package kit

import (
	"image/color"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ShimmerTextView highlights readable text without replacing it with a skeleton.
type ShimmerTextView struct {
	text                    string
	duration                time.Duration
	spread                  float32
	reverse, once, disabled bool
	start                   time.Time
	size                    float32
	lines                   int
	color, highlight        *color.NRGBA
}

func ShimmerText(text string) *ShimmerTextView {
	return &ShimmerTextView{text: text, duration: 2 * time.Second, spread: .3}
}
func (v *ShimmerTextView) SetText(text string) { v.text = text }
func (v *ShimmerTextView) Duration(d time.Duration) *ShimmerTextView {
	if d > 0 {
		v.duration = d
	}
	return v
}
func (v *ShimmerTextView) Spread(f float32) *ShimmerTextView {
	if f > 0 && f <= 1 {
		v.spread = f
	}
	return v
}
func (v *ShimmerTextView) Reverse(on bool) *ShimmerTextView { v.reverse = on; return v }
func (v *ShimmerTextView) Once(on bool) *ShimmerTextView    { v.once = on; return v }
func (v *ShimmerTextView) Enabled(on bool) *ShimmerTextView {
	if v.disabled == on {
		v.start = time.Time{}
	}
	v.disabled = !on
	return v
}
func (v *ShimmerTextView) Size(sp float32) *ShimmerTextView {
	if sp >= 0 && finiteNumber(float64(sp)) {
		v.size = sp
	}
	return v
}
func (v *ShimmerTextView) MaxLines(n int) *ShimmerTextView {
	if n >= 0 {
		v.lines = n
	}
	return v
}
func (v *ShimmerTextView) Color(c color.NRGBA) *ShimmerTextView     { v.color = &c; return v }
func (v *ShimmerTextView) Highlight(c color.NRGBA) *ShimmerTextView { v.highlight = &c; return v }

// Restart begins a new sweep on the next render, including after Once completes.
func (v *ShimmerTextView) Restart() { v.start = time.Time{} }
func (v *ShimmerTextView) Render(cx *el.Context) el.Element {
	text := el.Text(v.text).MaxLines(v.lines)
	if v.size > 0 {
		text.TextSize(v.size)
	}
	if v.color != nil {
		text.TextColor(*v.color)
	}
	if v.disabled || el.ReducedMotion() || v.text == "" {
		return text
	}
	now := cx.Now()
	if v.start.IsZero() || now.Before(v.start) {
		v.start = now
	}
	duration := v.duration
	if duration <= 0 {
		duration = 2 * time.Second
	}
	spread := v.spread
	if spread <= 0 {
		spread = .3
	}
	elapsed := now.Sub(v.start)
	if v.once && elapsed >= duration {
		return text
	}
	phase := float32(elapsed%duration) / float32(duration)
	if v.reverse {
		phase = 1 - phase
	}
	highlight := theme.PrimaryText
	if v.highlight != nil {
		highlight = *v.highlight
	}
	cx.Animating()
	return text.Shimmer(phase, spread, highlight)
}
