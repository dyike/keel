package kit

import (
	"image/color"
	"strconv"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CalendarView shows one month and picks a date, or a range with Range.
// Tab reaches the focused day; arrows move by a day or week, PageUp and
// PageDown by a month, Home and End to the week's ends, Enter or Space picks.
// Week layout, day and month names come from ui/locale.
type CalendarView struct {
	size                  CalendarSize
	months                int
	firstWeekday          *time.Weekday
	draft                 time.Time
	rangeError            bool
	choosing, yearEditing bool
	chooseYear            int
	yearPicker            *NumberInputView
	start, end            time.Time // chosen dates at midnight; end == start unless Range
	rangeMode             bool
	pending               bool // Range: the first end is picked, the second is not
	month, focus          time.Time
	lo, hi                time.Time
	blocked               func(time.Time) bool
	disabled              bool
	onChange              func(start, end time.Time)
}

func Calendar() *CalendarView { return &CalendarView{months: 1} }

// FirstWeekday overrides the locale's first weekday. Invalid values are ignored.
// The override controls weekday headers, day layout and Home/End navigation.
func (v *CalendarView) FirstWeekday(day time.Weekday) *CalendarView {
	if day >= time.Sunday && day <= time.Saturday {
		v.firstWeekday = &day
	}
	return v
}

// ResetFirstWeekday restores the current locale's week layout.
func (v *CalendarView) ResetFirstWeekday() *CalendarView { v.firstWeekday = nil; return v }
func (v *CalendarView) weekStart() time.Weekday {
	if v.firstWeekday != nil {
		return *v.firstWeekday
	}
	return locale.Current().FirstWeekday
}

// Range picks a span: the first click sets one end, the second the other.
func (v *CalendarView) Range() *CalendarView { v.rangeMode = true; return v }

// Bounds limits selectable dates to [min, max]; a zero time leaves that side open.
func (v *CalendarView) Bounds(min, max time.Time) *CalendarView {
	v.lo, v.hi = day(min), day(max)
	if !v.lo.IsZero() && !v.hi.IsZero() && v.hi.Before(v.lo) {
		v.lo, v.hi = v.hi, v.lo
	}
	v.CancelRange()
	return v
}

// DisableDates blocks dates for which fn returns true, e.g. weekends.
func (v *CalendarView) DisableDates(fn func(time.Time) bool) *CalendarView {
	v.blocked = fn
	v.CancelRange()
	return v
}

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
	v.CancelRange()
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
func (v *CalendarView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.CancelRange()
		v.choosing = false
		v.yearEditing = false
	}
}

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
		v.draft, v.pending, v.rangeError = d, true, false
		return
	default:
		start, end := v.draft, d
		if end.Before(start) {
			start, end = end, start
		}
		// A range cannot bridge a disabled date. Restart from the clicked endpoint.
		for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
			if !v.allowed(date) {
				v.draft = d
				v.rangeError = true
				return
			}
		}
		v.start, v.end = start, end
		v.CancelRange()
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
	last := v.month.AddDate(0, max(1, v.months), 0)
	if v.focus.IsZero() || v.focus.Before(v.month) || !v.focus.Before(last) {
		v.focus = v.month
		if monthOf(today) == v.month {
			v.focus = today
		}
		if !v.start.IsZero() && monthOf(v.start) == v.month {
			v.focus = v.start
		}
	}
	if !v.allowed(v.focus) {
		if next, ok := v.seek(v.focus, 1); ok {
			v.focus = next
		}
	}
	move := func(d time.Time) { v.moveFocus(cx, d) }
	prevName, nextName := text.PrevMonth, text.NextMonth
	if v.choosing {
		prevName, nextName = text.PrevYear, text.NextYear
	}
	prev := Button("", func() {
		if v.choosing {
			v.chooseYear = max(1, v.chooseYear-1)
			v.yearEditing = false
		} else {
			v.month = v.month.AddDate(0, -1, 0)
		}
	}).Name(prevName).Icon(IconChevronLeft).Variant(ButtonGhost).Size(v.metrics().nav)
	next := Button("", func() {
		if v.choosing {
			v.chooseYear = min(9999, v.chooseYear+1)
			v.yearEditing = false
		} else {
			v.month = v.month.AddDate(0, 1, 0)
		}
	}).Name(nextName).Icon(IconChevronRight).Variant(ButtonGhost).Size(v.metrics().nav)
	prev.SetDisabled(v.disabled || !v.choosing && !v.monthAllowed(v.month.AddDate(0, -1, 0)))
	next.SetDisabled(v.disabled || !v.choosing && !v.monthAllowed(v.month.AddDate(0, 1, 0)))
	title := Button(v.monthTitle(), func() { v.choosing = !v.choosing; v.chooseYear = v.month.Year(); v.yearEditing = false }).Variant(ButtonGhost).Size(v.metrics().nav)
	head := el.Div().Row().WFull().Items(el.Center).Child(prev.Render(cx), el.Div().Grow().Items(el.Center).Child(title.Render(cx)), next.Render(cx))
	body := el.Div().Row().Wrap().Gap(theme.SpaceXl).WFull()
	count := max(1, v.months)
	if v.choosing {
		body.Child(v.yearMonths(cx))
		count = 1
	} else {
		for i := range count {
			body.Child(v.monthGrid(v.month.AddDate(0, i, 0), today, move))
		}
	}
	root := el.Div().Disabled(v.disabled).W(el.Dp(float32(count)*v.metrics().width*7+float32(count-1)*16)).MaxW(el.Full).Gap(theme.SpaceMd).Items(el.Stretch).Child(head, body)
	root.OnKey(func(e el.KeyEvent) bool {
		if key.Name(e.Name) == key.NameEscape && v.pending {
			if e.State == el.KeyPress {
				v.CancelRange()
			}
			return true
		}
		return false
	})
	if v.pending {
		root.Child(Button(text.Cancel, func() { v.CancelRange(); cx.Focus(v.FocusID()) }).Variant(ButtonGhost).Render(cx))
	}
	if v.rangeError {
		root.Child(el.Text(text.RangeUnavailable).TextColor(theme.Danger))
	}
	return root
}
func (v *CalendarView) monthGrid(month, today time.Time, move func(time.Time)) el.Element {
	text := locale.Current()
	m := v.metrics()
	cw, ch := m.width, m.height
	week := el.Div().WFull().Row()
	for i := 0; i < 7; i++ {
		wd := (int(v.weekStart()) + i) % 7
		week.Child(el.Div().W(el.Dp(0)).Grow().MaxW(el.Dp(cw)).H(el.Dp(m.weekday)).Center().Child(el.Text(text.Weekdays[wd]).TextSize(theme.TextSm).TextColor(theme.Muted)))
	}
	title := text.Month(month.Year(), month.Month())
	grid := el.Div().W(el.Dp(cw * 7)).MaxW(el.Full).NoShrink().Role("grid").Name(title)
	if v.months > 1 {
		// Same weight as the header button text: the header names the range.
		grid.Child(el.Div().H(el.Dp(m.nav)).Center().Child(el.Text(title).TextSize(theme.TextMd)))
	}
	grid.Child(week)
	first := month.AddDate(0, 0, -((int(month.Weekday()) - int(v.weekStart()) + 7) % 7))
	for w := 0; w < 6; w++ {
		row := el.Div().WFull().Row()
		for i := 0; i < 7; i++ {
			d := first.AddDate(0, 0, w*7+i)
			if v.months > 1 && monthOf(d) != month {
				row.Child(el.Div().W(el.Dp(0)).Grow().MaxW(el.Dp(cw)).H(el.Dp(ch)))
			} else {
				row.Child(v.cell(d, today, month, cw, ch, move))
			}
		}
		grid.Child(row)
	}
	return grid
}

