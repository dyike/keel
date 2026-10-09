package kit

import (
	"math"
	"strconv"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
)

func TestSliderLogarithmicPointerAndKeys(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		v := Slider("Log", 1, 1000).Scale(SliderLogarithmic)
		if vertical {
			v.Vertical(216)
		}
		var released []float64
		v.OnRelease(func(x float64) { released = append(released, x) })
		h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(216)).Child(v.Render(cx)) })
		n, _ := semanticNode(h, "slider:1")
		b := n.Desc.Bounds
		x, y := float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2)
		h.Click(x, y)
		if math.Abs(v.Value()-math.Sqrt(1000)) > 1e-8 || len(released) != 1 || released[0] != v.Value() {
			t.Fatalf("log midpoint/release: %g %v", v.Value(), released)
		}
		h.Key(key.NameRightArrow, 0)
		if math.Abs(v.Value()-math.Pow(1000, .51)) > 1e-8 || len(released) != 2 {
			t.Fatalf("log key: %g %v", v.Value(), released)
		}
		v.Step(1)
		before := v.Value()
		h.Key(key.NameUpArrow, 0)
		if v.Value() != before+1 {
			t.Fatal("explicit numeric step ignored")
		}
		h.Key(key.NameEnd, 0)
		if v.Value() != 1000 {
			t.Fatal("max not exact")
		}
		h.Key(key.NameHome, 0)
		if v.Value() != 1 {
			t.Fatal("min not exact")
		}
	}
}

func TestRangeSliderLogNearestAndRelease(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		v := RangeSlider("Log range", 1, 1000).Scale(SliderLogarithmic)
		if vertical {
			v.Vertical(216)
		}
		v.SetValues(1, 1000)
		changes, releases := 0, 0
		var result [2]float64
		v.OnRangeChange(func(float64, float64) { changes++ }).OnRangeRelease(func(a, b float64) { releases++; result = [2]float64{a, b} })
		h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(216)).Child(v.Render(cx)) })
		upper := bounds(h, "上限 Log range")
		// 60% of the track is nearer the upper thumb, despite its smaller numeric value.
		if vertical {
			h.Click(float32(upper.Min.X+8), float32(upper.Min.Y+88))
		} else {
			h.Click(128, float32(upper.Min.Y+8))
		}
		lo, hi := v.Values()
		if lo != 1 || math.Abs(hi-math.Pow(1000, .6)) > 1e-4 || releases != 1 || changes == 0 || result != [2]float64{lo, hi} {
			t.Fatalf("nearest/range release: %g %g %v %d", lo, hi, result, releases)
		}
		h.Key(key.NameHome, 0)
		if lo, hi = v.Values(); lo != 1 || hi != 1 || releases != 2 {
			t.Fatal("keyboard range release")
		}
		before := releases
		v.SetValues(10, 100)
		v.SetRange(1, 10000)
		v.SetValue(20)
		if releases != before {
			t.Fatal("programmatic release")
		}
	}
}

func TestSliderReleaseRepeatCancellationAndDisable(t *testing.T) {
	v := Slider("Value", 0, 100)
	releases := 0
	v.OnRelease(func(float64) { releases++ })
	disabled := false
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(216)).Disabled(disabled).Child(v.Render(cx)) })
	clickRole(t, h, "slider", "Value")
	releases = 0
	for range 3 {
		h.Router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
		h.Frame()
	}
	if releases != 0 {
		t.Fatal("key press emitted release")
	}
	h.Router.Queue(key.Event{Name: key.NameRightArrow, State: key.Release})
	h.Frame()
	if releases != 1 {
		t.Fatal("repeat not coalesced at release")
	}
	n, _ := semanticNode(h, "slider:"+strconv.FormatFloat(v.Value(), 'f', -1, 64))
	p := f32.Pt(float32(n.Desc.Bounds.Min.X+60), float32(n.Desc.Bounds.Min.Y+10))
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
	h.Frame()
	if releases != 1 {
		t.Fatal("canceled drag emitted release")
	}
	h.Router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	h.Frame()
	disabled = true
	h.Frame()
	h.Frame()
	h.Router.Queue(key.Event{Name: key.NameRightArrow, State: key.Release})
	h.Frame()
	if releases != 1 {
		t.Fatal("disabled ancestor emitted release")
	}
	disabled = false
	h.Frame()
	v.SetDisabled(true)
	h.Frame()
	h.Click(p.X, p.Y)
	if releases != 1 {
		t.Fatal("disabled control emitted release")
	}
}

func TestSliderLogInvalidAndExtremeRanges(t *testing.T) {
	v := Slider("", 0, 100).Scale(SliderLogarithmic)
	if v.atFraction(.5) != 50 {
		t.Fatal("nonpositive log bounds not linear")
	}
	v.SetRange(1, 1)
	if v.atFraction(.5) != 1 || v.fraction64(1) != 0 {
		t.Fatal("constant range")
	}
	v.SetRange(1e-200, 1e200)
	if x := v.atFraction(.5); !finiteNumber(x) || math.Abs(x-1) > 1e-10 {
		t.Fatalf("extreme log interpolation %g", x)
	}
	if f := v.fraction64(1); math.Abs(f-.5) > 1e-10 {
		t.Fatalf("extreme log fraction %g", f)
	}
	a := 1e100
	b := math.Nextafter(a, math.Inf(1))
	v.SetRange(a, b)
	if !finiteNumber(v.atFraction(.5)) || v.fraction64(b) != 1 {
		t.Fatal("adjacent bounds")
	}
}
