package kit

import (
	"math"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestToggleSegmentedSpacing(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, width := range []int{110, 400} {
			v := ToggleGroup("Alpha", "Beta", "Gamma").Segmented(true).Variant(ToggleOutline)
			h := renderView(v, width, scale)
			a, b := bounds(h, "Alpha"), bounds(h, "Beta")
			if a.Empty() || b.Empty() || a.Max.X != b.Min.X {
				t.Fatalf("scale %d width %d: %v %v", scale, width, a, b)
			}
			v.Gap(8)
			h.Frame()
			a, b = bounds(h, "Alpha"), bounds(h, "Beta")
			if b.Min.X-a.Max.X != 8*scale {
				t.Fatalf("gap %v %v", a, b)
			}
			v.Gap(float32(math.NaN())).Gap(-1)
			h.Frame()
			if b != bounds(h, "Beta") {
				t.Fatal("invalid gap changed geometry")
			}
			v.ResetGap()
			h.Frame()
			a, b = bounds(h, "Alpha"), bounds(h, "Beta")
			if a.Max.X != b.Min.X {
				t.Fatal("reset did not reconnect")
			}
		}
	}
}

func TestToggleSegmentedAncestorDisabled(t *testing.T) {
	v := ToggleGroup("A", "B").Segmented(true).Multiple()
	disabled := true
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 300, 1)
	click(t, h, "A")
	if len(v.Value()) != 0 {
		t.Fatal("ancestor disabled ignored")
	}
	disabled = false
	h.Frame()
	click(t, h, "A")
	if len(v.Value()) != 1 {
		t.Fatal("not restored")
	}
}
