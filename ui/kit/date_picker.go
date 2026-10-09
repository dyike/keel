package kit

import (
	"image/color"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// DatePickerView is a field that opens a calendar. Picking a date (or
// completing a range with Range) closes it; Esc or a click outside closes it
// without a change. WithTime keeps the popup open for immediate date/time
// editing. Dates are shown with locale's Date format.
type DatePickerView struct {
	name                    string // accessible name from a Form row when label is empty
	label, placeholder, err string
	cal                     *CalendarView
	open, disabled          bool
	months                  int
	dateFormat              string
	clearable               bool
	presets                 []DatePickerPreset
	height                  float32
	plain                   bool
	clock                   *TimeFieldView
	startTime, endTime      time.Duration
	selectedDay             time.Time
	revealed                time.Time
	onChange                func(start, end time.Time)
}

func DatePicker(label string) *DatePickerView {
	v := &DatePickerView{label: label, cal: Calendar(), months: 1}
	v.cal.OnChange(func(start, end time.Time) {
		if v.editsTime() {
			if start.Equal(v.selectedDay) {
				v.close()
				return
			}
			v.selectedDay = start
		} else {
			v.open = false
		}
		v.err = ""
		start, end = v.DateTimeValue()
		if v.onChange != nil {
			v.onChange(start, end)
		}
	})
	return v
}

// Size sets the field's minimum height in dp, scaling its text, icons and
// spacing. Recommended heights are 28, 36 and 48. Zero restores theme defaults.
// It does not change the calendar or preset sizes; invalid values are ignored.
func (v *DatePickerView) Size(dp float32) *DatePickerView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Appearance controls the field background, border, rounding and padding.
// The calendar popup keeps its own appearance; minimum height is preserved.
func (v *DatePickerView) Appearance(on bool) *DatePickerView { v.plain = !on; return v }

// Months requests consecutive month panels; narrow windows show fewer panels.
func (v *DatePickerView) Months(n int) *DatePickerView { v.months = max(1, min(12, n)); return v }

// FirstWeekday overrides the locale week layout for this picker's calendar.
// Invalid values are ignored; dates and range drafts are preserved.
func (v *DatePickerView) FirstWeekday(day time.Weekday) *DatePickerView {
	previous := v.cal.weekStart()
	v.cal.FirstWeekday(day)
	if previous != v.cal.weekStart() {
		v.revealed = time.Time{}
	}
	return v
}

// ResetFirstWeekday restores the locale week layout.
func (v *DatePickerView) ResetFirstWeekday() *DatePickerView {
	previous := v.cal.weekStart()
	v.cal.ResetFirstWeekday()
	if previous != v.cal.weekStart() {
		v.revealed = time.Time{}
	}
	return v
}

// Range picks a span of dates instead of one.
func (v *DatePickerView) Range() *DatePickerView               { v.cal.Range(); return v }
func (v *DatePickerView) Placeholder(s string) *DatePickerView { v.placeholder = s; return v }

// Format sets a Go time layout for both endpoints. Empty restores locale.Date.
// It changes display only, without changing selection or firing OnChange.
func (v *DatePickerView) Format(layout string) *DatePickerView { v.dateFormat = layout; return v }

// Clearable shows an independent clear button when a date is selected.
func (v *DatePickerView) Clearable(on bool) *DatePickerView { v.clearable = on; return v }

// Bounds limits selectable dates; a zero time leaves that side open.
func (v *DatePickerView) Bounds(min, max time.Time) *DatePickerView { v.cal.Bounds(min, max); return v }
func (v *DatePickerView) DisableDates(fn func(time.Time) bool) *DatePickerView {
	v.cal.DisableDates(fn)
	return v
}

// OnChange runs when the user picks a date or completes a range (start == end for one date).
// WithTime also reports committed clock edits. Clearing sends two zero times.
func (v *DatePickerView) OnChange(fn func(start, end time.Time)) *DatePickerView {
	v.onChange = fn
	return v
}
func (v *DatePickerView) Value() (start, end time.Time) {
	start, end = v.cal.Value()
	if v.editsTime() && !start.IsZero() {
		start = dateWithClock(start, v.clock.Value())
		end = start
	}
	return
}
func (v *DatePickerView) SetValue(start, end time.Time) {
	v.cal.SetValue(start, end)
	v.selectedDay = day(start)
	if v.editsTime() && !start.IsZero() {
		v.clock.SetValue(time.Duration(start.Hour())*time.Hour + time.Duration(start.Minute())*time.Minute + time.Duration(start.Second())*time.Second)
	}
}
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
	start, end := v.DateTimeValue()
	if start.IsZero() {
		return ""
	}
	f := locale.Current().Date
	if v.dateFormat != "" {
		f = func(t time.Time) string { return t.Format(v.dateFormat) }
	}
	if v.cal.rangeMode && !end.IsZero() && f(end) != f(start) {
		return f(start) + " – " + f(end)
	}
	if v.editsTime() && v.dateFormat == "" {
		layout := "15:04"
		if v.clock.seconds {
			layout += ":05"
		}
		if v.clock.uses12() {
			layout = "03:04"
			if v.clock.seconds {
				layout += ":05"
			}
			layout += " PM"
		}
		return f(start) + " " + start.Format(layout)
	}
	return f(start)
}

