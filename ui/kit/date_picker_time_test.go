package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
	"image"
	"testing"
	"time"
)

func TestDatePickerTimeSelection(t *testing.T) {
	calls := 0
	d := DatePicker("Appointment").TimeSeconds().TimeHour12(false).DefaultTime(9*time.Hour + 30*time.Minute + 12*time.Second)
	d.OnChange(func(a, b time.Time) {
		calls++
		if a.Hour() != 9 || a.Minute() != 30 || a.Second() != 12 || !a.Equal(b) {
			t.Fatalf("callback %v %v", a, b)
		}
	})
	d.open = true
	date := testDay("2026-10-03")
	d.cal.pick(date)
	a, _ := d.Value()
	if !d.open || calls != 1 || a.Hour() != 9 {
		t.Fatal("date did not preserve clock", a, calls)
	}
	d.cal.pick(date)
	if d.open || calls != 1 {
		t.Fatal("repeat date should close without callback")
	}
	d.SetValue(time.Date(2026, 10, 4, 13, 14, 15, 999, time.FixedZone("test", 8*3600)), time.Time{})
	a, b := d.Value()
	if a.Hour() != 13 || a.Second() != 15 || a.Nanosecond() != 0 || !a.Equal(b) || calls != 1 {
		t.Fatal(a, b, calls)
	}
	if d.text() != locale.Current().Date(a)+" 13:14:15" {
		t.Fatal(d.text())
	}
	d.Format("2006-01-02 15:04:05")
	if d.text() != "2026-10-04 13:14:15" {
		t.Fatal(d.text())
	}
	d.Range()
	a, _ = d.Value()
	if a.Hour() != 0 || d.editsTime() {
		t.Fatal("range must remain date-only")
	}
}

func TestDatePickerTimeKeyboardAndDismiss(t *testing.T) {
	calls := 0
	d := DatePicker("Appointment").WithTime().TimeHour12(false).Clearable(true).OnChange(func(time.Time, time.Time) { calls++ })
	d.SetValue(time.Date(2026, 10, 3, 9, 30, 45, 0, time.UTC), time.Time{})
	root := el.Root(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(400, 700); root.Layout(gtx) })
	clickClass(t, h, "Button", "Appointment")
	h.Frame()
	name := locale.Current().Name(locale.Current().Hour, "Appointment")
	click(t, h, name)
	h.Key(key.NameUpArrow, 0)
	h.Frame()
	a, _ := d.Value()
	if a.Hour() != 10 || a.Second() != 0 || calls != 1 || !d.open {
		t.Fatal("time keyboard", a, calls, d.open)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if d.open {
		t.Fatal("escape failed")
	}
	a, _ = d.Value()
	if a.Hour() != 10 {
		t.Fatal("escape reverted committed time")
	}
	d.open = true
	d.clock.parts[0] = "22"
	d.close()
	if d.clock.parts[0] != "10" {
		t.Fatal("close retained draft")
	}
	d.SetValue(time.Time{}, time.Time{})
	d.open = true
	d.clock.setUser(11 * time.Hour)
	if calls != 1 {
		t.Fatal("empty date emitted time change")
	}
	if a, _ = d.Value(); !a.IsZero() {
		t.Fatal("time manufactured date")
	}
}

func TestDatePickerTimePreset(t *testing.T) {
	d := DatePicker("Appointment").WithTime().DefaultTime(9 * time.Hour).Presets(DatePickerPreset{ID: "day", Label: "Day", Start: testDay("2026-10-03")})
	var got time.Time
	d.OnChange(func(a, b time.Time) { got = a })
	if !d.selectPreset("day") || got.Hour() != 9 {
		t.Fatal("date preset lost clock", got)
	}
	d.open = true
	d.cal.pick(testDay("2026-10-03"))
	if d.open {
		t.Fatal("preset did not update selected day")
	}
}
