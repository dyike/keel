package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
	"time"
)

func TestDatePickerSizingAndAppearance(t *testing.T) {
	for _, scale := range []int{1, 2} {
		d := DatePicker("Dates").Clearable(true)
		date := testDay("2026-10-01")
		d.SetValue(date, time.Time{})
		h := renderView(d, 300, scale)
		loc := locale.Current()
		clear := loc.Name(loc.Clear, "Dates")
		normal := bounds(h, clear).Dy()
		clickClass(t, h, "Button", "Dates")
		h.Frame()
		d.Size(48).Appearance(false)
		h.Frame()
		if !d.open {
			t.Fatal("restyle closed popup")
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		large := bounds(h, clear).Dy()
		if large <= normal {
			t.Fatal("large clear button not scaled")
		}
		d.Size(28).Appearance(true)
		h.Frame()
		small := bounds(h, clear).Dy()
		if small >= normal {
			t.Fatal("small clear button not scaled")
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if !d.open {
			t.Fatal("restyle lost trigger focus")
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		d.Size(float32(math.NaN())).Size(float32(math.Inf(1))).Size(-1)
		h.Frame()
		if bounds(h, clear).Dy() != small {
			t.Fatal("invalid size accepted")
		}
		d.Size(0)
		h.Frame()
		if bounds(h, clear).Dy() != normal {
			t.Fatal("default not restored")
		}
		click(t, h, clear)
		h.Frame()
		if a, _ := d.Value(); !a.IsZero() || d.open {
			t.Fatal("clear after restyle")
		}
	}
}
