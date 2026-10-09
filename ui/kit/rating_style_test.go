package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"strconv"
	"testing"
)

func TestRatingSizesAndFilledStarClick(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		disabled := false
		v := Rating("Score", 5).OnChange(func(int) { calls++ })
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 400, scale)
		last := 0
		for _, size := range []float32{12, 18, 22, 32} {
			v.Size(size)
			h.Frame()
			n, ok := semanticNode(h, "slider:0/5")
			if !ok || n.Desc.Bounds.Dy() <= last {
				t.Fatal("size progression", size, n)
			}
			last = n.Desc.Bounds.Dy()
		}
		clickStar := func(i int) {
			t.Helper()
			n, ok := semanticNode(h, "slider:"+strconv.Itoa(v.Value())+"/5")
			if !ok {
				t.Fatal("rating missing")
			}
			h.Click(float32(n.Desc.Bounds.Min.X)+float32((i-1)*34+16)*float32(scale), float32(n.Desc.Bounds.Min.Y)+16*float32(scale))
		}
		v.SetValue(4)
		h.Frame()
		clickStar(2)
		if v.Value() != 1 || calls != 1 {
			t.Fatal("filled star should choose preceding value", v.Value(), calls)
		}
		clickStar(1)
		if v.Value() != 0 || calls != 2 {
			t.Fatal("cannot clear")
		}
		clickStar(5)
		if v.Value() != 5 || calls != 3 {
			t.Fatal("unfilled star")
		}
		h.Key(key.NameLeftArrow, 0)
		if v.Value() != 4 || calls != 4 {
			t.Fatal("keyboard step")
		}
		disabled = true
		h.Frame()
		clickStar(1)
		h.Key(key.NameHome, 0)
		if v.Value() != 4 || calls != 4 {
			t.Fatal("ancestor disabled")
		}
		disabled = false
		v.ReadOnly()
		h.Frame()
		clickStar(1)
		if v.Value() != 4 || calls != 4 {
			t.Fatal("read only")
		}
	}
}
