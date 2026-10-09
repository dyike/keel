package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
)

func TestNumberInputSizeAppearancePreserveFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		n := NumberInput("size")
		h := renderView(n, 280, scale)
		loc := locale.Current()
		plus := loc.Name(loc.Increase, "size")
		baseline := bounds(h, plus).Dy()
		clickClass(t, h, "Editor", "size")
		n.Size(48)
		h.Frame()
		large := bounds(h, plus).Dy()
		h.Key(key.NameUpArrow, 0)
		if large <= baseline || n.Value() != 1 {
			t.Fatal("large size or focus", baseline, large, n.Value())
		}
		n.Size(28).Appearance(false)
		h.Frame()
		small := bounds(h, plus).Dy()
		h.Key(key.NameUpArrow, 0)
		if small >= baseline || n.Value() != 2 {
			t.Fatal("small/plain size or focus", small, n.Value())
		}
		n.Size(float32(math.NaN())).Size(-1).Size(float32(math.Inf(1)))
		h.Frame()
		if bounds(h, plus).Dy() != small {
			t.Fatal("invalid size accepted")
		}
		n.Size(0).Appearance(true)
		h.Frame()
		if bounds(h, plus).Dy() != baseline {
			t.Fatal("default not restored")
		}
		click(t, h, plus)
		if n.Value() != 3 {
			t.Fatal("step button after appearance changes")
		}
		n.SetDisabled(true)
		h.Frame()
		click(t, h, plus)
		if n.Value() != 3 {
			t.Fatal("disabled appearance bypass")
		}
	}
}
