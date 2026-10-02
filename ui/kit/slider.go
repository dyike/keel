package kit

import (
	"math"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SliderView picks a number in a range by dragging or with the keyboard:
// ← ↓ and → ↑ step, PageUp / PageDown move ten steps, Home / End jump to the ends.
type SliderView struct {
	label                 string
	name                  string
	min, max, step, value float64
	disabled              bool
	onChange              func(float64)
	paired                bool
	upper                 float64
	vertical              float32
	dragUpper             bool
	onRangeChange         func(float64, float64)
}

// Slider creates a slider over [min, max]; reversed bounds are swapped and the
// value starts at min. Without Step it moves by 1% of the range.
func Slider(label string, min, max float64) *SliderView {
	v := &SliderView{label: label}
	v.SetRange(min, max)
	v.value = v.min
	return v
}
func (v *SliderView) OnChange(fn func(float64)) *SliderView { v.onChange = fn; return v }

// Step snaps values to min + k·step; 0 means continuous.
func (v *SliderView) Step(s float64) *SliderView {
	if s >= 0 && !math.IsInf(s, 0) && !math.IsNaN(s) {
		v.step = s
		v.value = v.snap(v.value)
		v.upper = max(v.value, v.snap(v.upper))
	}
	return v
}
func (v *SliderView) Value() float64 { return v.value }
func (v *SliderView) SetValue(x float64) {
	v.value = v.snap(x)
	if v.paired {
		v.value = min(v.value, v.upper)
	}
}
func (v *SliderView) SetDisabled(on bool) { v.disabled = on }
func (v *SliderView) SetRange(a, b float64) {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) || math.IsInf(b-a, 0) {
		a, b = 0, 1
	}
	v.min, v.max = min(a, b), max(a, b)
	v.value = v.snap(v.value)
	v.upper = max(v.value, v.snap(v.upper))
}

func (v *SliderView) snap(x float64) float64 {
	if math.IsNaN(x) {
		x = v.min
	}
	if x <= v.min {
		return v.min
	}
	if x >= v.max {
		return v.max
	}
	if v.step > 0 {
		x = v.min + math.Round((x-v.min)/v.step)*v.step
	}
	return min(v.max, max(v.min, x))
}

func (v *SliderView) set(x float64) {
	x = v.snap(x)
	if x == v.value {
		return
	}
	v.value = x
	if v.onChange != nil {
		v.onChange(x)
	}
}

func (v *SliderView) keyStep() float64 {
	if v.step > 0 {
		return v.step
	}
	return (v.max - v.min) / 100
}

// RangeSlider creates a two-ended slider initialized to the full range.
func RangeSlider(label string, min, max float64) *SliderView {
	v := Slider(label, min, max)
	v.paired = true
	v.upper = v.max
	return v
}
func (v *SliderView) Values() (float64, float64) { return v.value, v.upper }

// SetValues sorts and clamps both ends without calling user callbacks.
func (v *SliderView) SetValues(a, b float64) {
	a, b = v.snap(a), v.snap(b)
	v.value, v.upper = min(a, b), max(a, b)
}
func (v *SliderView) OnRangeChange(fn func(float64, float64)) *SliderView {
	v.onRangeChange = fn
	return v
}

