package kit

import (
	"gioui.org/io/key"
	"math"
	"testing"
)

func TestCheckboxSizesPreserveMixedAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		c := Checkbox("Accept", false).OnChange(func(bool) { calls++ })
		c.SetMixed(true)
		h := renderView(c, 220, scale)
		initial := bounds(h, "Accept")
		c.Size(36).TextSize(24)
		h.Frame()
		large := bounds(h, "Accept")
		if large.Dx() <= initial.Dx() || large.Dy() <= initial.Dy() {
			t.Fatal("size ignored", initial, large)
		}
		n, _ := node(h, "Accept")
		if n.Desc.Description != "checkbox:mixed" {
			t.Fatal("mixed semantics", n.Desc)
		}
		click(t, h, "Accept")
		if !c.Value() || c.mixed || calls != 1 {
			t.Fatal("mixed click")
		}
		c.Size(12).TextSize(12)
		h.Frame()
		h.Key(key.NameSpace, 0)
		h.Frame()
		if c.Value() || calls != 2 {
			t.Fatal("size change lost focus")
		}
		c.Size(float32(math.NaN())).TextSize(float32(math.Inf(1)))
		if c.size != 12 || c.textSize != 12 {
			t.Fatal("invalid values")
		}
		c.Size(0).TextSize(0)
		h.Frame()
		if bounds(h, "Accept").Size() != initial.Size() {
			t.Fatal("defaults not restored")
		}
		c.SetDisabled(true)
		h.Frame()
		click(t, h, "Accept")
		if calls != 2 {
			t.Fatal("disabled changed")
		}
	}
}
