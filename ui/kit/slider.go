package kit

import (
	"math"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// SliderView picks a number in a range by dragging or with the keyboard:
// ← ↓ and → ↑ step, PageUp / PageDown move ten steps, Home / End jump to the ends.
type SliderView struct {
	label                 string
	min, max, step, value float64
	disabled              bool
	onChange              func(float64)
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
	}
	return v
}
func (v *SliderView) Value() float64      { return v.value }
func (v *SliderView) SetValue(x float64)  { v.value = v.snap(x) }
func (v *SliderView) SetDisabled(on bool) { v.disabled = on }
func (v *SliderView) SetRange(a, b float64) {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		a, b = 0, 1
	}
	v.min, v.max = min(a, b), max(a, b)
	v.value = v.snap(v.value)
}

func (v *SliderView) snap(x float64) float64 {
	if math.IsNaN(x) {
		x = v.min
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

func (v *SliderView) Render(cx *el.Context) el.Element {
	frac := float32(0)
	if v.max > v.min {
		frac = float32((v.value - v.min) / (v.max - v.min))
	}
	fill, thumb := theme.Primary, theme.Surface
	if v.disabled {
		fill = theme.Muted
	}
	const thumbSize = 16
	// The track spans the element; the thumb's center sits at frac of the
	// width inside half a thumb of padding, so it never leaves the box.
	bar := el.Div().H(el.Dp(4)).Rounded(2).Bg(theme.Border).Items(el.Start).Child(
		el.Div().H(el.Dp(4)).Rounded(2).Bg(fill).W(el.Frac(frac)),
	)
	// The thumb rides a layer one thumb narrower than the track, so frac of
	// that layer puts its center at frac of the travel.
	knob := el.Div().Absolute().Top(2).Left(0).Right(thumbSize).Row().Child(
		el.Div().W(el.Frac(frac)).NoShrink(),
		el.Div().Size(el.Dp(thumbSize)).NoShrink().Rounded(thumbSize/2).Bg(thumb).Border(2, fill),
	)
	track := el.Div().ID(autoID("slider", v)).Role("slider").Name(v.label).Value(strconv.FormatFloat(v.value, 'f', -1, 64)).
		Disabled(v.disabled).H(el.Dp(20)).Px(thumbSize/2).Justify(el.Center).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).Rounded(10).
		Child(bar, knob).
		OnDrag(func(e el.DragEvent) {
			w := e.W - thumbSize
			if w > 0 {
				v.set(v.min + float64((e.X-thumbSize/2)/w)*(v.max-v.min))
			}
		}).
		OnKey(func(e el.KeyEvent) bool {
			d := 0.0
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameDownArrow:
				d = -v.keyStep()
			case key.NameRightArrow, key.NameUpArrow:
				d = v.keyStep()
			case key.NamePageDown:
				d = -10 * v.keyStep()
			case key.NamePageUp:
				d = 10 * v.keyStep()
			case key.NameHome:
				d = math.Inf(-1)
			case key.NameEnd:
				d = math.Inf(1)
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.set(v.value + d)
			}
			return true
		})
	if !v.disabled {
		track.CursorPointer()
	}
	return labelled(v.label, track, "")
}