func (v *CalendarView) cell(d, today, month time.Time, cw, ch float32, move func(time.Time)) el.Element {
	inMonth := monthOf(d) == month
	ok := v.allowed(d)
	chosen := !v.start.IsZero() && (d.Equal(v.start) || !v.end.IsZero() && d.Equal(v.end))
	between := v.rangeMode && !v.end.IsZero() && d.After(v.start) && d.Before(v.end)
	if v.pending && d.Equal(v.draft) {
		chosen = true
	}
	var bg color.NRGBA // transparent unless chosen or in the range
	fg := theme.Text
	switch {
	case chosen:
		bg, fg = theme.Primary, theme.OnColor
	case between:
		bg = theme.Highlight
	}
	if !inMonth && !chosen {
		fg = theme.Muted
	}
	// Blocked days are fainter than days of other months, so they read as
	// unavailable rather than merely out of range.
	faint := theme.Muted
	faint.A = 0x66
	label := locale.Current().Date(d)
	dayText := el.Text(strconv.Itoa(d.Day()))
	if v.size != CalendarSizeMedium {
		dayText.TextSize(v.metrics().font)
	}
	c := el.Div().ID(v.cellID(d)).Role("gridcell").Name(label).Selected(chosen || between).
		W(el.Dp(0)).Grow().MaxW(el.Dp(cw)).H(el.Dp(ch)).Rounded(theme.RadiusMd).Bg(bg).TextColor(fg).Center().
		Focusable(d.Equal(v.focus)).Disabled(!ok).DisabledStyle(func(s *el.Style) { s.TextColor(faint) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		Child(dayText)
	if d.Equal(today) && !chosen {
		c.Border(1, theme.Border)
	}
	if ok && !chosen && !between {
		c.Hover(func(s *el.Style) { s.Bg(theme.Subtle) })
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
				to = d.AddDate(0, 0, -((int(d.Weekday()) - int(v.weekStart()) + 7) % 7))
			case key.NameEnd:
				to = d.AddDate(0, 0, 6-(int(d.Weekday())-int(v.weekStart())+7)%7)
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
