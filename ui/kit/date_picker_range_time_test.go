package kit

import (
	"testing"
	"time"
)

func TestDatePickerRangeTimeStorage(t *testing.T) {
	zone := time.FixedZone("range", 8*3600)
	a := time.Date(2026, 10, 3, 9, 15, 32, 123, zone)
	b := time.Date(2026, 10, 5, 18, 45, 41, 456, zone)
	d := DatePicker("Dates").Range()
	calls := 0
	var gotA, gotB time.Time
	d.OnChange(func(a, b time.Time) { calls++; gotA, gotB = a, b })
	d.SetDateTimeValue(b, a)
	x, y := d.DateTimeValue()
	if !x.Equal(a) || !y.Equal(b) || calls != 0 {
		t.Fatal(x, y, calls)
	}
	x, y = d.Value()
	if x.Hour() != 0 || y.Hour() != 0 {
		t.Fatal("legacy Value must retain dates")
	}
	d.SetDateValue(a.AddDate(0, 0, 1), b.AddDate(0, 0, 1))
	x, y = d.DateTimeValue()
	if x.Day() != 4 || x.Hour() != 9 || y.Day() != 6 || y.Hour() != 18 {
		t.Fatal(x, y)
	}
	d.open = true
	d.cal.pick(testDay("2026-10-10"))
	if calls != 0 {
		t.Fatal("range draft emitted")
	}
	d.cal.pick(testDay("2026-10-12"))
	if calls != 1 || gotA.Hour() != 9 || gotB.Hour() != 18 || d.open {
		t.Fatal(gotA, gotB, calls)
	}
	d.Format("2006-01-02 15:04")
	if d.text() != "2026-10-10 09:15 – 2026-10-12 18:45" {
		t.Fatal(d.text())
	}
	d.SetDateTimeValue(time.Time{}, time.Time{})
	x, y = d.DateTimeValue()
	if !x.IsZero() || !y.IsZero() {
		t.Fatal("empty manufactured date")
	}
}

func TestDatePickerRangeTimePrecisionAndPreset(t *testing.T) {
	a := time.Date(2026, 10, 3, 18, 15, 32, 123, time.UTC)
	b := time.Date(2026, 10, 3, 9, 45, 41, 456, time.UTC)
	for _, seconds := range []bool{false, true} {
		d := DatePicker("Dates").Range().WithTime()
		if seconds {
			d.TimeSeconds()
		}
		d.Presets(DatePickerPreset{ID: "span", Label: "Span", Start: a, End: b, IncludeTime: true})
		if !d.selectPreset("span") {
			t.Fatal("preset rejected")
		}
		x, y := d.DateTimeValue()
		if x.Hour() != 9 || y.Hour() != 18 || x.Nanosecond() != 0 || y.Nanosecond() != 0 {
			t.Fatal(x, y)
		}
		if seconds && (x.Second() != 41 || y.Second() != 32) {
			t.Fatal(x, y)
		}
		if !seconds && (x.Second() != 0 || y.Second() != 0) {
			t.Fatal(x, y)
		}
		d.SetDateTimeValue(a, time.Time{})
		x, y = d.DateTimeValue()
		if !x.Equal(y) {
			t.Fatal("missing end did not copy start")
		}
		d.DefaultTime(7 * time.Hour)
		x, y = d.DateTimeValue()
		if x.Hour() != 7 || y.Hour() != 7 {
			t.Fatal("default time not applied to both", x, y)
		}
	}
}

func TestDatePickerEnablingSecondsPreservesStoredSeconds(t *testing.T) {
	d := DatePicker("Dates").Range()
	a := time.Date(2026, 10, 3, 9, 15, 32, 123, time.UTC)
	b := a.Add(3*time.Hour + 9*time.Second)
	d.SetDateTimeValue(a, b)
	d.TimeSeconds()
	x, y := d.DateTimeValue()
	if x.Second() != 32 || y.Second() != 41 || x.Nanosecond() != 0 || y.Nanosecond() != 0 {
		t.Fatal(x, y)
	}
}
