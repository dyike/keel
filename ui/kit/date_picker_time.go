package kit

import (
	"github.com/dyike/keel/ui/el"
	"time"
)

// WithTime enables minute precision time editing for a single date. Range
// pickers edit dates only; DateTimeValue retains both endpoint clocks.
// Changes commit immediately; selecting the selected date again closes the
// popup. Single-date Value and all OnChange callbacks include time.
func (v *DatePickerView) WithTime() *DatePickerView {
	if v.clock == nil {
		v.clock = TimeField("").Segmented()
		v.clock.SetValue(v.startTime)
		v.endTime = v.endTime.Truncate(time.Minute)
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
	start, end := v.startTime, v.endTime
	fresh := v.clock == nil
	v.WithTime().clock.Seconds()
	if fresh {
		v.clock.SetValue(start)
		v.endTime = end.Truncate(time.Second)
	}
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
	v.endTime = v.clock.Value()
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

// DateValue returns only the calendar dates, including in time editing mode.
func (v *DatePickerView) DateValue() (start, end time.Time) { return v.cal.Value() }

// SetDateValue changes the dates while preserving both clocks, without a callback.
func (v *DatePickerView) SetDateValue(start, end time.Time) {
	v.cal.SetValue(start, end)
	v.selectedDay, _ = v.cal.Value()
}

// DateTimeValue combines calendar dates with their stored clocks. Unlike Value,
// it includes both endpoint clocks in range mode. Zero dates remain zero.
func (v *DatePickerView) DateTimeValue() (start, end time.Time) {
	start, end = v.cal.Value()
	clock := v.startTime
	if v.clock != nil {
		clock = v.clock.Value()
	}
	start = dateWithClock(start, clock)
	if v.cal.rangeMode {
		end = dateWithClock(end, v.endTime)
	} else {
		end = start
	}
	return
}

// SetDateTimeValue stores dates and endpoint clocks without a callback or
// enabling time editing. Missing End uses Start; reversed ranges are sorted
// with their clocks. Configured time precision applies to both endpoints.
func (v *DatePickerView) SetDateTimeValue(start, end time.Time) {
	if !v.cal.rangeMode || end.IsZero() {
		end = start
	}
	if !start.IsZero() && !end.IsZero() && (day(end).Before(day(start)) || (day(end).Equal(day(start)) && end.Before(start))) && v.cal.rangeMode {
		start, end = end, start
	}
	v.SetDateValue(start, end)
	if start.IsZero() {
		return
	}
	v.startTime = clockOfDate(start)
	v.endTime = clockOfDate(end)
	if v.clock != nil {
		v.clock.SetValue(v.startTime)
		precision := time.Minute
		if v.clock.seconds {
			precision = time.Second
		}
		v.endTime = v.endTime.Truncate(precision)
	}
}

func clockOfDate(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond())
}
func dateWithClock(date time.Time, clock time.Duration) time.Time {
	if date.IsZero() {
		return date
	}
	return time.Date(date.Year(), date.Month(), date.Day(), int(clock/time.Hour), int(clock/time.Minute)%60, int(clock/time.Second)%60, int(clock%time.Second), date.Location())
}
