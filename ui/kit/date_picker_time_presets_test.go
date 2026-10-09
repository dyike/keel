package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestDatePickerExplicitTimePreset(t *testing.T) {
	zone := time.FixedZone("preset", 8*3600)
	start := time.Date(2026, 10, 4, 13, 14, 15, 999, zone)
	for _, seconds := range []bool{false, true} {
		d := DatePicker("Appointment").WithTime().DefaultTime(9 * time.Hour)
		if seconds {
			d.TimeSeconds()
		}
		calls := 0
		var got time.Time
		d.OnChange(func(a, b time.Time) {
			calls++
			got = a
			if !a.Equal(b) {
				t.Fatal("single endpoints differ")
			}
		})
		d.Presets(DatePickerPreset{ID: "time", Label: "Afternoon", Start: start, IncludeTime: true})
		d.open = true
		d.SetError("old")
		if !d.selectPreset("time") {
			t.Fatal("preset rejected")
		}
		wantSecond := 0
		if seconds {
			wantSecond = 15
		}
		if got.Hour() != 13 || got.Minute() != 14 || got.Second() != wantSecond || got.Nanosecond() != 0 || got.Location() != zone || calls != 1 || d.open || d.Error() != "" {
			t.Fatal(got, calls, d.open, d.Error())
		}
		d.DisableDates(func(time.Time) bool { return true })
		d.DefaultTime(8 * time.Hour)
		if d.selectPreset("time") || d.clock.Value() != 8*time.Hour || calls != 1 {
			t.Fatal("blocked preset changed time")
		}
	}
}

func TestDatePickerTimePresetClickAndFocus(t *testing.T) {
	d := DatePicker("Appointment").TimeSeconds().Clearable(true)
	d.Presets(DatePickerPreset{ID: "midnight", Label: "Midnight", Start: testDay("2026-10-04"), IncludeTime: true})
	d.DefaultTime(23 * time.Hour)
	calls := 0
	d.OnChange(func(time.Time, time.Time) { calls++ })
	root := el.Root(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(400, 700); root.Layout(gtx) })
	clickClass(t, h, "Button", "Appointment")
	h.Frame()
	click(t, h, "Midnight")
	h.Frame()
	a, _ := d.Value()
	if a.Hour() != 0 || a.Day() != 4 || calls != 1 || d.open {
		t.Fatal(a, calls, d.open)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !d.open {
		t.Fatal("preset did not return trigger focus")
	}
}

func TestDatePickerTimePresetStoredInDateModes(t *testing.T) {
	start := time.Date(2026, 10, 4, 13, 14, 15, 0, time.UTC)
	for _, span := range []bool{false, true} {
		d := DatePicker("Dates")
		if span {
			d.WithTime().DefaultTime(9 * time.Hour).Range()
		}
		d.Presets(DatePickerPreset{ID: "time", Label: "Time", Start: start, End: start.AddDate(0, 0, 2), IncludeTime: true})
		if !d.selectPreset("time") {
			t.Fatal("rejected")
		}
		a, b := d.Value()
		if a.Hour() != 0 || b.Hour() != 0 {
			t.Fatal("date-only mode leaked clock")
		}
		dt, _ := d.DateTimeValue()
		if dt.Hour() != 13 {
			t.Fatal("explicit preset did not store clock", dt)
		}
	}
}
