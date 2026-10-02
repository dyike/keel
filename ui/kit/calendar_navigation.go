package kit

import (
	"strconv"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Months displays 1–12 consecutive months, wrapping when the viewport is narrow.
func (v *CalendarView) Months(count int) *CalendarView { v.months = max(1, min(12, count)); return v }

// RangePending reports whether the first endpoint is waiting for completion.
func (v *CalendarView) RangePending() bool { return v.pending }

// CancelRange discards the unfinished range without changing Value or callbacks.
func (v *CalendarView) CancelRange() { v.pending = false; v.draft = time.Time{}; v.rangeError = false }

// Limit predicate calls for unbounded calendars whose predicate rejects every
// date. Normal navigation skips disabled days up to one year per key press.
func (v *CalendarView) seek(d time.Time, direction int) (time.Time, bool) {
	if !v.lo.IsZero() && d.Before(v.lo) {
		d = v.lo
		direction = 1
	}
	if !v.hi.IsZero() && d.After(v.hi) {
		d = v.hi
		direction = -1
	}
	for range 366 {
		if !v.lo.IsZero() && d.Before(v.lo) || !v.hi.IsZero() && d.After(v.hi) {
			break
		}
		if v.allowed(d) {
			return d, true
		}
		d = d.AddDate(0, 0, direction)
	}
	return time.Time{}, false
}
func (v *CalendarView) monthAllowed(month time.Time) bool {
	return (v.lo.IsZero() || !month.AddDate(0, 1, -1).Before(v.lo)) && (v.hi.IsZero() || !month.After(v.hi))
}
func (v *CalendarView) moveFocus(cx *el.Context, d time.Time) {
	direction := 1
	if d.Before(v.focus) {
		direction = -1
	}
	next, ok := v.seek(d, direction)
	if !ok {
		return
	}
	v.focus = next
	last := v.month.AddDate(0, max(1, v.months), 0)
	if next.Before(v.month) || !next.Before(last) {
		v.month = monthOf(next)
	}
	cx.Focus(v.cellID(next))
}
func (v *CalendarView) yearMonths(cx *el.Context) el.Element {
	text := locale.Current()
	if v.yearPicker == nil {
		v.yearPicker = NumberInput(text.Year).Range(1, 9999).Decimals(0).OnChange(func(year float64) { v.chooseYear = int(year) })
	}
	v.yearPicker.label = text.Year
	v.yearPicker.SetDisabled(v.disabled)
	if !v.yearEditing {
		v.yearPicker.SetValue(float64(v.chooseYear))
		v.yearEditing = true
	}
	grid := el.Div().Grid(3).Gap(theme.SpaceXs).W(el.Dp(252))
	for i := 1; i <= 12; i++ {
		month := time.Date(v.chooseYear, time.Month(i), 1, 0, 0, 0, 0, v.month.Location())
		button := Button(text.MonthNames[i-1], func() { v.month = month; v.choosing = false; v.yearEditing = false; v.moveFocus(cx, month) }).Size(30)
		button.SetDisabled(v.disabled || !v.monthAllowed(month))
		grid.Child(button.Render(cx))
	}
	return el.Div().W(el.Dp(252)).Gap(theme.SpaceMd).Child(v.yearPicker.Render(cx), grid, Button(text.Cancel, func() { v.choosing = false; v.yearEditing = false; cx.Focus(v.FocusID()) }).Render(cx))
}
func (v *CalendarView) monthTitle() string {
	text := locale.Current()
	if v.choosing {
		return strconv.Itoa(v.chooseYear)
	}
	title := text.Month(v.month.Year(), v.month.Month())
	if v.months > 1 {
		last := v.month.AddDate(0, v.months-1, 0)
		title += " – " + text.Month(last.Year(), last.Month())
	}
	return title
}
