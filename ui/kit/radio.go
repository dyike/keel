package kit

import (
	"github.com/dyike/keel/ui/el"
)

// RadioView is one radio button on its own, for layouts a RadioGroup does
// not fit, such as options spread through a form or a table. Clicking or
// Space checks it; it never unchecks itself, so the app clears the others
// in OnChange. For a plain list of options, RadioGroup handles that and the
// arrow keys.
type RadioView struct {
	label          string
	checked        bool
	disabled       bool
	size, textSize float32
	content        el.View
	onChange       func(bool)
	id             string
}

func Radio(label string) *RadioView { return &RadioView{label: label} }

// OnChange runs when the user checks the radio; programmatic SetValue is silent.
func (v *RadioView) OnChange(fn func(checked bool)) *RadioView { v.onChange = fn; return v }
func (v *RadioView) Value() bool                               { return v.checked }
func (v *RadioView) SetValue(checked bool)                     { v.checked = checked }
func (v *RadioView) SetDisabled(on bool)                       { v.disabled = on }

// Checked sets the initial state.
func (v *RadioView) Checked(on bool) *RadioView { v.checked = on; return v }

// Size sets the ring's diameter in dp, 18 by default; TextSize the label's.
func (v *RadioView) Size(dp float32) *RadioView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}
func (v *RadioView) TextSize(sp float32) *RadioView {
	if sp > 0 && finiteNumber(float64(sp)) {
		v.textSize = sp
	}
	return v
}

// Content replaces the text label with any view, such as a title and a
// description on two lines. The constructor label stays the accessible name.
func (v *RadioView) Content(view el.View) *RadioView { v.content = view; return v }

// ID names the radio for cx.Focus and anchored layers.
func (v *RadioView) ID(id string) *RadioView { v.id = id; return v }

// FocusID is the element that takes keyboard focus.
func (v *RadioView) FocusID() string {
	if v.id != "" {
		return v.id
	}
	return autoID("radio", v)
}

func (v *RadioView) Render(cx *el.Context) el.Element {
	label := v.label
	if v.content != nil {
		label = ""
	}
	row := check(v.FocusID(), "radio", label, v.label, v.checked, v.disabled, radioDot(v.size, v.checked, v.disabled), func() {
		if v.disabled || v.checked {
			return
		}
		v.checked = true
		if v.onChange != nil {
			v.onChange(true)
		}
	})
	if v.content != nil {
		row.Child(el.Div().ID("label").Child(v.content.Render(cx)))
	}
	if v.textSize > 0 {
		row.TextSize(v.textSize)
	}
	return row
}
