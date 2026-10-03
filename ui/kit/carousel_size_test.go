package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestCarouselFixedItemSizes(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			car := Carousel(text("one"), text("two"), text("three"), text("four")).ItemsPerView(2).ItemBasis(0, .25).ItemSize(0, 320).ItemSize(1, 80).Height(240).Gap(8).Vertical(vertical).Loop(false).OnChange(func(int) { calls++ })
			content := car.Content()
			width := float32(240)
			var cx *el.Context
			h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(width)).Child(content.Render(c)) }), 300, scale)
			settle := func() {
				for i := 0; i < 5; i++ {
					h.Frame()
				}
			}
			state := func() (float32, float32, float32) {
				id := autoID("carousel", car) + "/stage"
				if vertical {
					return cx.ScrollState(id)
				}
				return cx.ScrollStateX(id)
			}
			settle()
			off, view, total := state()
			if off != 0 || view != 240 || total != 656 {
				t.Fatal("fixed overrides fraction", off, view, total)
			}
			car.Next()
			settle()
			off, _, _ = state()
			if off != 328 || calls != 1 {
				t.Fatal("oversized item prefix", off, calls)
			}
			width = 180
			car.Height(180)
			settle()
			off, view, total = state()
			if off != 328 || view != 180 || total != 596 || calls != 1 {
				t.Fatal("resize changed fixed size", off, view, total, calls)
			}
			car.ItemSize(0, 0)
			settle()
			off, _, total = state()
			if off != 47 || total != 315 || calls != 1 {
				t.Fatal("restored item basis", off, total, calls)
			}
			car.ItemSize(1, float32(math.NaN())).ItemSize(1, float32(math.Inf(1))).ItemSize(1, -1).ItemSize(-1, 42).ItemSize(99, 42)
			if len(car.itemSizes) != 1 || car.itemSizes[1] != 80 {
				t.Fatal("invalid sizes")
			}
			car.ItemSize(1, 0).ItemBasis(0, 0).ItemsPerView(1)
			settle()
			if bounds(h, "two").Empty() || !bounds(h, "one").Empty() || calls != 1 {
				t.Fatal("single mode restore")
			}
		}
	}
}
