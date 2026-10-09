package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"testing"
	"time"
)

func TestCalendarMonthKeysClampDay(t *testing.T) {
	for _, tc := range []struct {
		from, want string
		key        key.Name
	}{
		{"2026-01-31", "2026-02-28", key.NamePageDown},
		{"2024-01-31", "2024-02-29", key.NamePageDown},
		{"2026-03-31", "2026-02-28", key.NamePageUp},
	} {
		t.Run(tc.from+string(tc.key), func(t *testing.T) {
			d, _ := time.Parse("2006-01-02", tc.from)
			cal := Calendar()
			cal.SetValue(d, d)
			h := page(cal)
			click(t, h, tc.from)
			h.Key(tc.key, 0)
			h.Frame()
			if got := cal.focus.Format("2006-01-02"); got != tc.want {
				t.Fatalf("month navigation: %s want %s", got, tc.want)
			}
			h.Key(key.NameReturn, 0)
			if d, _ := cal.Value(); d.Format("2006-01-02") != tc.want {
				t.Fatal("keyboard focus was lost")
			}
		})
	}
}
