package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestCarouselFractionalAndMixedBasis(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			car := Carousel(text("one"), text("two"), text("three"), text("four")).Height(240).Basis(.5).ItemBasis(1, .75).ItemBasis(2, .25).Gap(0).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
			var cx *el.Context
			content := car.Content()
			width := float32(240)
			h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(width)).Child(content.Render(c)) }), 300, scale)
			state := func() (float32, float32, float32) {
				id := autoID("carousel", car) + "/stage"
				if vertical {
					return cx.ScrollState(id)
				}
				return cx.ScrollStateX(id)
			}
			settle := func() {
				for i := 0; i < 5; i++ {
					h.Frame()
				}
			}
			settle()
			off, view, total := state()
			if off != 0 || view != 240 || total != 480 {
				t.Fatal("mixed geometry", vertical, scale, off, view, total)
			}
			car.Next()
			settle()
			off, _, _ = state()
			if off != 120 || calls != 1 {
				t.Fatal("first prefix", off, calls)
			}
			car.Next()
			settle()
			off, _, _ = state()
			if off != 240 || calls != 2 {
				t.Fatal("clamped second prefix", off, calls)
			}
			car.SetValue(1)
			car.ItemBasis(0, .25)
			settle()
			off, _, total = state()
			if off != 60 || total != 420 || calls != 2 {
				t.Fatal("dynamic preceding size", off, total, calls)
			}
			car.ItemBasis(0, 0).ItemBasis(1, 0).ItemBasis(2, 0).Basis(2.0 / 3).Gap(8)
			settle()
			off, _, total = state()
			// 240dp viewport: (240+8)*2/3-8, rounded to the current pixel grid.
			cell := float32(math.Round(float64(((float32(240)+8)*car.basis-8)*float32(scale)))) / float32(scale)
			if math.Abs(float64(off-(cell+8))) > .01 || math.Abs(float64(total-(4*cell+24))) > .01 {
				t.Fatal("fraction/spacing rounding", off, total, cell)
			}
			car.Basis(float32(math.NaN())).Basis(-1).Basis(2).ItemBasis(1, float32(math.Inf(1))).ItemBasis(-1, .5).ItemBasis(9, .5)
			if car.basis != float32(2.0/3) || len(car.itemBasis) != 0 {
				t.Fatal("invalid basis")
			}
			car.ItemsPerView(2)
			settle()
			off, _, total = state()
			if off != 124 || total != 488 || calls != 2 {
				t.Fatal("restore equal slots", off, total, calls)
			}
		}
	}
}
