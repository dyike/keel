package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"time"
)

// SwitchDuration is the duration of a thumb position transition.
const SwitchDuration = 180 * time.Millisecond

type SwitchSize uint8

const (
	SwitchMedium SwitchSize = iota
	SwitchSmall
)

// SwitchView turns a setting on or off immediately; use Checkbox for choices
// that are applied later by a submit button.
type SwitchView struct {
	name            string // accessible name from a Form or Settings row when label is empty
	label           string
	value, disabled bool
	onChange        func(bool)
	size            SwitchSize
	color           *color.NRGBA
	labelLeft       bool
	motion          valueMotion
}

func Switch(label string, on bool) *SwitchView           { return &SwitchView{label: label, value: on} }
func (v *SwitchView) OnChange(fn func(bool)) *SwitchView { v.onChange = fn; return v }
func (v *SwitchView) Value() bool                        { return v.value }
func (v *SwitchView) SetValue(on bool)                   { v.value = on }
func (v *SwitchView) SetDisabled(on bool)                { v.disabled = on }

// Size selects a 36×20dp (medium) or 28×16dp (small) track.
func (v *SwitchView) Size(size SwitchSize) *SwitchView {
	if size <= SwitchSmall {
		v.size = size
	}
	return v
}

// Color overrides the checked track color. ClearColor restores the theme.
func (v *SwitchView) Color(c color.NRGBA) *SwitchView { v.color = &c; return v }
func (v *SwitchView) ClearColor() *SwitchView         { v.color = nil; return v }

// LabelSide accepts el.Left and el.Right; other sides are ignored.
func (v *SwitchView) LabelSide(side el.Side) *SwitchView {
	if side == el.Left || side == el.Right {
		v.labelLeft = side == el.Left
	}
	return v
}

func (v *SwitchView) Render(cx *el.Context) el.Element {
	track := theme.Border
	if v.value {
		track = theme.Primary
		if v.color != nil {
			track = *v.color
		}
	}
	if v.disabled {
		track = theme.Subtle
		if v.value && v.color != nil {
			track = *v.color
			track.A /= 2
		}
	}
	width, height, knobSize := float32(36), float32(20), float32(16)
	if v.size == SwitchSmall {
		width, height, knobSize = 28, 16, 12
	}
	target := float32(0)
	if v.value {
		target = 1
	}
	position := v.motion.sample(cx, target, SwitchDuration)
	knob := el.Div().NoShrink().Size(el.Dp(knobSize)).Rounded(theme.RadiusFull).Bg(theme.Surface)
	if v.disabled {
		knob.Bg(theme.Border)
	}
	spacer := el.Div().W(el.Dp(max(0, width-knobSize-2*theme.SpaceXxs) * position)).NoShrink()
	body := el.Div().ID("track").W(el.Dp(width)).H(el.Dp(height)).NoShrink().Rounded(theme.RadiusFull).Bg(track).Px(theme.SpaceXxs).Row().Items(el.Center).Child(spacer, knob)
	return checkLabelSide(autoID("switch", v), "switch", v.label, v.name, v.value, v.disabled, body, func() {
		v.value = !v.value
		if v.onChange != nil {
			v.onChange(v.value)
		}
	}, v.labelLeft)
}

func (v *SwitchView) setName(s string) { v.name = s }

func (v *SwitchView) FocusID() string { return autoID("switch", v) }
