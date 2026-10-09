package kit

import (
	"fmt"
	"gioui.org/io/key"
	"strconv"
	"strings"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TimeFieldView edits a time of day as HH:MM. It accepts 9:30, 0930 or 930;
// Enter or leaving the field normalizes the text, and text that is not a
// valid time reverts to the last value. ↑ ↓ change it by a minute,
// PageUp and PageDown by an hour, wrapping around midnight.
type TimeFieldView struct {
	size               TimeFieldSize
	segmentKeys        bool
	name               string // accessible name from a Form row when label is empty
	label, text, err   string
	value              time.Duration // since midnight
	disabled, focused  bool
	onChange           func(time.Duration)
	segmented, seconds bool
	hour12             *bool
	parts              [3]string
	partFocused        [3]bool
	lastHour12         bool
	localeReady        bool
}

func TimeField(label string) *TimeFieldView {
	v := &TimeFieldView{label: label}
	v.text = formatClock(0)
	return v
}
func (v *TimeFieldView) OnChange(fn func(time.Duration)) *TimeFieldView { v.onChange = fn; return v }

// Value is the time since midnight, in minutes or whole seconds with Seconds.
func (v *TimeFieldView) Value() time.Duration { return v.value }

// SetValue sets the time without calling OnChange; it wraps around 24 hours.
func (v *TimeFieldView) SetValue(d time.Duration) {
	precision := time.Minute
	if v.seconds {
		precision = time.Second
	}
	v.value = d.Truncate(precision) % (24 * time.Hour)
	if v.value < 0 {
		v.value += 24 * time.Hour
	}
	v.text = formatClock(v.value)
	v.syncParts()
}
func (v *TimeFieldView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.focused = false
		v.partFocused = [3]bool{}
		v.text = formatClock(v.value)
		v.syncParts()
	}
}
func (v *TimeFieldView) SetError(msg string) { v.err = msg }
func (v *TimeFieldView) Error() string       { return v.err }
func (v *TimeFieldView) FocusID() string {
	if v.segmented {
		return v.segmentID(0)
	}
	return autoID("time", v) + "/text"
}

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
	if v.segmented {
		return v.renderSegments(cx)
	}
	id := autoID("time", v)
	height, font, icon, padding, verticalPadding := v.sizeMetrics()
	if v.focused && !cx.Enabled(id) {
		v.focused = false
		v.text = formatClock(v.value)
	}
	focused := cx.FocusWithin(id)
	if v.focused && !focused {
		cx.AfterEnabled(id, timeBlurKey{v, -1}, 0, func() { v.commit(); v.focused = false })
	} else {
		v.focused = focused
	}
	field := fieldText(el.Input().ID(v.FocusID()).Name(v.a11y()).Placeholder("HH:MM").Bind(&v.text)).
		Filter("0123456789:").MaxLen(5).TextSize(font).W(el.Dp(64 * font / theme.TextControl)).NoShrink().
		OnSubmit(func(string) { v.commit() }).
		OnKey(func(e el.KeyEvent) bool {
			if e.State == el.KeyPress {
				steps := map[key.Name]time.Duration{key.NameUpArrow: time.Minute, key.NameDownArrow: -time.Minute, key.NamePageUp: time.Hour, key.NamePageDown: -time.Hour}
				v.commit()
				old := v.value
				v.SetValue(v.value + steps[key.Name(e.Name)])
				if v.value != old && v.onChange != nil {
					v.onChange(v.value)
				}
			}
			return true
		})
	box := fieldFrame(id, focused, v.err != "", v.disabled, false).FocusOnPress(v.FocusID()).W(el.Auto).MinH(el.Dp(height)).Px(padding).Py(verticalPadding).TextSize(font).
		Child(Icon(IconClock).Size(icon).Color(theme.Muted).Render(cx), field)
	return labelled(v.label, el.Div().Items(el.Start).Child(box), v.err)
}

func (v *TimeFieldView) setName(s string) { v.name = s }
func (v *TimeFieldView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}

func (v *TimeFieldView) commitForm() {
	if v.disabled {
		return
	}
	if !v.segmented {
		v.commit()
		return
	}
	for i, active := range v.partFocused {
		if active {
			v.commitSegment(i)
			return
		}
	}
}
