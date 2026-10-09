package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"testing"
	"time"
)

func TestCalendarSizesPreserveSelectionAndNavigation(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, size := range []CalendarSize{CalendarSizeXSmall, CalendarSizeSmall, CalendarSizeMedium, CalendarSizeLarge} {
			t.Run(fmt.Sprintf("%d/%d", scale, size), func(t *testing.T) {
				v := Calendar().Size(size).FirstWeekday(time.Monday)
				date := testDay("2026-10-08")
				v.SetValue(date, time.Time{})
				calls := 0
				v.OnChange(func(time.Time, time.Time) { calls++ })
				h := renderView(v, 700, scale)
				r := bounds(h, "2026-10-08")
				m := v.metrics()
				if r.Dx() != int(m.width)*scale || r.Dy() != int(m.height)*scale {
					t.Fatal("cell size", r, m)
				}
				click(t, h, "2026-10-08")
				h.Key(key.NameRightArrow, 0)
				h.Key(key.NameReturn, 0)
				start, _ := v.Value()
				if !start.Equal(date.AddDate(0, 0, 1)) {
					t.Fatal("navigation", start)
				}
				before := calls
				v.Size(CalendarSizeLarge)
				v.Size(CalendarSize(255))
				h.Frame()
				start, _ = v.Value()
				if calls != before || !start.Equal(date.AddDate(0, 0, 1)) || v.size != CalendarSizeLarge {
					t.Fatal("size mutation changed value")
				}
				click(t, h, v.monthTitle())
				h.Frame()
				if !v.choosing || bounds(h, locale.Current().MonthNames[0]).Empty() {
					t.Fatal("month chooser unavailable")
				}
			})
		}
	}
}
