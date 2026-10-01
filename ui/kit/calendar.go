package kit

import (
	"image/color"
	"strconv"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CalendarView shows one month and picks a date, or a range with Range.
// Tab reaches the focused day; arrows move by a day or week, PageUp and
// PageDown by a month, Home and End to the week's ends, Enter or Space picks.
// Week layout, day and month names come from ui/locale.
type CalendarView struct {
	start, end   time.Time // chosen dates at midnight; end == start unless Range
	rangeMode    bool
	pending      bool // Range: the first end is picked, the second is not
	month, focus time.Time
	lo, hi       time.Time
	blocked      func(time.Time) bool
	disabled     bool
	onChange     func(start, end time.Time)
}

func Calendar() *CalendarView { return &CalendarView{} }

// Range picks a span: the first click sets one end, the second the other.
func (v *CalendarView) Range() *CalendarView { v.rangeMode = true; return v }

// Bounds limits selectable dates to [min, max]; a zero time leaves that side open.
func (v *CalendarView) Bounds(min, max time.Time) *CalendarView {
	v.lo, v.hi = day(min), day(max)
	return v
}

// DisableDates blocks dates for which fn returns true, e.g. weekends.
func (v *CalendarView) DisableDates(fn func(time.Time) bool) *CalendarView { v.blocked = fn; return v }

// OnChange runs when the user picks a date or completes a range; for a
// single date start == end.
func (v *CalendarView) OnChange(fn func(start, end time.Time)) *CalendarView {
	v.onChange = fn
	return v
}

// Value returns the chosen date (start == end) or range; zero when none.
func (v *CalendarView) Value() (start, end time.Time) { return v.start, v.end }

// SetValue chooses dates without calling OnChange and shows start's month.
// A single-date calendar uses only start.
func (v *CalendarView) SetValue(start, end time.Time) {
	v.start, v.end, v.pending = day(start), day(end), false
	if !v.rangeMode || v.end.IsZero() {
		v.end = v.start
	}
	if v.end.Before(v.start) {
		v.start, v.end = v.end, v.start
	}
	if !v.start.IsZero() {
		v.month, v.focus = monthOf(v.start), v.start
	}
}
func (v *CalendarView) SetDisabled(on bool) { v.disabled = on }

// SetMonth shows the month containing t.
func (v *CalendarView) SetMonth(t time.Time) { v.month = monthOf(t) }

func day(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
func monthOf(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func (v *CalendarView) allowed(d time.Time) bool {
	return !v.disabled && (v.lo.IsZero() || !d.Before(v.lo)) && (v.hi.IsZero() || !d.After(v.hi)) &&
		(v.blocked == nil || !v.blocked(d))
}

func (v *CalendarView) pick(d time.Time) {
	if !v.allowed(d) {
		return
	}
	v.focus = d
	switch {
	case !v.rangeMode:
		v.start, v.end = d, d
	case !v.pending:
		v.start, v.end, v.pending = d, time.Time{}, true
		return // wait for the other end
	default:
		v.end, v.pending = d, false
		if v.end.Before(v.start) {
			v.start, v.end = v.end, v.start
		}
	}
	if v.onChange != nil {
		v.onChange(v.start, v.end)
	}
}

func (v *CalendarView) cellID(d time.Time) string {
	return autoID("calendar", v) + "/" + d.Format("20060102")
}

// FocusID is the element ID of the focused day, for cx.Focus.
func (v *CalendarView) FocusID() string { return v.cellID(v.focus) }

func (v *CalendarView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	today := day(cx.Now())
	if v.month.IsZero() {
		v.month = monthOf(today)
	}
	if v.focus.IsZero() || monthOf(v.focus) != v.month {
		v.focus = v.month
		if monthOf(today) == v.month {
			v.focus = today
		}
		if !v.start.IsZero() && monthOf(v.start) == v.month {
			v.focus = v.start
		}
	}
	move := func(d time.Time) {
		v.focus, v.month = d, monthOf(d)
		cx.Focus(v.cellID(d))
	}
	prev := Button("", func() { v.month = v.month.AddDate(0, -1, 0) }).Name(text.PrevMonth).Icon(IconChevronLeft).Variant(ButtonGhost).Size(28)
	next := Button("", func() { v.month = v.month.AddDate(0, 1, 0) }).Name(text.NextMonth).Icon(IconChevronRight).Variant(ButtonGhost).Size(28)
	prev.SetDisabled(v.disabled)
	next.SetDisabled(v.disabled)
	title := text.Month(v.month.Year(), v.month.Month())
	head := el.Div().Row().Items(el.Center).Child(prev.Render(cx), el.Div().Grow().Items(el.Center).Child(el.Text(title).Bold().MaxLines(1)), next.Render(cx))

	const cw, ch = 36, 32
	week := el.Div().Row()
	for i := 0; i < 7; i++ {
		wd := (int(text.FirstWeekday) + i) % 7
		week.Child(el.Div().W(el.Dp(cw)).H(el.Dp(24)).Center().Child(el.Text(text.Weekdays[wd]).TextSize(12).TextColor(theme.Muted)))
	}
	grid := el.Div().Role("grid").Name(title).Child(week)
	first := v.month.AddDate(0, 0, -((int(v.month.Weekday()) - int(text.FirstWeekday) + 7) % 7))
	for w := 0; w < 6; w++ {
		row := el.Div().Row()
		for i := 0; i < 7; i++ {
			d := first.AddDate(0, 0, w*7+i)
			row.Child(v.cell(d, today, cw, ch, move))
		}
		grid.Child(row)
	}
	return el.Div().Gap(8).Items(el.Start).Child(head, grid)
}

func (v *CalendarView) cell(d, today time.Time, cw, ch float32, move func(time.Time)) el.Element {
	inMonth := monthOf(d) == v.month
	ok := v.allowed(d)
	chosen := !v.start.IsZero() && (d.Equal(v.start) || !v.end.IsZero() && d.Equal(v.end))
	between := v.rangeMode && !v.end.IsZero() && d.After(v.start) && d.Before(v.end)
	var bg color.NRGBA // transparent unless chosen or in the range
	fg := theme.Text
	switch {
	case chosen:
		bg, fg = theme.Primary, theme.OnColor
	case between:
		bg = theme.Highlight
	}
	if !inMonth || !ok {
		fg = theme.Muted
	}
	label := locale.Current().Date(d)
	c := el.Div().ID(v.cellID(d)).Role("gridcell").Name(label).Selected(chosen || between).
		W(el.Dp(cw)).H(el.Dp(ch)).Rounded(6).Bg(bg).TextColor(fg).Center().
		Focusable(d.Equal(v.focus)).Disabled(!ok).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		Child(el.Text(strconv.Itoa(d.Day())))
	if d.Equal(today) && !chosen {
		c.Border(1, theme.Border)
	}
	if ok {
		c.CursorPointer().OnClick(func() { v.pick(d); move(d) }).OnKey(func(e el.KeyEvent) bool {
			var to time.Time
			switch key.Name(e.Name) {
			case key.NameLeftArrow:
				to = d.AddDate(0, 0, -1)
			case key.NameRightArrow:
				to = d.AddDate(0, 0, 1)
			case key.NameUpArrow:
				to = d.AddDate(0, 0, -7)
			case key.NameDownArrow:
				to = d.AddDate(0, 0, 7)
			case key.NamePageUp:
				to = calendarMonthDay(d, -1)
			case key.NamePageDown:
				to = calendarMonthDay(d, 1)
			case key.NameHome:
				to = d.AddDate(0, 0, -((int(d.Weekday()) - int(locale.Current().FirstWeekday) + 7) % 7))
			case key.NameEnd:
				to = d.AddDate(0, 0, 6-(int(d.Weekday())-int(locale.Current().FirstWeekday)+7)%7)
			default:
				return false
			}
			if e.State == el.KeyPress {
				move(to)
			}
			return true
		})
		if !chosen {
			c.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	return c
}

// calendarMonthDay keeps navigation within the destination month. AddDate on
// Jan 31 otherwise normalizes February 31 into March.
func calendarMonthDay(d time.Time, delta int) time.Time {
	first := monthOf(d).AddDate(0, delta, 0)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d.Day(), last)-1)
}
