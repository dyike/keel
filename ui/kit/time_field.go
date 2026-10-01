package kit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TimeFieldView edits a time of day as HH:MM. It accepts 9:30, 0930 or 930;
// Enter or leaving the field normalizes the text, and text that is not a
// valid time reverts to the last value.
type TimeFieldView struct {
	name              string // accessible name from a Form row when label is empty
	label, text, err  string
	value             time.Duration // since midnight
	disabled, focused bool
	onChange          func(time.Duration)
}

func TimeField(label string) *TimeFieldView {
	v := &TimeFieldView{label: label}
	v.text = formatClock(0)
	return v
}
func (v *TimeFieldView) OnChange(fn func(time.Duration)) *TimeFieldView { v.onChange = fn; return v }

// Value is the time since midnight, in whole minutes.
func (v *TimeFieldView) Value() time.Duration { return v.value }

// SetValue sets the time without calling OnChange; it wraps around 24 hours.
func (v *TimeFieldView) SetValue(d time.Duration) {
	v.value = d.Truncate(time.Minute) % (24 * time.Hour)
	if v.value < 0 {
		v.value += 24 * time.Hour
	}
	v.text = formatClock(v.value)
}
func (v *TimeFieldView) SetDisabled(on bool) { v.disabled = on }
func (v *TimeFieldView) SetError(msg string) { v.err = msg }
func (v *TimeFieldView) Error() string       { return v.err }
func (v *TimeFieldView) FocusID() string     { return autoID("time", v) + "/text" }

func formatClock(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d/time.Hour), int(d%time.Hour/time.Minute))
}

// parseClock reads H:MM, HH:MM, HMM or HHMM.
func parseClock(s string) (time.Duration, bool) {
	s = strings.TrimSpace(s)
	h, m, colon := strings.Cut(s, ":")
	if !colon {
		if len(s) < 3 || len(s) > 4 {
			return 0, false
		}
		h, m = s[:len(s)-2], s[len(s)-2:]
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || len(m) != 2 || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, false
	}
	return time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute, true
}

func (v *TimeFieldView) commit() {
	d, ok := parseClock(v.text)
	if !ok {
		v.text = formatClock(v.value)
		return
	}
	v.text, v.err = formatClock(d), ""
	if d != v.value {
		v.value = d
		if v.onChange != nil {
			v.onChange(d)
		}
	}
}

func (v *TimeFieldView) Render(cx *el.Context) el.Element {
	id := autoID("time", v)
	focused := cx.FocusWithin(id)
	if v.focused && !focused {
		v.commit()
	}
	v.focused = focused
	border := theme.Border
	switch {
	case v.err != "":
		border = theme.Danger
	case focused && !v.disabled:
		border = theme.Primary
	}
	field := el.Input().ID(v.FocusID()).Name(v.a11y()).Placeholder("HH:MM").Bind(&v.text).
		Filter("0123456789:").MaxLen(5).Border(0, theme.Border).Bg(theme.Surface).P(0).W(el.Dp(64)).
		OnSubmit(func(string) { v.commit() })
	box := el.Div().ID(id).Row().Items(el.Center).Gap(8).Px(10).Py(8).Rounded(6).Border(1, border).Bg(theme.Surface).
		Disabled(v.disabled).Child(Icon(IconClock).Size(16).Color(theme.Muted).Render(cx), field)
	if v.disabled {
		box.Bg(theme.Subtle)
		field.Bg(theme.Subtle)
	}
	return labelled(v.label, el.Div().Items(el.Start).Child(box), v.err)
}

func (v *TimeFieldView) setName(s string) { v.name = s }
func (v *TimeFieldView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
