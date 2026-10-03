package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestCarouselMultipleItemsGeometryAndSelection(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			car := Carousel(text("first"), text("second"), text("third"), text("fourth")).ItemsPerView(2).Gap(8).Height(160).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
			content := car.Content()
			var cx *el.Context
			width := float32(240)
			h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(width)).Child(content.Render(c)) }), 300, scale)
			for i := 0; i < 5; i++ {
				h.Frame()
			}
			id := autoID("carousel", car) + "/stage"
			state := func() (float32, float32, float32) {
				if vertical {
					return cx.ScrollState(id)
				}
				return cx.ScrollStateX(id)
			}
			off, view, total := state()
			expected := float32(240)
			if vertical {
				expected = 160
			}
			if off != 0 || math.Abs(float64(view-expected)) > 1 || math.Abs(float64(total-(2*expected+8))) > 2 {
				t.Fatal("geometry", vertical, scale, off, view, total)
			}
			car.Next()
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			off, _, _ = state()
			if math.Abs(float64(off-(expected+8)/2)) > 1 || calls != 1 {
				t.Fatal("next alignment", off, calls)
			}
			car.SetValue(3)
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			off, view, total = state()
			if math.Abs(float64(off-(total-view))) > 1 || calls != 1 {
				t.Fatal("end alignment", off, view, total, calls)
			}
			car.SetDisabled(true)
			car.SetValue(0)
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			if bounds(h, "first").Empty() {
				t.Fatal("disabled programmatic selection is not visible")
			}
			car.SetDisabled(false)
			width = 140
			car.Vertical(!vertical)
			vertical = !vertical
			for i := 0; i < 6; i++ {
				h.Frame()
			}
			off, view, total = state()
			expected = 140
			if vertical {
				expected = 160
			}
			if off != 0 || math.Abs(float64(view-expected)) > 1 || math.Abs(float64(total-(2*expected+8))) > 2 {
				t.Fatal("resize/axis", off, view, total)
			}
			car.Gap(float32(math.Inf(1))).Gap(-1).ItemsPerView(0)
			if car.gap != 8 || car.perView != 2 {
				t.Fatal("invalid configuration accepted")
			}
		}
	}
}
