package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"testing"
)

func TestAccordionSizesBordersAndKeyboard(t *testing.T) {
	for _, scale := range []int{1, 2} {
		changes := 0
		v := Accordion().Add("First", text("First body")).Add("Disabled", text("Unavailable")).Add("Last", text("Last body")).OnChange(func([]int) { changes++ })
		v.SetItemDisabled(1, true)
		h := renderView(v, 200, scale)
		previous := 0
		for _, size := range []AccordionSize{AccordionSizeXSmall, AccordionSizeSmall, AccordionSizeMedium, AccordionSizeLarge} {
			v.Size(size)
			h.Frame()
			b := bounds(h, "First")
			if b.Dy() <= previous || b.Dx() > 200*scale {
				t.Fatal("size progression", size, b)
			}
			previous = b.Dy()
		}
		v.SetValue(0)
		h.Frame()
		before := bounds(h, "First")
		v.Bordered(false)
		h.Frame()
		after := bounds(h, "First")
		if after.Min.X >= before.Min.X || !slices.Equal(v.Value(), []int{0}) || changes != 0 {
			t.Fatal("border toggle changed state/layout")
		}
		click(t, h, "First")
		h.Key(key.NameDownArrow, 0)
		h.Key(key.NameSpace, 0)
		if !slices.Equal(v.Value(), []int{2}) || changes != 2 {
			t.Fatal("keyboard skip or toggle", v.Value(), changes)
		}
		v.Heading(2, viewFunc(func(*el.Context) el.Element { return el.Text("Rich title").Bold() }))
		v.Bordered(true).Size(AccordionSizeSmall)
		h.Frame()
		if !shown(h, "Rich title") || !slices.Equal(v.Value(), []int{2}) {
			t.Fatal("heading/state")
		}
		v.SetDisabled(true)
		h.Frame()
		click(t, h, "First")
		if changes != 2 {
			t.Fatal("disabled toggle")
		}
	}
}
