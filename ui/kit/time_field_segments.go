package kit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Segmented enables independently editable hour and minute fields.
func (v *TimeFieldView) Segmented() *TimeFieldView { v.segmented = true; v.syncParts(); return v }

// Seconds enables segmented editing with second precision.
func (v *TimeFieldView) Seconds() *TimeFieldView { v.seconds = true; return v.Segmented() }

// Hour12 overrides the locale's segmented clock format without changing Value.
func (v *TimeFieldView) Hour12(on bool) *TimeFieldView { v.hour12 = &on; return v.Segmented() }
func (v *TimeFieldView) uses12() bool {
	if v.hour12 != nil {
		return *v.hour12
	}
	return locale.Current().Clock12
}
func (v *TimeFieldView) segmentID(i int) string {
	return autoID("time", v) + "/part/" + strconv.Itoa(i)
}
func (v *TimeFieldView) syncParts() {
	h := int(v.value / time.Hour)
	if v.uses12() {
		h = (h+11)%12 + 1
	}
	v.parts = [3]string{fmt.Sprintf("%02d", h), fmt.Sprintf("%02d", int(v.value/time.Minute)%60), fmt.Sprintf("%02d", int(v.value/time.Second)%60)}
}
func (v *TimeFieldView) segmentDraft(i int) time.Duration {
	n, err := strconv.Atoi(v.parts[i])
	if err != nil {
		return v.value
	}
	lo, hi := 0, 59
	if i == 0 {
		hi = 23
		if v.uses12() {
			lo, hi = 1, 12
		}
	}
	if n < lo || n > hi {
		return v.value
	}
	unit := []time.Duration{time.Hour, time.Minute, time.Second}[i]
	old := (v.value / unit) % 60
	if i == 0 {
		old = v.value / time.Hour
		if v.uses12() {
			n = n%12 + int(old/12)*12
		}
	}
	return v.value + (time.Duration(n)-old)*unit
}
func (v *TimeFieldView) setUser(d time.Duration) {
	old := v.value
	v.SetValue(d)
	v.err = ""
	if old != v.value && v.onChange != nil {
		v.onChange(v.value)
	}
}
func (v *TimeFieldView) commitSegment(i int) {
	if !v.disabled {
		v.setUser(v.segmentDraft(i))
	}
}

type timeBlurKey struct {
	view *TimeFieldView
	part int
}

func (v *TimeFieldView) renderSegments(cx *el.Context) el.Element {
	id := autoID("time", v)
	height, font, icon, padding, verticalPadding := v.sizeMetrics()
	mode := v.uses12()
	if !v.localeReady || mode != v.lastHour12 {
		v.syncParts()
		v.lastHour12 = mode
		v.localeReady = true
	}
	focused := cx.FocusWithin(id)
	if !cx.Enabled(id) {
		v.partFocused = [3]bool{}
		v.syncParts()
	}
	box := fieldFrame(id, focused, v.err != "", v.disabled, false).FocusOnPress(v.segmentID(0)).W(el.Auto).MinH(el.Dp(height)).Px(padding).Py(verticalPadding).TextSize(font).Gap(theme.SpaceXs * font / theme.TextControl).
		Child(Icon(IconClock).Size(icon).Color(theme.Muted).Render(cx))
	text := locale.Current()
	names := []string{text.Hour, text.Minute, text.Second}
	count := 2
	if v.seconds {
		count = 3
	}
	for i := range count {
		partID := v.segmentID(i)
		active := cx.Focused(partID)
		if v.partFocused[i] && !active {
			cx.AfterEnabled(id, timeBlurKey{v, i}, 0, func() { v.commitSegment(i); v.partFocused[i] = false })
		} else {
			v.partFocused[i] = active
		}
		if i > 0 {
			box.Child(el.Text(":"))
		}
		field := fieldText(el.Input().ID(partID).Name(text.Name(names[i], v.a11y())).Bind(&v.parts[i])).Filter("0123456789").MaxLen(2).
			TextSize(font).W(el.Dp(28 * font / theme.TextControl)).NoShrink().
			OnSubmit(func(string) { v.commitSegment(i) }).
			OnChange(func(text string) { v.segmentChanged(cx, i, text) }).
			OnKey(func(e el.KeyEvent) bool {
				if v.segmentKeys {
					return v.segmentKey(cx, i, e)
				}
				steps := map[key.Name]int{key.NameUpArrow: 1, key.NameDownArrow: -1, key.NamePageUp: 10, key.NamePageDown: -10}
				step, ok := steps[key.Name(e.Name)]
				if !ok || e.Modifiers != 0 {
					return false
				}
				if e.State == el.KeyPress {
					unit := []time.Duration{time.Hour, time.Minute, time.Second}[i]
					v.setUser(v.segmentDraft(i) + time.Duration(step)*unit)
				}
				return true
			})
		if v.segmentKeys {
			field.SelectOnFocus(true).CaptureKeys(string(key.NameLeftArrow), string(key.NameRightArrow), string(key.NameDeleteBackward), string(key.NameDeleteForward))
			if mode {
				field.CaptureKeys(string(key.NameLeftArrow), string(key.NameRightArrow), string(key.NameDeleteBackward), string(key.NameDeleteForward), "A", "P")
			}
		}
		box.Child(field)
	}
	if mode {
		period := text.AM
		if v.value >= 12*time.Hour {
			period = text.PM
		}
		periodButton := Button(period, func() {
			base := v.value
			for i, active := range v.partFocused {
				if active {
					base = v.segmentDraft(i)
					break
				}
			}
			v.setUser(base + 12*time.Hour)
		}).ID(v.periodID()).Name(text.Name(text.Period, v.a11y())).Variant(ButtonGhost).Size(max(16, height-8)).Render(cx)
		if v.segmentKeys {
			periodButton.(*el.DivEl).OnKey(func(e el.KeyEvent) bool { return v.segmentKey(cx, count, e) })
		}
		box.Child(periodButton)
	}
	return labelled(v.label, el.Div().Items(el.Start).Child(box), v.err)
}
