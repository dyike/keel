package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
	"time"
)

func TestDatePickerFormatAndClear(t *testing.T) {
	for _, span := range []bool{false, true} {
		calls := 0
		d := DatePicker("Dates").Format("02/01/2006").Clearable(true).OnChange(func(a, b time.Time) {
			calls++
			if !a.IsZero() || !b.IsZero() {
				t.Fatal("nonempty clear callback")
			}
		})
		if span {
			d.Range()
		}
		d.SetValue(testDay("2026-10-01"), testDay("2026-10-05"))
		want := "01/10/2026"
		if span {
			want += " – 05/10/2026"
		}
		if d.text() != want || calls != 0 {
			t.Fatal("format", d.text(), calls)
		}
		h := page(d)
		loc := locale.Current()
		name := loc.Name(loc.Clear, "Dates")
		d.SetError("required")
		h.Frame()
		click(t, h, name)
		h.Frame()
		a, b := d.Value()
		if !a.IsZero() || !b.IsZero() || d.open || d.Error() != "" || calls != 1 {
			t.Fatal("clear opened calendar or retained state")
		}
		if shown(h, name) {
			t.Fatal("empty clear button still visible")
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if !d.open {
			t.Fatal("focus not restored to trigger")
		}
		d.close()
		d.SetValue(testDay("2026-10-01"), time.Time{})
		d.Format("")
		if d.text() != locale.Current().Date(testDay("2026-10-01")) {
			t.Fatal("locale not restored")
		}
	}
}

func TestDatePickerClearDisabledAndHidden(t *testing.T) {
	disabled := true
	calls := 0
	d := DatePicker("Dates").Clearable(true).OnChange(func(time.Time, time.Time) { calls++ })
	d.SetValue(testDay("2026-10-01"), time.Time{})
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(d.Render(cx)) })
	loc := locale.Current()
	name := loc.Name(loc.Clear, "Dates")
	disabled = false
	h.Frame()
	r := bounds(h, name)
	disabled = true
	h.Frame()
	h.Click(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2))
	if a, _ := d.Value(); a.IsZero() || calls != 0 {
		t.Fatal("ancestor disabled cleared")
	}
	disabled = false
	d.SetDisabled(true)
	h.Frame()
	h.Click(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2))
	if a, _ := d.Value(); a.IsZero() || calls != 0 {
		t.Fatal("disabled cleared")
	}
	d.SetDisabled(false)
	d.Clearable(false)
	h.Frame()
	if shown(h, name) {
		t.Fatal("clearable false ignored")
	}
}
