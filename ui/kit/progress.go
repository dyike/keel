package kit

import (
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
}

func Progress(label string) *ProgressView { return &ProgressView{label: label} }
func (v *ProgressView) Value() float32    { return v.value }

// SetValue sets the fraction done, clamped to 0..1, and leaves indeterminate mode.
func (v *ProgressView) SetValue(f float32)       { v.value, v.indeterminate = min(max(f, 0), 1), false }
func (v *ProgressView) SetIndeterminate(on bool) { v.indeterminate = on }
func (v *ProgressView) SetLabel(s string)        { v.label = s }

func (v *ProgressView) Render(cx *el.Context) el.Element {
	pct := strconv.Itoa(int(v.value*100+0.5)) + "%"
	value := pct
	track := el.Div().H(el.Dp(8)).Rounded(theme.RadiusSm).Bg(theme.Subtle).Row()
	if v.indeterminate {
		value = "indeterminate"
		phase := float32(0.3)
		if !el.ReducedMotion() {
			phase = float32(cx.Now().UnixMilli()%1500) / 1500
			cx.Animating()
		}
		track.Child(el.Div().W(el.Frac(phase*0.7)).NoShrink(), el.Div().W(el.Frac(0.3)).H(el.Dp(8)).Rounded(theme.RadiusSm).Bg(theme.Primary).BgGradient(theme.PrimaryGradient))
	} else {
		track.Child(el.Div().W(el.Frac(v.value)).H(el.Dp(8)).Rounded(theme.RadiusSm).Bg(theme.Primary).BgGradient(theme.PrimaryGradient))
	}
	head := el.Div().Row().TextSize(theme.TextMd).TextColor(theme.Muted).Child(el.Text(v.label).Grow())
	if !v.indeterminate {
		head.Child(el.Text(pct))
	}
	return el.Div().Role("progressbar").Name(v.label).Value(value).Gap(theme.SpaceSm).Items(el.Stretch).Child(head, track)
}
