package kit

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestDatePickerDismissDiscardsRangeDraft(t *testing.T) {
	for _, mode := range []string{"escape", "outside", "disabled", "ancestor"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			d := DatePicker("Dates").Range().OnChange(func(time.Time, time.Time) { calls++ })
			d.SetValue(testDay("2026-10-01"), testDay("2026-10-02"))
			disabled := false
			h := render(func(cx *el.Context) el.Element { return el.Div().P(10).Disabled(disabled).Child(d.Render(cx)) })
			clickClass(t, h, "Button", "Dates")
			h.Frame()
			click(t, h, "2026-10-10")
			if !d.cal.RangePending() {
				t.Fatal("missing range draft")
			}
			switch mode {
			case "escape":
				h.Key(key.NameEscape, 0)
			case "outside":
				h.Click(399, 299)
			case "disabled":
				d.SetDisabled(true)
			case "ancestor":
				disabled = true
			}
			h.Frame()
			h.Frame()
			if d.open || d.cal.RangePending() {
				t.Fatal("dismiss retained popup or draft")
			}
			s, e := d.Value()
			if s.Day() != 1 || e.Day() != 2 || calls != 0 {
				t.Fatal("dismiss changed committed range")
			}
			disabled = false
			d.SetDisabled(false)
			h.Frame()
			clickClass(t, h, "Button", "Dates")
			h.Frame()
			click(t, h, "2026-10-15")
			if !d.cal.RangePending() || calls != 0 {
				t.Fatal("reopening reused draft")
			}
		})
	}
}

func TestDatePickerOpensOnAllowedDay(t *testing.T) {
	d := DatePicker("Dates").DisableDates(func(d time.Time) bool { return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday })
	d.SetValue(testDay("2026-10-03"), time.Time{})
	h := page(d)
	clickClass(t, h, "Button", "Dates")
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s, _ := d.Value(); s.Format("2006-01-02") != "2026-10-05" || d.open {
		t.Fatalf("disabled initial focus: %v open=%v", s, d.open)
	}
}

func TestDatePickerResponsiveMonths(t *testing.T) {
	for _, width := range []int{320, 800} {
		d := DatePicker("Dates").Months(2)
		d.SetValue(testDay("2026-10-01"), time.Time{})
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(10).Child(d.Render(cx)) }), width, 1)
		clickClass(t, h, "Button", "Dates")
		h.Frame()
		want := 1
		if width == 800 {
			want = 2
		}
		if d.cal.months != want {
			t.Fatalf("width %d months %d", width, d.cal.months)
		}
		if width == 800 && !shown(h, "2026-11-15") {
			t.Fatal("second month missing")
		}
	}
}

func TestDatePickerShortWindowRevealsKeyboardTarget(t *testing.T) {
	d := DatePicker("Dates")
	d.SetValue(testDay("2026-10-01"), time.Time{})
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(10).Child(d.Render(cx)) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(400, 180); root.Layout(gtx) })
	clickClass(t, h, "Button", "Dates")
	h.Frame()
	h.Frame()
	for range 4 {
		h.Key(key.NameDownArrow, 0)
		h.Frame()
		h.Frame()
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s, _ := d.Value(); s.Day() != 29 || d.open {
		t.Fatalf("clipped target lost focus: %v open=%v", s, d.open)
	}
}
