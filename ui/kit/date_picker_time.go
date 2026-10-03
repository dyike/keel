package kit

import (
	"github.com/dyike/keel/ui/el"
	"time"
)

// WithTime enables minute precision time editing for a single date. Range
// pickers remain date-only. Date and time changes commit immediately; selecting
// the selected date again closes the popup. Value and OnChange include time.
func (v *DatePickerView) WithTime() *DatePickerView {
	if v.clock == nil {
		v.clock = TimeField("").Segmented()
		v.selectedDay, _ = v.cal.Value()
		v.clock.OnChange(func(time.Duration) {
			if !v.open || v.disabled || !v.editsTime() {
				return
			}
			a, b := v.Value()
			if a.IsZero() {
				return
			}
			v.err = ""
			if v.onChange != nil {
				v.onChange(a, b)
			}
		})
	}
	return v
}

// TimeSeconds enables time editing with whole-second precision.
func (v *DatePickerView) TimeSeconds() *DatePickerView {
	v.WithTime().clock.Seconds()
	return v
}

// TimeHour12 overrides the locale clock format and enables time editing.
func (v *DatePickerView) TimeHour12(on bool) *DatePickerView {
	v.WithTime().clock.Hour12(on)
	return v
}

// DefaultTime sets the clock used for subsequent date selections. It wraps
// within one day and truncates to the configured precision, without a callback.
func (v *DatePickerView) DefaultTime(value time.Duration) *DatePickerView {
	v.WithTime().clock.SetValue(value)
	return v
}

func (v *DatePickerView) editsTime() bool { return v.clock != nil && !v.cal.rangeMode }
func (v *DatePickerView) renderTime(cx *el.Context) el.Element {
	if !v.editsTime() {
		return el.Div().Hidden(true)
	}
	v.clock.setName(v.a11y())
	return v.clock.Render(cx)
}
