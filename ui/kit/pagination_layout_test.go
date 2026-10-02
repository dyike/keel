package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestPaginationNarrowNavigationAndLargeTotals(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Pagination(195, 10)
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(224)).Child(v.Render(cx)) }), 224, scale)
		for _, name := range []string{"1", "2", "3", "4", "20", "下一页"} {
			b := bounds(h, name)
			if b.Empty() || b.Min.X < 0 || b.Max.X > 224*scale {
				t.Fatalf("scale %d: %s outside pagination: %v", scale, name, b)
			}
		}
		click(t, h, "20")
		h.Frame()
		if v.Value() != 20 {
			t.Fatal(v.Value())
		}
	}
	maxInt := int(^uint(0) >> 1)
	v := Pagination(maxInt, 10)
	if v.Pages() != maxInt/10+1 {
		t.Fatal(v.Pages())
	}
	v.SetValue(v.Pages())
	start, end := v.Bounds()
	if start < 0 || end != maxInt || start >= end {
		t.Fatal(start, end)
	}
}
