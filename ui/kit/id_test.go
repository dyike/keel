package kit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAutoIDCompatibility(t *testing.T) {
	var nilView *CalendarView
	for _, view := range []any{Calendar(), nilView, new(int), nil, 42, []int{1}, map[string]int{"a": 1}} {
		for _, kind := range []string{"calendar", "", strings.Repeat("kind", 100)} {
			if got, want := autoID(kind, view), fmt.Sprintf("kit-%s-%p", kind, view); got != want {
				t.Fatalf("%T: %q != %q", view, got, want)
			}
		}
	}
	cal := Calendar()
	for _, d := range []time.Time{{}, time.Date(2026, 10, 13, 0, 0, 0, 0, time.Local), time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)} {
		if got, want := cal.cellID(d), fmt.Sprintf("kit-calendar-%p/%s", cal, d.Format("20060102")); got != want {
			t.Fatalf("date %v: %q != %q", d, got, want)
		}
	}
}

var benchmarkID string

func BenchmarkAutoID(b *testing.B) {
	view := Calendar()
	b.ReportAllocs()
	for b.Loop() {
		benchmarkID = autoID("calendar", view)
	}
}

func BenchmarkCalendarCellID(b *testing.B) {
	view := Calendar()
	date := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkID = view.cellID(date)
	}
}

func BenchmarkTableCellID(b *testing.B) {
	view := Table(Col("Name"))
	view.beginCellIDs()
	b.ReportAllocs()
	for b.Loop() {
		benchmarkID = view.cellID(42, 0)
	}
}
