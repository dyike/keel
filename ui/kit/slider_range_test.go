package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestRangeSliderKeysClampAndIndependentFocus(t *testing.T) {
	calls := 0
	v := RangeSlider("Price", 0, 100).Step(10).OnRangeChange(func(float64, float64) { calls++ })
	v.SetValues(20, 70)
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(216)).Child(v.Render(cx)) })
	click(t, h, "下限 Price")
	h.Key(key.NameEnd, 0)
	if a, b := v.Values(); a != 70 || b != 70 {
		t.Fatalf("crossed upper %v %v", a, b)
	}
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	if a, b := v.Values(); a != 70 || b != 80 {
		t.Fatalf("upper tab %v %v", a, b)
	}
	h.Key(key.NameHome, 0)
	if a, b := v.Values(); a != 70 || b != 70 {
		t.Fatal("upper crossed lower")
	}
	before := calls
	v.SetValues(90, 10)
	if a, b := v.Values(); a != 10 || b != 90 || calls != before {
		t.Fatal("program values")
	}
	v.SetDisabled(true)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	if calls != before {
		t.Fatal("disabled callback")
	}
}

func TestRangeSliderDragNearestAndOverlap(t *testing.T) {
	v := RangeSlider("Price", 0, 100).Step(10)
	v.SetValues(20, 80)
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(216)).Child(v.Render(cx)) })
	b := bounds(h, "上限 Price")
	x, y := center(b)
	h.Drag(x, y, x-200, y)
	if a, b := v.Values(); a != 20 || b != 20 {
		t.Fatalf("drag crossed %v %v", a, b)
	}
	b = bounds(h, "上限 Price")
	x, y = center(b)
	h.Drag(x, y, x+120, y)
	if a, b := v.Values(); a != 20 || b != 80 {
		t.Fatalf("overlap cannot separate %v %v", a, b)
	}
}

func TestVerticalSliderPointerAndBoundary(t *testing.T) {
	v := Slider("Level", 0, 100).Vertical(216).Step(10)
	h := page(v)
	n, _ := semanticNode(h, "slider:0")
	b := n.Desc.Bounds
	x := float32(b.Min.X + b.Dx()/2)
	h.Drag(x, float32(b.Max.Y-8), x, float32(b.Min.Y+8))
	if v.Value() != 100 {
		t.Fatalf("vertical pointer %v bounds %v", v.Value(), b)
	}
	h.Key(key.NameDownArrow, 0)
	if v.Value() != 90 {
		t.Fatal("vertical down")
	}
	v.SetRange(0, 10)
	v.Step(3)
	h.Frame()
	h.Key(key.NameEnd, 0)
	if v.Value() != 10 {
		t.Fatal("nondivisible max unreachable")
	}
	v.SetRange(-math.MaxFloat64, math.MaxFloat64)
	v.SetValue(math.NaN())
	if !finiteNumber(v.Value()) {
		t.Fatal("nonfinite range")
	}
}
