package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"testing"
	"time"
)

func TestCalendarWeekdayLayoutAndNavigation(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, first := range []time.Weekday{time.Sunday, time.Monday, time.Wednesday} {
			cal := Calendar().FirstWeekday(first)
			cal.SetValue(testDay("2026-10-08"), time.Time{})
			h := renderView(cal, 400, scale)
			head := bounds(h, locale.Current().Weekdays[first])
			firstDate := testDay("2026-10-04")
			for firstDate.Weekday() != first {
				firstDate = firstDate.AddDate(0, 0, 1)
			}
			cell := bounds(h, firstDate.Format("2006-01-02"))
			if head.Empty() || cell.Empty() || head.Min.X+head.Dx()/2 != cell.Min.X+cell.Dx()/2 {
				t.Fatal("header/day mismatch", first, head, cell)
			}
			click(t, h, "2026-10-08")
			h.Key(key.NameHome, 0)
			h.Key(key.NameReturn, 0)
			a, _ := cal.Value()
			if a.Weekday() != first {
				t.Fatal("Home", first, a)
			}
			h.Key(key.NameEnd, 0)
			h.Key(key.NameReturn, 0)
			a, _ = cal.Value()
			if a.Weekday() != time.Weekday((int(first)+6)%7) {
				t.Fatal("End", first, a)
			}
		}
	}
}

func TestDatePickerWeekdayPreservesSelectionAndDraft(t *testing.T) {
	calls := 0
	d := DatePicker("Dates").Range().FirstWeekday(time.Sunday).OnChange(func(time.Time, time.Time) { calls++ })
	d.SetValue(testDay("2026-10-01"), testDay("2026-10-02"))
	h := renderView(d, 500, 1)
	clickClass(t, h, "Button", "Dates")
	h.Frame()
	click(t, h, "2026-10-10")
	d.FirstWeekday(time.Wednesday)
	h.Frame()
	h.Frame()
	if !d.open || !d.cal.RangePending() || calls != 0 {
		t.Fatal("override changed interaction state")
	}
	revealed := d.revealed
	d.FirstWeekday(time.Wednesday)
	if !d.revealed.Equal(revealed) {
		t.Fatal("repeated configuration retriggered focus reveal")
	}
	d.FirstWeekday(time.Weekday(7))
	if d.cal.weekStart() != time.Wednesday {
		t.Fatal("invalid override accepted")
	}
	d.ResetFirstWeekday()
	h.Frame()
	if d.cal.weekStart() != locale.Current().FirstWeekday || !d.cal.RangePending() {
		t.Fatal("reset discarded draft")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	a, b := d.Value()
	if a.Day() != 1 || b.Day() != 2 || calls != 0 {
		t.Fatal("configuration changed committed dates")
	}
}
