package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// CheckboxView is a labelled check box. Click, Space or Enter toggles it.
type CheckboxView struct {
	label                  string
	value, mixed, disabled bool
	onChange               func(bool)
}

func Checkbox(label string, checked bool) *CheckboxView {
	return &CheckboxView{label: label, value: checked}
}
func (v *CheckboxView) OnChange(fn func(bool)) *CheckboxView { v.onChange = fn; return v }
func (v *CheckboxView) Value() bool                          { return v.value }

// SetValue sets the state without calling OnChange and clears mixed.
func (v *CheckboxView) SetValue(on bool) { v.value, v.mixed = on, false }

// SetMixed shows the indeterminate state, e.g. for "select all" when some
// rows are selected. The next click checks it.
func (v *CheckboxView) SetMixed(on bool)    { v.mixed = on }
func (v *CheckboxView) SetDisabled(on bool) { v.disabled = on }
func (v *CheckboxView) SetLabel(s string)   { v.label = s }

func (v *CheckboxView) Render(cx *el.Context) el.Element {
	on := v.value || v.mixed
	box := el.Div().Size(el.Dp(18)).NoShrink().Rounded(4).Center()
	switch {
	case v.disabled:
		box.Bg(theme.Subtle).Border(1, theme.Border)
	case on:
		box.Bg(theme.Primary)
	default:
		box.Bg(theme.Surface).Border(1, theme.Muted)
	}
	mark := theme.OnColor
	if v.disabled {
		mark = theme.Muted
	}
	if v.mixed {
		box.Child(el.Div().W(el.Dp(10)).H(el.Dp(2)).Rounded(1).Bg(mark))
	} else if v.value {
		box.Child(Icon(IconDone).Size(14).Color(mark).Render(cx))
	}
	row := check(autoID("checkbox", v), "checkbox", v.label, v.value, v.disabled, box, func() {
		v.value, v.mixed = !v.value || v.mixed, false
		if v.onChange != nil {
			v.onChange(v.value)
		}
	})
	if v.mixed {
		row.Value("mixed")
	}
	return row
}
