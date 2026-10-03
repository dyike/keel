package kit

import (
	"math"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestResizableMaximumInteraction(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			pane := func(name string) el.View {
				return viewFunc(func(*el.Context) el.Element { return el.Div().Grow().Role("group").Name(name) })
			}
			v := Resizable(pane("first"), pane("second")).Min(50, 60).Max(200, 220).OnChange(func(float32) { calls++ })
			if vertical {
				v.Vertical()
			}
			length := float32(400)
			h := renderView(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().W(el.Dp(length)).H(el.Dp(length)).Items(el.Stretch).Child(v.Render(cx))
			}), 600, scale)
			h.Frame()
			if v.Value() != 200 {
				t.Fatal("first maximum", v.Value())
			}
			n, ok := semanticNode(h, "separator:200")
			if !ok {
				t.Fatal("separator missing")
			}
			b := n.Desc.Bounds
			x, y := float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2)
			dx, dy := float32(-150*scale), float32(0)
			if vertical {
				dx, dy = dy, dx
			}
			h.Drag(x, y, x+dx, y+dy)
			h.Frame()
			if v.Value() != 174 {
				t.Fatal("second maximum not enforced", v.Value())
			}
			h.Key(key.NameEnd, 0)
			if v.Value() != 200 {
				t.Fatal("keyboard max", v.Value())
			}
			before := calls
			v.SetValue(float32(math.NaN()))
			v.SetValue(float32(math.Inf(1)))
			v.Max(float32(math.NaN()), -1)
			if v.Value() != 200 {
				t.Fatal("invalid value accepted")
			}
			length = 600
			h.Frame()
			h.Frame()
			second := bounds(h, "second")
			size := second.Dx()
			if vertical {
				size = second.Dy()
			}
			if size != 220*scale || v.Value() != 200 || calls != before {
				t.Fatal("maxima should leave spare space", second, v.Value(), calls, before)
			}
			length = 100
			h.Frame()
			h.Frame()
			if v.Value() != 34 {
				t.Fatal("narrow container", v.Value())
			}
			v.Max(0, 0)
			length = 400
			h.Frame()
			h.Frame()
			v.SetValue(300)
			h.Frame()
			if v.Value() != 300 {
				t.Fatal("max reset", v.Value())
			}
		}
	}
}

func TestResizableContradictoryAndPrelayoutLimits(t *testing.T) {
	v := Resizable(nil, nil).Min(80, 90).Max(50, 40)
	v.SetValue(500)
	if v.Value() != 80 {
		t.Fatal("minimum should take precedence", v.Value())
	}
	v.Min(float32(math.NaN()), float32(math.Inf(1)))
	if v.min1 != 80 || v.min2 != 90 {
		t.Fatal("nonfinite minimum accepted")
	}
}
