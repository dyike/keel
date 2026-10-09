package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func testDay(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }

func TestCalendarSkipsDisabledDaysWithoutLosingFocus(t *testing.T) {
	cal := Calendar().DisableDates(func(d time.Time) bool { return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday })
	cal.SetValue(testDay("2026-10-02"), time.Time{})
	h := page(cal)
	click(t, h, "2026-10-02")
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameReturn, 0)
	if s, _ := cal.Value(); s.Format("2006-01-02") != "2026-10-05" {
		t.Fatalf("weekend navigation %v", s)
	}
	h.Key(key.NameLeftArrow, 0)
	h.Key(key.NameReturn, 0)
	if s, _ := cal.Value(); s.Format("2006-01-02") != "2026-10-02" {
		t.Fatal("reverse weekend navigation")
	}
	cal.Bounds(testDay("2026-10-01"), testDay("2026-10-04"))
	h.Frame()
	h.Key(key.NamePageDown, 0)
	h.Key(key.NameReturn, 0)
	if s, _ := cal.Value(); s.Format("2006-01-02") != "2026-10-02" {
		t.Fatal("disabled upper boundary")
	}
	cal.DisableDates(func(time.Time) bool { return true })
	if _, ok := cal.seek(testDay("2026-10-01"), 1); ok {
		t.Fatal("all disabled seek")
	}
}

func TestCalendarPendingRangeCancelAndBlockedInterior(t *testing.T) {
	calls := 0
	cal := Calendar().Range().OnChange(func(time.Time, time.Time) { calls++ })
	cal.SetValue(testDay("2026-10-01"), testDay("2026-10-02"))
	cal.pick(testDay("2026-10-10"))
	if !cal.RangePending() {
		t.Fatal("no draft")
	}
	if s, e := cal.Value(); s.Day() != 1 || e.Day() != 2 {
		t.Fatal("draft replaced committed range")
	}
	cal.CancelRange()
	if cal.RangePending() || calls != 0 {
		t.Fatal("cancel committed")
	}
	cal.DisableDates(func(d time.Time) bool { return d.Day() == 12 })
	cal.pick(testDay("2026-10-10"))
	cal.pick(testDay("2026-10-14"))
	if calls != 0 || !cal.RangePending() || !cal.rangeError {
		t.Fatal("range bridged blocked interior")
	}
	cal.pick(testDay("2026-10-15"))
	if s, e := cal.Value(); s.Day() != 14 || e.Day() != 15 || calls != 1 {
		t.Fatalf("restart %v %v %d", s, e, calls)
	}
	cal.pick(testDay("2026-10-20"))
	cal.SetDisabled(true)
	if cal.RangePending() {
		t.Fatal("disable retained draft")
	}
}

func TestCalendarMultipleMonthsAndYearChooser(t *testing.T) {
	cal := Calendar().Months(2)
	cal.SetMonth(testDay("2026-10-01"))
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return cal.Render(cx) }), 620, 1)
	if !shown(h, "2026-10-15") || !shown(h, "2026-11-15") {
		t.Fatal("missing second month")
	}
	click(t, h, "2026-11-15")
	if cal.month.Month() != time.October {
		t.Fatal("second month click unnecessarily shifted panels")
	}
	click(t, h, "2026年10月 – 2026年11月")
	if !shown(h, "年份") || !shown(h, "2月") {
		t.Fatal("month/year chooser missing")
	}
	cal.yearPicker.text = "2030"
	cal.yearPicker.commit()
	h.Frame()
	click(t, h, "2月")
	h.Frame()
	if cal.month.Year() != 2030 || cal.month.Month() != time.February || cal.choosing {
		t.Fatal("year/month jump")
	}
}
