package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"testing"
	"time"
)

func TestDatePickerPresetValidationAndCopies(t *testing.T) {
	a, b := testDay("2026-10-01"), testDay("2026-10-05")
	items := []DatePickerPreset{{ID: "week", Label: "Week", Start: b, End: a}, {ID: "week", Label: "Duplicate", Start: a, End: a}, {Label: "Empty ID", Start: a, End: a}}
	calls := 0
	d := DatePicker("Dates").Range().Presets(items...).OnChange(func(time.Time, time.Time) { calls++ })
	items[0].Start = time.Time{}
	if len(d.presets) != 1 || !d.selectPreset("week") {
		t.Fatal("slice copy or dedup")
	}
	start, end := d.Value()
	if !start.Equal(a) || !end.Equal(b) || calls != 1 {
		t.Fatal("range normalization")
	}
	d.DisableDates(func(t time.Time) bool { return t.Day() == 3 })
	if d.selectPreset("week") || calls != 1 {
		t.Fatal("range bridged disabled date")
	}
	d.DisableDates(nil).Bounds(a.AddDate(0, 0, 1), b)
	if d.selectPreset("week") {
		t.Fatal("preset bypassed bounds")
	}
	d.Bounds(time.Time{}, time.Time{}).SetDisabled(true)
	if d.selectPreset("week") {
		t.Fatal("disabled selection")
	}
	d.SetDisabled(false)
	d.Presets(DatePickerPreset{ID: "empty", Label: "Empty"})
	if d.selectPreset("empty") || d.selectPreset("week") {
		t.Fatal("empty or stale preset accepted")
	}
	d.Presets()
	if len(d.presets) != 0 {
		t.Fatal("presets not cleared")
	}
}

func TestDatePickerPresetClickAndKeyboard(t *testing.T) {
	for _, span := range []bool{false, true} {
		calls := 0
		d := DatePicker("Dates").Presets(DatePickerPreset{ID: "later", Label: "Later", Start: testDay("2026-10-15"), End: testDay("2026-10-20")}).OnChange(func(time.Time, time.Time) { calls++ })
		if span {
			d.Range()
		}
		d.SetValue(testDay("2026-10-01"), time.Time{})
		h := renderView(d, 500, 1)
		clickClass(t, h, "Button", "Dates")
		h.Frame()
		if span {
			click(t, h, "2026-10-10")
			if !d.cal.RangePending() {
				t.Fatal("missing draft")
			}
		}
		click(t, h, "Later")
		h.Frame()
		a, b := d.Value()
		wantEnd := 15
		if span {
			wantEnd = 20
		}
		if a.Day() != 15 || b.Day() != wantEnd || d.open || d.cal.RangePending() || calls != 1 {
			t.Fatal("preset failed", a, b, calls)
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if !d.open {
			t.Fatal("focus not returned")
		}
	}
}
