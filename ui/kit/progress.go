package kit

import (
	"image/color"
	"math"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ProgressView is a labelled progress bar. SetValue shows a fraction;
// SetIndeterminate shows work of unknown length as a sliding segment (still,
// with reduced motion).
type ProgressView struct {
	label         string
	value         float32
	indeterminate bool
	height        float32
	radius        *float32
	color         *color.NRGBA
	trackStyle    func(*el.DivEl)
}

func Progress(label string) *ProgressView { return &ProgressView{label: label} }
func (v *ProgressView) Value() float32    { return v.value }

// SetValue sets the fraction done, clamped to 0..1, and leaves indeterminate mode.
func (v *ProgressView) SetValue(f float32) {
	if math.IsNaN(float64(f)) {
		f = 0
	}
	v.value, v.indeterminate = min(max(f, 0), 1), false
}
func (v *ProgressView) SetIndeterminate(on bool) { v.indeterminate = on }
func (v *ProgressView) SetLabel(s string)        { v.label = s }

// Height sets the track height in dp (1–128), ignoring invalid values.
func (v *ProgressView) Height(dp float32) *ProgressView {
	if dp >= 1 && dp <= 128 {
		v.height = dp
	}
	return v
}

// Color overrides the fill color and suppresses the theme's primary gradient.
func (v *ProgressView) Color(c color.NRGBA) *ProgressView { v.color = &c; return v }

// Rounded sets the track and segment radius; invalid values are ignored.
func (v *ProgressView) Rounded(dp float32) *ProgressView {
	if dp >= 0 && !math.IsInf(float64(dp), 0) {
		v.radius = &dp
	}
	return v
}

// TrackStyle refines the fresh track each frame, after defaults. Nil restores
// defaults. Do not retain the element. The segment still fills its inner height.
func (v *ProgressView) TrackStyle(fn func(*el.DivEl)) *ProgressView { v.trackStyle = fn; return v }
func (v *ProgressView) Render(cx *el.Context) el.Element {
	pct := strconv.Itoa(int(v.value*100+0.5)) + "%"
	value := pct
	height := v.height
	if height == 0 {
		height = 8
	}
	radius := float32(theme.RadiusSm)
	if v.radius != nil {
		radius = min(*v.radius, height/2)
	}
	track := el.Div().WFull().H(el.Dp(height)).Rounded(radius).Bg(theme.Subtle).Row()
	if v.trackStyle != nil {
		v.trackStyle(track)
	}
	segment := func(width float32) *el.DivEl {
		fill := el.Div().W(el.Frac(width)).H(el.Full).Rounded(radius).Bg(theme.Primary)
		if v.color != nil {
			fill.Bg(*v.color)
		} else {
			fill.BgGradient(theme.PrimaryGradient)
		}
		return fill
	}
	if v.indeterminate {
		value = "indeterminate"
		phase := float32(0.3)
		if !el.ReducedMotion() {
			phase = float32(cx.Now().UnixMilli()%1500) / 1500
			cx.Animating()
		}
		track.Child(el.Div().W(el.Frac(phase*0.7)).NoShrink(), segment(.3))
	} else {
		track.Child(segment(v.value))
	}
	head := el.Div().Row().TextSize(theme.TextMd).TextColor(theme.Muted).Child(el.Text(v.label).Grow())
	if !v.indeterminate {
		head.Child(el.Text(pct))
	}
	return el.Div().WFull().Role("progressbar").Name(v.label).Value(value).Gap(theme.SpaceSm).Items(el.Stretch).Child(head, track)
}
