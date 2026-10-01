package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// SwitchView turns a setting on or off immediately; use Checkbox for choices
// that are applied later by a submit button.
type SwitchView struct {
	name            string // accessible name from a Form or Settings row when label is empty
	label           string
	value, disabled bool
	onChange        func(bool)
}

func Switch(label string, on bool) *SwitchView           { return &SwitchView{label: label, value: on} }
func (v *SwitchView) OnChange(fn func(bool)) *SwitchView { v.onChange = fn; return v }
func (v *SwitchView) Value() bool                        { return v.value }
func (v *SwitchView) SetValue(on bool)                   { v.value = on }
func (v *SwitchView) SetDisabled(on bool)                { v.disabled = on }

func (v *SwitchView) Render(cx *el.Context) el.Element {
	track := theme.Border
	if v.value {
		track = theme.Primary
	}
	if v.disabled {
		track = theme.Subtle
	}
	knob := el.Div().Size(el.Dp(16)).Rounded(8).Bg(theme.Surface)
	if v.disabled {
		knob.Bg(theme.Border)
	}
	body := el.Div().W(el.Dp(36)).H(el.Dp(20)).NoShrink().Rounded(10).Bg(track).Px(2).Row().Items(el.Center).Child(knob)
	if v.value {
		body.Justify(el.End)
	}
	return check(autoID("switch", v), "switch", v.label, v.name, v.value, v.disabled, body, func() {
		v.value = !v.value
		if v.onChange != nil {
			v.onChange(v.value)
		}
	})
}

func (v *SwitchView) setName(s string) { v.name = s }
