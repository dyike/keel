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
	onChange                func(start, end time.Time)
}

func DatePicker(label string) *DatePickerView {
	v := &DatePickerView{label: label, cal: Calendar()}
	v.cal.OnChange(func(start, end time.Time) {
		v.open, v.err = false, ""
		if v.onChange != nil {
			v.onChange(start, end)
		}
	})
	return v
}

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
	if on {
		v.open = false
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
	border := theme.Border
	if v.err != "" {
		border = theme.Danger
	}
	setOpen := func(open bool) {
		v.open = open
		if open {
			// Focus the chosen day, or today, before the calendar first renders.
			v.cal.pending = false
			if s, _ := v.cal.Value(); !s.IsZero() {
				v.cal.focus = s
			} else {
				v.cal.focus = day(cx.Now())
			}
			v.cal.month = monthOf(v.cal.focus)
			cx.Focus(v.cal.FocusID())
		}
	}
	field := el.Div().ID(id).WFull().Role("button").Name(v.a11y()).Value(v.text()).Disabled(v.disabled).
		Row().Items(el.Center).Gap(8).H(el.Dp(36)).Px(10).Rounded(6).Bg(theme.Surface).Border(1, border).
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
		field.Bg(theme.Subtle).TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	if v.open {
		cx.Overlay(id, el.Anchored(id, surface().Role("dialog").Name(v.a11y()).P(12).Child(v.cal.Render(cx))).
			Modal().TrapFocus().OnDismiss(func() { v.open = false }))
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
