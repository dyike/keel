package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// CheckboxView is a labelled check box. Click, Space or Enter toggles it.
type CheckboxView struct {
	name                   string // accessible name from a Form or Settings row when label is empty
	label                  string
	value, mixed, disabled bool
	onChange               func(bool)
	size, textSize         float32
	tabStop                *bool
	tabIndex               *int
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

// Size sets the square size in dp (12–64); zero restores 18dp.
func (v *CheckboxView) Size(dp float32) *CheckboxView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.size = dp
		if dp > 0 {
			v.size = min(64, max(12, dp))
		}
	}
	return v
}

// TextSize sets label size in sp (8–128); zero restores inherited text size.
func (v *CheckboxView) TextSize(sp float32) *CheckboxView {
	if sp >= 0 && finiteNumber(float64(sp)) {
		v.textSize = sp
		if sp > 0 {
			v.textSize = min(128, max(8, sp))
		}
	}
	return v
}

func (v *CheckboxView) TabStop(on bool) *CheckboxView    { v.tabStop = &on; return v }
func (v *CheckboxView) TabIndex(index int) *CheckboxView { v.tabIndex = &index; return v }

func (v *CheckboxView) Render(cx *el.Context) el.Element {
	on := v.value || v.mixed
	size := v.size
	if size == 0 {
		size = 18
	}
	box := el.Div().Size(el.Dp(size)).NoShrink().Rounded(theme.RadiusSm).Center()
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
		box.Child(el.Div().W(el.Dp(size * 10 / 18)).H(el.Dp(max(1, size/9))).Rounded(theme.RadiusFull).Bg(mark))
	} else if v.value {
		box.Child(Icon(IconDone).Size(size * 14 / 18).Color(mark).Render(cx))
	}
	row := check(autoID("checkbox", v), "checkbox", v.label, v.name, v.value, v.disabled, box, func() {
		v.value, v.mixed = !v.value || v.mixed, false
		if v.onChange != nil {
			v.onChange(v.value)
		}
	})
	if v.mixed {
		row.Value("mixed")
	}
	if v.textSize > 0 {
		row.TextSize(v.textSize)
	}
	if v.tabStop != nil {
		row.TabStop(*v.tabStop)
	}
	if v.tabIndex != nil {
		row.TabIndex(*v.tabIndex)
	}
	return row
}

func (v *CheckboxView) setName(s string) { v.name = s }

func (v *CheckboxView) FocusID() string { return autoID("checkbox", v) }
