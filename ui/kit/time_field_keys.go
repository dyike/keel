package kit

import (
	"strconv"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
)

// SegmentKeys enables segmented replacement, automatic advance after two valid
// digits, arrow navigation, deletion reset and independent unit stepping.
// False keeps draft editing with Enter/blur commit and carry between units.
func (v *TimeFieldView) SegmentKeys(on bool) *TimeFieldView {
	v.segmentKeys = on
	if !v.segmented {
		v.Segmented()
	}
	return v
}

func (v *TimeFieldView) segmentCount() int {
	if v.seconds {
		return 3
	}
	return 2
}
func (v *TimeFieldView) periodID() string { return autoID("time", v) + "/period" }
func (v *TimeFieldView) focusSegment(cx *el.Context, i int) {
	if i < 0 {
		return
	}
	if i < v.segmentCount() {
		cx.Focus(v.segmentID(i))
	} else if v.uses12() {
		cx.Focus(v.periodID())
	}
}
func (v *TimeFieldView) segmentLimits(i int) (int, int) {
	if i != 0 {
		return 0, 59
	}
	if v.uses12() {
		return 1, 12
	}
	return 0, 23
}
func (v *TimeFieldView) segmentChanged(cx *el.Context, i int, text string) {
	if !v.segmentKeys || len(text) != 2 {
		return
	}
	n, err := strconv.Atoi(text)
	lo, hi := v.segmentLimits(i)
	if err != nil || n < lo || n > hi {
		return
	}
	v.partFocused[i] = false
	v.commitSegment(i)
	v.focusSegment(cx, i+1)
	cx.SelectInput(v.segmentID(i), 0, 2)
}
func (v *TimeFieldView) setPeriod(pm bool) {
	d := v.value
	for i, active := range v.partFocused {
		if active {
			d = v.segmentDraft(i)
			break
		}
	}
	d %= 12 * time.Hour
	if pm {
		d += 12 * time.Hour
	}
	v.partFocused = [3]bool{}
	v.setUser(d)
}
func (v *TimeFieldView) segmentKey(cx *el.Context, i int, e el.KeyEvent) bool {
	if e.Modifiers != 0 {
		return false
	}
	name := key.Name(e.Name)
	if name == key.NameLeftArrow || name == key.NameRightArrow {
		if e.State == el.KeyPress {
			if i < v.segmentCount() {
				v.commitSegment(i)
				v.partFocused[i] = false
			}
			delta := 1
			if name == key.NameLeftArrow {
				delta = -1
			}
			v.focusSegment(cx, i+delta)
		}
		return true
	}
	if v.uses12() && (e.Name == "A" || e.Name == "P") {
		if e.State == el.KeyPress {
			v.setPeriod(e.Name == "P")
		}
		return true
	}
	reset := name == key.NameDeleteBackward || name == key.NameDeleteForward
	steps := map[key.Name]int{key.NameUpArrow: 1, key.NameDownArrow: -1, key.NamePageUp: 10, key.NamePageDown: -10}
	step, stepping := steps[name]
	if !reset && !stepping {
		return false
	}
	if e.State != el.KeyPress {
		return true
	}
	if i == v.segmentCount() {
		if reset {
			v.setPeriod(false)
		} else {
			v.setPeriod(v.value < 12*time.Hour)
		}
		return true
	}
	lo, hi := v.segmentLimits(i)
	n, err := strconv.Atoi(v.parts[i])
	if err != nil || n < lo || n > hi {
		n = lo
	}
	if reset {
		n = lo
		if i == 0 && v.uses12() {
			n = 12
		}
	} else {
		span := hi - lo + 1
		n = lo + ((n-lo+step)%span+span)%span
	}
	v.parts[i] = strconv.Itoa(n)
	v.commitSegment(i)
	cx.SelectInput(v.segmentID(i), 0, 2)
	return true
}