func (v *DatePickerView) Render(cx *el.Context) el.Element {
	id := v.FocusID()
	frameID := id + "/frame"
	height := float32(theme.ControlHeight)
	if v.height > 0 {
		height = v.height
	}
	ratio := height / float32(theme.ControlHeight)
	shown, fg := v.text(), theme.Text
	if shown == "" {
		shown, fg = v.placeholder, theme.Muted
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
	field := el.Div().ID(id).Row().Items(el.Center).Gap(theme.SpaceMd*ratio).Grow().MinW(el.Dp(0)).Disabled(v.disabled).Role("button").Name(v.a11y()).Value(v.text()).
		Focusable(true).FocusStyle(func(s *el.Style) {
		if !v.plain {
			s.BorderColor(color.NRGBA{})
		}
	}).OnClick(func() { setOpen(!v.open) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) != key.NameDownArrow {
				return false
			}
			if e.State == el.KeyPress {
				setOpen(true)
			}
			return true
		}).
		Child(el.Text(shown).TextColor(fg).Grow().MaxLines(1), Icon(IconCalendar).Size(16*ratio).Color(theme.Muted).Render(cx))
	if v.disabled {
		field.TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	frame := fieldFrame(frameID, cx.FocusWithin(frameID), v.err != "", v.disabled, false).MinH(el.Dp(height)).Px(10 * ratio).Py(theme.SpaceXs * ratio).Gap(theme.SpaceMd * ratio).Child(field)
	if v.height > 0 {
		frame.TextSize(float32(theme.BodySize) * ratio)
	}
	if v.plain {
		frame.Bg(color.NRGBA{}).Border(0, color.NRGBA{}).Rounded(0).P(0)
	}
	start, _ := v.Value()
	if v.clearable && !start.IsZero() {
		clear := Button("", func() {
			v.close()
			v.SetValue(time.Time{}, time.Time{})
			v.err = ""
			cx.Focus(id)
			if v.onChange != nil {
				v.onChange(time.Time{}, time.Time{})
			}
		}).ID(id + "/clear").Name(locale.Current().Name(locale.Current().Clear, v.a11y())).Icon(IconClose).Variant(ButtonGhost).Size(24 * ratio)
		clear.SetDisabled(v.disabled)
		frame.Child(clear.Render(cx))
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
			week := ((int(first.Weekday())-int(v.cal.weekStart())+7)%7 + focus.Day() - 1) / 7
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
		cx.Overlay(id, el.Anchored(frameID, floating(theme.ElevationMd).ID(viewportID).Role("dialog").Name(v.a11y()).
			MaxW(el.Dp(max(1, width-16))).MaxH(el.Dp(max(1, height-16))).
			ScrollY().ScrollX().P(theme.SpaceLg).Gap(theme.SpaceMd).Child(calendar, v.renderTime(cx), v.renderPresets(cx))).
			Modal().TrapFocus().OnDismiss(v.close))
	}
	return labelled(v.label, el.Div().Items(el.Start).MinW(el.Dp(200)).Child(frame), v.err)
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
	if v.clock != nil {
		v.clock.SetDisabled(true)
		v.clock.SetDisabled(false)
	}
	v.cal.choosing, v.cal.yearEditing = false, false
}
