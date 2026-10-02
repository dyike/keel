package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
	"time"
)

func TestCalendarNarrowWeekAndLastColumn(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Calendar()
		d := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
		v.SetValue(d, time.Time{})
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(224)).Child(v.Render(cx)) }), 224, scale)
		for i := 0; i < 31; i++ {
			day := d.AddDate(0, 0, i)
			b := bounds(h, locale.Current().Date(day))
			if b.Empty() || b.Min.X < 0 || b.Max.X > 224*scale {
				t.Fatalf("scale %d: %v outside calendar: %v", scale, day, b)
			}
		}
		last := d.AddDate(0, 0, 30)
		click(t, h, locale.Current().Date(last))
		h.Frame()
		start, _ := v.Value()
		if !start.Equal(last) {
			t.Fatal(v.Value())
		}
	}
}
