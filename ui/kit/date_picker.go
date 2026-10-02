package kit

import (
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// DatePickerView is a field that opens a calendar. Picking a date (or
// completing a range with Range) closes it; Esc or a click outside closes it
// without a change. Dates are shown with locale's Date format.
type DatePickerView struct {
	name                    string // accessible name from a Form row when label is empty
	label, placeholder, err string
	cal                     *CalendarView
	open, disabled          bool
	months                  int
	revealed                time.Time
	onChange                func(start, end time.Time)
}

func DatePicker(label string) *DatePickerView {
	v := &DatePickerView{label: label, cal: Calendar(), months: 1}
	v.cal.OnChange(func(start, end time.Time) {
		v.open, v.err = false, ""
		if v.onChange != nil {
			v.onChange(start, end)
		}
	})
	return v
}

// Months requests consecutive month panels; narrow windows show fewer panels.
func (v *DatePickerView) Months(n int) *DatePickerView { v.months = max(1, min(12, n)); return v }

// Range picks a span of dates instead of one.
func (v *DatePickerView) Range() *DatePickerView               { v.cal.Range(); return v }
func (v *DatePickerView) Placeholder(s string) *DatePickerView { v.placeholder = s; return v }

// Bounds limits selectable dates; a zero time leaves that side open.
func (v *DatePickerView) Bounds(min, max time.Time) *DatePickerView { v.cal.Bounds(min, max); return v }
func (v *DatePickerView) DisableDates(fn func(time.Time) bool) *DatePickerView {
	v.cal.DisableDates(fn)
	return v
}

// OnChange runs when the user picks a date or completes a range (start == end for one date).
func (v *DatePickerView) OnChange(fn func(start, end time.Time)) *DatePickerView {
	v.onChange = fn
	return v
}
func (v *DatePickerView) Value() (start, end time.Time) { return v.cal.Value() }
func (v *DatePickerView) SetValue(start, end time.Time) { v.cal.SetValue(start, end) }
func (v *DatePickerView) SetDisabled(on bool) {
	v.disabled = on
	v.cal.SetDisabled(on)
	if on {
		v.close()
	}
}
func (v *DatePickerView) SetError(msg string) { v.err = msg }
func (v *DatePickerView) Error() string       { return v.err }
func (v *DatePickerView) FocusID() string     { return autoID("datepicker", v) }

func (v *DatePickerView) text() string {
	start, end := v.cal.Value()
	if start.IsZero() {
		return ""
	}
	f := locale.Current().Date
	if v.cal.rangeMode && !end.IsZero() && !end.Equal(start) {
		return f(start) + " – " + f(end)
	}
	return f(start)
}

func (v *DatePickerView) Render(cx *el.Context) el.Element {
	id := v.FocusID()
	shown, color := v.text(), theme.Text
	if shown == "" {
		shown, color = v.placeholder, theme.Muted
	}
	setOpen := func(open bool) {
		v.close()
		if !open || v.disabled {
			return
		}
		v.open = true
		focus, _ := v.cal.Value()
		if focus.IsZero() {
			focus = day(cx.Now())
		}
		if next, ok := v.cal.seek(focus, 1); ok {
			focus = next
		}
		v.cal.focus, v.cal.month = focus, monthOf(focus)
		cx.Focus(v.cal.FocusID())
	}
	field := fieldFrame(id, false, v.err != "", v.disabled, false).Role("button").Name(v.a11y()).Value(v.text()).
		Focusable(true).OnClick(func() { setOpen(!v.open) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) != key.NameDownArrow {
				return false
			}
			if e.State == el.KeyPress {
				setOpen(true)
			}
			return true
		}).
		Child(el.Text(shown).TextColor(color).Grow().MaxLines(1), Icon(IconCalendar).Size(16).Color(theme.Muted).Render(cx))
	if v.disabled {
		field.TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	if v.open {
		width, height := cx.ViewportSize()
		// Keep panels on one row. A scroller retains access to the calendar
		// and range controls when the window cannot fit their full height.
		v.cal.Months(min(v.months, max(1, int((width-24+16)/268))))
		calendar := v.cal.Render(cx)
		viewportID := id + "/viewport"
		if !v.cal.choosing && !v.revealed.Equal(v.cal.focus) {
			focus := v.cal.focus
			first := monthOf(focus)
			week := ((int(first.Weekday())-int(locale.Current().FirstWeekday)+7)%7 + focus.Day() - 1) / 7
			top := float32(12 + 28 + 8 + 24 + week*32)
			if v.cal.months > 1 {
				top += 28
			}
			cx.ScrollIntoView(viewportID, top, top+32)
			// Retry after the first paint creates the scroll viewport, then
			// focus the now-visible cell rather than a clipped-out target.
			cx.AfterEnabled(viewportID, v, 0, func() {
				if !v.open || !v.cal.focus.Equal(focus) {
					return
				}
				cx.ScrollIntoView(viewportID, top, top+32)
				cx.Focus(v.cal.FocusID())
				v.revealed = focus
			})
		}
		cx.Overlay(id, el.Anchored(id, floating(theme.ElevationMd).ID(viewportID).Role("dialog").Name(v.a11y()).
			MaxW(el.Dp(max(1, width-16))).MaxH(el.Dp(max(1, height-16))).
			ScrollY().ScrollX().P(12).Child(calendar)).
			Modal().TrapFocus().OnDismiss(v.close))
	}
	return labelled(v.label, el.Div().Items(el.Start).MinW(el.Dp(200)).Child(field), v.err)
}

func (v *DatePickerView) setName(s string) { v.name = s }
func (v *DatePickerView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}

// close also discards drafts when the anchor is disabled or removed.
func (v *DatePickerView) close() {
	v.open = false
	v.revealed = time.Time{}
	v.cal.CancelRange()
	v.cal.choosing, v.cal.yearEditing = false, false
}
