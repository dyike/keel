package kit

import (
	"math"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SliderScale controls the mapping between values and track position.
type SliderScale uint8

const (
	SliderLinear SliderScale = iota
	SliderLogarithmic
)

// SliderView picks a number in a range by dragging or with the keyboard:
// ← ↓ and → ↑ step, PageUp / PageDown move ten steps, Home / End jump to the ends.
type SliderView struct {
	appearance            func(*SliderAppearance)
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
	scale                 SliderScale
	onRelease             func(float64)
	onRangeRelease        func(float64, float64)
	dragging              bool
	releaseKey            string
	releaseUpper          bool
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

// Scale selects a mapping. Logarithmic ranges require 0 < min < max;
// otherwise the mapping falls back to linear until valid bounds are supplied.
func (v *SliderView) Scale(scale SliderScale) *SliderView {
	if scale <= SliderLogarithmic {
		v.scale = scale
	}
	return v
}

// OnRelease runs once on pointer release or a navigation key's release.
// Canceled/disabled interactions and programmatic changes do not call it.
func (v *SliderView) OnRelease(fn func(float64)) *SliderView { v.onRelease = fn; return v }
func (v *SliderView) OnRangeRelease(fn func(float64, float64)) *SliderView {
	v.onRangeRelease = fn
	return v
}
func (v *SliderView) release() {
	if v.disabled {
		return
	}
	if v.paired {
		if v.onRangeRelease != nil {
			v.onRangeRelease(v.value, v.upper)
		}
	} else if v.onRelease != nil {
		v.onRelease(v.value)
	}
}

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
func (v *SliderView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.dragging = false
		v.releaseKey = ""
	}
}
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
func (v *SliderView) logarithmic() bool {
	return v.scale == SliderLogarithmic && v.min > 0 && v.max > v.min
}
func logDistance(a, b float64) float64 {
	ratio := (b - a) / a
	if math.IsInf(ratio, 0) {
		return math.Log(b) - math.Log(a)
	}
	return math.Log1p(ratio)
}
func (v *SliderView) fraction64(x float64) float64 {
	if v.max <= v.min || x <= v.min {
		return 0
	}
	if x >= v.max {
		return 1
	}
	if v.logarithmic() {
		return logDistance(v.min, x) / logDistance(v.min, v.max)
	}
	return (x - v.min) / (v.max - v.min)
}
func (v *SliderView) fraction(x float64) float32 { return float32(v.fraction64(x)) }
func (v *SliderView) atFraction(f float64) float64 {
	if f <= 0 {
		return v.min
	}
	if f >= 1 {
		return v.max
	}
	if v.logarithmic() {
		d := logDistance(v.min, v.max) * f
		if d < 1 {
			return min(v.max, max(v.min, v.min+v.min*math.Expm1(d)))
		}
		return min(v.max, max(v.min, math.Exp(math.Log(v.min)+d)))
	}
	return v.min + f*(v.max-v.min)
}
func (v *SliderView) advance(x, steps float64) float64 {
	if v.logarithmic() && v.step == 0 {
		return v.atFraction(v.fraction64(x) + steps/100)
	}
	return x + steps*v.keyStep()
}
func (v *SliderView) Render(cx *el.Context) el.Element {
	id := autoID("slider", v)
	if !cx.Enabled(id) {
		v.dragging = false
		v.releaseKey = ""
	}
	if v.releaseKey != "" {
		focusID := id
		if v.paired {
			focusID = v.endpointID(v.releaseUpper)
		}
		if !cx.Focused(focusID) {
			v.releaseKey = ""
		}
	}
	lo, hi := float32(0), v.fraction(v.value)
	if v.paired {
		lo, hi = v.fraction(v.value), v.fraction(v.upper)
	}
	appearance := v.resolveAppearance()
	thumbSize := appearance.ThumbSize
	crossSize := max(float32(20), thumbSize+4, appearance.TrackSize)
	thumbInset := (crossSize - thumbSize) / 2
	fill := appearance.FillColor
	bar := el.Div().Rounded(appearance.TrackRadius).Bg(appearance.TrackColor)
	if v.vertical > 0 {
		bar.W(el.Dp(appearance.TrackSize)).HFull().Child(el.Div().H(el.Frac(1-hi)).NoShrink(), el.Div().W(el.Dp(appearance.TrackSize)).H(el.Frac(hi-lo)).NoShrink().Bg(fill))
	} else {
		bar.H(el.Dp(appearance.TrackSize)).WFull().Row().Child(el.Div().W(el.Frac(lo)).NoShrink(), el.Div().H(el.Dp(appearance.TrackSize)).W(el.Frac(hi-lo)).NoShrink().Bg(fill))
	}
	keyHandler := func(upper bool) func(el.KeyEvent) bool {
		return func(e el.KeyEvent) bool {
			if v.releaseKey != "" && e.State == el.KeyRelease && e.Name == v.releaseKey && upper == v.releaseUpper {
				v.releaseKey = ""
				v.release()
				return true
			}
			if e.Modifiers != 0 {
				return false
			}
			x := v.value
			if upper {
				x = v.upper
			}
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameDownArrow:
				x = v.advance(x, -1)
			case key.NameRightArrow, key.NameUpArrow:
				x = v.advance(x, 1)
			case key.NamePageDown:
				x = v.advance(x, -10)
			case key.NamePageUp:
				x = v.advance(x, 10)
			case key.NameHome:
				x = v.min
			case key.NameEnd:
				x = v.max
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.releaseKey, v.releaseUpper = e.Name, upper
				v.setEndpoint(x, upper)
			}
			return true
		}
	}
	knob := func(x float64, upper bool) el.Element {
		frac := v.fraction(x)
		thumb := el.Div().Size(el.Dp(thumbSize)).NoShrink().Rounded(appearance.ThumbRadius).Bg(appearance.ThumbColor).Border(appearance.ThumbBorderWidth, appearance.ThumbBorderColor)
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
			return layer.Left(thumbInset).Top(0).Bottom(thumbSize).Child(el.Div().H(el.Frac(1-frac)).NoShrink(), thumb)
		}
		return layer.Top(thumbInset).Left(0).Right(thumbSize).Row().Child(el.Div().W(el.Frac(frac)).NoShrink(), thumb)
	}
	track := el.Div().ID(autoID("slider", v)).Disabled(v.disabled).Rounded(theme.RadiusFull).Child(bar, knob(v.value, false))
	if v.vertical > 0 {
		track.W(el.Dp(crossSize)).H(el.Dp(max(v.vertical, thumbSize))).Py(thumbSize / 2).Items(el.Center)
	} else {
		track.H(el.Dp(crossSize)).MinW(el.Dp(thumbSize)).Px(thumbSize / 2).Justify(el.Center)
	}
	if v.paired {
		track.Child(knob(v.upper, true))
	} else {
		track.Role("slider").Name(v.a11y()).Value(strconv.FormatFloat(v.value, 'f', -1, 64)).Focusable(true).OnKey(keyHandler(false)).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	}
	track.OnDrag(func(e el.DragEvent) {
		if e.Canceled {
			v.dragging = false
			return
		}
		pos, size := e.X-thumbSize/2, e.W-thumbSize
		if v.vertical > 0 {
			pos, size = e.H-thumbSize/2-e.Y, e.H-thumbSize
		}
		if size <= 0 {
			v.dragging = false
			return
		}
		f := float64(max(0, min(1, pos/size)))
		x := v.atFraction(f)
		if e.Kind == el.DragStart {
			v.dragging = true
			v.releaseKey = ""
		}
		if e.Kind == el.DragStart && v.paired {
			a, b := math.Abs(f-v.fraction64(v.value)), math.Abs(f-v.fraction64(v.upper))
			v.dragUpper = b < a || b == a && x >= v.upper
			cx.Focus(v.endpointID(v.dragUpper))
		}
		v.setEndpoint(x, v.paired && v.dragUpper)
		if e.Kind == el.DragEnd && v.dragging {
			v.dragging = false
			v.release()
		}
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
