package kit

import (
	"math"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestProgressCircleValueAndConstraints(t *testing.T) {
	p := ProgressCircle("Import").Size(64).Child(viewFunc(func(*el.Context) el.Element { return el.Text("42%") }))
	for _, tc := range []struct{ input, want float32 }{{-.5, 0}, {.42, .42}, {2, 1}, {float32(math.NaN()), 0}, {float32(math.Inf(1)), 1}, {float32(math.Inf(-1)), 0}} {
		p.SetIndeterminate(true)
		p.SetValue(tc.input)
		if p.Value() != tc.want {
			t.Fatalf("value %v want %v", p.Value(), tc.want)
		}
		h := renderView(p, 100, 1)
		if _, ok := semanticNode(h, "progressbar:indeterminate"); ok {
			t.Fatal("SetValue did not restore determinate state")
		}
	}
	p.SetValue(.42)
	for _, scale := range []int{1, 2} {
		for _, width := range []int{20, 100} {
			h := renderView(p, width, scale)
			h.Frame()
			n, ok := semanticNode(h, "progressbar:42%")
			if !ok {
				t.Fatal("missing progress value")
			}
			if n.Desc.Bounds.Dx() > width*scale || n.Desc.Bounds.Dy() > 64*scale {
				t.Fatalf("overflow: %v", n.Desc.Bounds)
			}
		}
	}
	p.SetIndeterminate(true)
	h := renderView(p, 100, 1)
	if _, ok := semanticNode(h, "progressbar:indeterminate"); !ok {
		t.Fatal("missing indeterminate value")
	}
	p.SetIndeterminate(false)
	p.SetLabel("Done")
	p.Child(nil)
	h.Frame()
	if n, ok := semanticNode(h, "progressbar:42%"); !ok || n.Desc.Label != "Done" {
		t.Fatal("lost value or label")
	}
}
