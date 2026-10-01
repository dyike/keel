package widget

import (
	"fmt"
	"image"
	"math"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// SliderView selects a single numeric value by dragging or keyboard.
// Keep it between frames, like Input. Setters do not call OnChange.
type SliderView struct {
	label                 string
	min, max, step, value float32
	disabled, dragging    bool
	pointer               pointer.ID
	width, inset          int
	onChange              func(float32)
}

// Slider creates a horizontal slider. Reversed bounds are swapped; non-finite
// bounds use [0, 1]. Values initially start at min.
func Slider(label string, minValue, maxValue float32) *SliderView {
	s := &SliderView{label: label}
	s.SetRange(minValue, maxValue)
	s.value = s.min
	return s
}
func (s *SliderView) OnChange(fn func(float32)) *SliderView { s.onChange = fn; return s }
func (s *SliderView) Step(v float32) *SliderView {
	if !finite(v) || v < 0 {
		v = 0
	}
	s.step = v
	s.SetValue(s.value)
	return s
}
func (s *SliderView) Value() float32     { return s.value }
func (s *SliderView) SetValue(v float32) { s.value = s.normalize(v) }
func (s *SliderView) SetDisabled(v bool) {
	s.disabled = v
	if v {
		s.dragging = false
	}
}
func (s *SliderView) SetRange(a, b float32) {
	if !finite(a) || !finite(b) || !finite(b-a) {
		a, b = 0, 1
	}
	if a > b {
		a, b = b, a
	}
	s.min, s.max = a, b
	s.SetValue(s.value)
}
func finite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func (s *SliderView) normalize(v float32) float32 {
	if math.IsNaN(float64(v)) || v <= s.min {
		return s.min
	}
	if v >= s.max {
		return s.max
	}
	if s.step > 0 {
		v = s.min + float32(math.Round(float64(v-s.min)/float64(s.step)))*s.step
	}
	return max(s.min, min(s.max, v))
}
func (s *SliderView) change(gtx C, v float32) {
	v = s.normalize(v)
	if v == s.value {
		return
	}
	s.value = v
	gtx.Execute(op.InvalidateCmd{})
	core.Call(gtx, func() {
		if s.onChange != nil {
			s.onChange(v)
		}
	})
}
func (s *SliderView) update(gtx C) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: s, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel},
			key.FocusFilter{Target: s},
			key.Filter{Focus: s, Name: key.NameLeftArrow}, key.Filter{Focus: s, Name: key.NameRightArrow},
			key.Filter{Focus: s, Name: key.NameUpArrow}, key.Filter{Focus: s, Name: key.NameDownArrow},
			key.Filter{Focus: s, Name: key.NameHome}, key.Filter{Focus: s, Name: key.NameEnd},
			key.Filter{Focus: s, Name: key.NamePageUp}, key.Filter{Focus: s, Name: key.NamePageDown})
		if !ok {
			break
		}
		if s.disabled || s.min == s.max {
			continue
		}
		switch e := ev.(type) {
		case pointer.Event:
			switch e.Kind {
			case pointer.Press:
				if e.Source == pointer.Mouse && !e.Buttons.Contain(pointer.ButtonPrimary) {
					continue
				}
				if s.dragging {
					continue
				}
				s.dragging, s.pointer = true, e.PointerID
				gtx.Execute(key.FocusCmd{Tag: s})
				gtx.Execute(pointer.GrabCmd{Tag: s, ID: e.PointerID})
			case pointer.Cancel:
				s.dragging = false
				continue
			}
			if s.dragging && e.PointerID == s.pointer {
				if e.Kind != pointer.Release || e.Source != pointer.Mouse || e.Buttons == 0 {
					if s.width > 2*s.inset {
						ratio := max(float32(0), min(float32(1), (e.Position.X-float32(s.inset))/float32(s.width-2*s.inset)))
						s.change(gtx, s.min+ratio*(s.max-s.min))
					}
				}
				if e.Kind == pointer.Release {
					s.dragging = false
				}
			}
		case key.Event:
			if e.State != key.Press {
				continue
			}
			if e.Name == key.NameHome {
				s.change(gtx, s.min)
				continue
			}
			if e.Name == key.NameEnd {
				s.change(gtx, s.max)
				continue
			}
			count := float64(1)
			if e.Name == key.NamePageUp || e.Name == key.NamePageDown {
				count = 10
			}
			up := e.Name == key.NameRightArrow || e.Name == key.NameUpArrow || e.Name == key.NamePageUp
			step := s.step
			if step == 0 {
				step = (s.max - s.min) / 100
			}
			if step == 0 {
				continue
			}
			n := float64(s.value-s.min) / float64(step)
			if up {
				n = math.Floor(n+1e-5) + count
			} else {
				n = math.Ceil(n-1e-5) - count
			}
			s.change(gtx, s.min+float32(n)*step)
		}
	}
}
func (s *SliderView) Layout(gtx C) D {
	if s.disabled || s.min == s.max {
		gtx = gtx.Disabled()
	}
	s.update(gtx)
	return core.Semantic(gtx, func(gtx C) D {
		return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
			giolayout.Rigid(func(gtx C) D {
				return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
					giolayout.Flexed(1, Muted(s.label).Layout),
					giolayout.Rigid(func(gtx C) D { return Muted(fmt.Sprintf("%g", s.value)).Layout(gtx) }))
			}),
			giolayout.Rigid(s.track))
	}, core.Role("slider", fmt.Sprintf("%g", s.value)), semantic.LabelOp(s.label), semantic.EnabledOp(gtx.Enabled()))
}
func (s *SliderView) track(gtx C) D {
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(32)))
	s.width, s.inset = size.X, min(gtx.Dp(10), size.X/2)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, s)
	if gtx.Enabled() {
		pointer.CursorPointer.Add(gtx.Ops)
	}
	y, r := size.Y/2, gtx.Dp(8)
	left, right := s.inset, size.X-s.inset
	x := left
	if s.max > s.min {
		x += int(float32(right-left)*(s.value-s.min)/(s.max-s.min) + .5)
	}
	col := theme.Primary
	if !gtx.Enabled() {
		col = theme.Muted
	}
	paint.FillShape(gtx.Ops, theme.Border, clip.UniformRRect(image.Rect(left, y-gtx.Dp(2), right, y+gtx.Dp(2)), gtx.Dp(2)).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rect(left, y-gtx.Dp(2), x, y+gtx.Dp(2)), gtx.Dp(2)).Op(gtx.Ops))
	if gtx.Focused(s) && gtx.Enabled() {
		paint.FillShape(gtx.Ops, theme.Highlight, clip.Ellipse(image.Rect(x-r-gtx.Dp(3), y-r-gtx.Dp(3), x+r+gtx.Dp(3), y+r+gtx.Dp(3))).Op(gtx.Ops))
	}
	paint.FillShape(gtx.Ops, col, clip.Ellipse(image.Rect(x-r, y-r, x+r, y+r)).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, theme.Surface, clip.Ellipse(image.Rect(x-r+gtx.Dp(2), y-r+gtx.Dp(2), x+r-gtx.Dp(2), y+r-gtx.Dp(2))).Op(gtx.Ops))
	return D{Size: size}
}