// Vertical uses a bottom-to-top track with the given height in dp.
func (v *SliderView) Vertical(height float32) *SliderView {
	if height > 0 && !math.IsInf(float64(height), 0) && !math.IsNaN(float64(height)) {
		v.vertical = max(20, height)
	}
	return v
}
func (v *SliderView) FocusID() string {
	if v.paired {
		return v.endpointID(false)
	}
	return autoID("slider", v)
}
func (v *SliderView) endpointID(upper bool) string {
	if upper {
		return autoID("slider", v) + "/upper"
	}
	return autoID("slider", v) + "/lower"
}
func (v *SliderView) setEndpoint(x float64, upper bool) {
	if v.disabled {
		return
	}
	if !v.paired {
		v.set(x)
		return
	}
	lo, hi := v.value, v.upper
	if upper {
		v.upper = max(v.value, v.snap(x))
	} else {
		v.value = min(v.upper, v.snap(x))
	}
	if lo == v.value && hi == v.upper {
		return
	}
	if v.onRangeChange != nil {
		v.onRangeChange(v.value, v.upper)
	}
}
func (v *SliderView) fraction(x float64) float32 {
	if v.max <= v.min {
		return 0
	}
	return float32((x - v.min) / (v.max - v.min))
}
func (v *SliderView) Render(cx *el.Context) el.Element {
	lo, hi := float32(0), v.fraction(v.value)
	if v.paired {
		lo, hi = v.fraction(v.value), v.fraction(v.upper)
	}
	fill := theme.Primary
	if v.disabled {
		fill = theme.Muted
	}
	bar := el.Div().Rounded(theme.RadiusFull).Bg(theme.Border)
	if v.vertical > 0 {
		bar.W(el.Dp(4)).HFull().Child(el.Div().H(el.Frac(1-hi)).NoShrink(), el.Div().W(el.Dp(4)).H(el.Frac(hi-lo)).NoShrink().Bg(fill))
	} else {
		bar.H(el.Dp(4)).WFull().Row().Child(el.Div().W(el.Frac(lo)).NoShrink(), el.Div().H(el.Dp(4)).W(el.Frac(hi-lo)).NoShrink().Bg(fill))
	}
	keyHandler := func(upper bool) func(el.KeyEvent) bool {
		return func(e el.KeyEvent) bool {
			if e.Modifiers != 0 {
				return false
			}
			x := v.value
			if upper {
				x = v.upper
			}
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameDownArrow:
				x -= v.keyStep()
			case key.NameRightArrow, key.NameUpArrow:
				x += v.keyStep()
			case key.NamePageDown:
				x -= 10 * v.keyStep()
			case key.NamePageUp:
				x += 10 * v.keyStep()
			case key.NameHome:
				x = v.min
			case key.NameEnd:
				x = v.max
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.setEndpoint(x, upper)
			}
			return true
		}
	}
	knob := func(x float64, upper bool) el.Element {
		frac := v.fraction(x)
		thumb := el.Div().Size(el.Dp(16)).NoShrink().Rounded(theme.RadiusLg).Bg(theme.Surface).Border(2, fill)
		if v.paired {
			name := locale.Current().LowerValue
			if upper {
				name = locale.Current().UpperValue
			}
			thumb.ID(v.endpointID(upper)).Role("slider").Name(locale.Current().Name(name, v.a11y())).Value(strconv.FormatFloat(x, 'f', -1, 64)).Focusable(true).
				FocusStyle(func(s *el.Style) { s.BorderColor(theme.Text) }).OnKey(keyHandler(upper))
		}
		layer := el.Div().Absolute()
		if v.vertical > 0 {
			return layer.Left(2).Top(0).Bottom(16).Child(el.Div().H(el.Frac(1-frac)).NoShrink(), thumb)
		}
		return layer.Top(2).Left(0).Right(16).Row().Child(el.Div().W(el.Frac(frac)).NoShrink(), thumb)
	}
	track := el.Div().ID(autoID("slider", v)).Disabled(v.disabled).Rounded(theme.RadiusFull).Child(bar, knob(v.value, false))
	if v.vertical > 0 {
		track.W(el.Dp(20)).H(el.Dp(v.vertical)).Py(theme.SpaceMd).Items(el.Center)
	} else {
		track.H(el.Dp(20)).Px(theme.SpaceMd).Justify(el.Center)
	}
	if v.paired {
		track.Child(knob(v.upper, true))
	} else {
		track.Role("slider").Name(v.a11y()).Value(strconv.FormatFloat(v.value, 'f', -1, 64)).Focusable(true).OnKey(keyHandler(false)).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	}
	track.OnDrag(func(e el.DragEvent) {
		if e.Canceled {
			return
		}
		pos, size := e.X-8, e.W-16
		if v.vertical > 0 {
			pos, size = e.H-8-e.Y, e.H-16
		}
		if size <= 0 {
			return
		}
		x := v.min + float64(max(0, min(1, pos/size)))*(v.max-v.min)
		if e.Kind == el.DragStart && v.paired {
			a, b := math.Abs(x-v.value), math.Abs(x-v.upper)
			v.dragUpper = b < a || b == a && x >= v.upper
			cx.Focus(v.endpointID(v.dragUpper))
		}
		v.setEndpoint(x, v.paired && v.dragUpper)
	})
	if !v.disabled {
		track.CursorPointer()
	}
	return labelled(v.label, track, "")
}

func (v *SliderView) setName(s string) { v.name = s }
func (v *SliderView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
