package kit

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ProgressCircleView shows determinate or indeterminate progress with optional
// centered content. It is a display, not a focusable control.
type ProgressCircleView struct {
	label         string
	value         float32
	size          float32
	indeterminate bool
	color         *color.NRGBA
	child         el.View
}

func ProgressCircle(label string) *ProgressCircleView {
	return &ProgressCircleView{label: label, size: 48}
}
func (v *ProgressCircleView) Value() float32 { return v.value }

// SetValue clamps the fraction to 0..1 and leaves indeterminate mode. NaN is zero.
func (v *ProgressCircleView) SetValue(value float32) {
	if math.IsNaN(float64(value)) {
		value = 0
	}
	v.value, v.indeterminate = min(max(value, 0), 1), false
}
func (v *ProgressCircleView) SetIndeterminate(on bool) { v.indeterminate = on }
func (v *ProgressCircleView) SetLabel(label string)    { v.label = label }
func (v *ProgressCircleView) Size(dp float32) *ProgressCircleView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}

// Color overrides the foreground; without an override it follows the theme.
func (v *ProgressCircleView) Color(c color.NRGBA) *ProgressCircleView { v.color = &c; return v }

// Child places a view in the center. Keep it compact; the ring bounds clip it.
func (v *ProgressCircleView) Child(child el.View) *ProgressCircleView { v.child = child; return v }
func (v *ProgressCircleView) Render(cx *el.Context) el.Element {
	value := strconv.Itoa(int(v.value*100+.5)) + "%"
	fraction, phase := v.value, float32(0)
	if v.indeterminate {
		value, fraction = "indeterminate", .25
		if !el.ReducedMotion() {
			phase = float32(cx.Now().UnixNano()%int64(time.Second)) / float32(time.Second)
			cx.Animating()
		}
	}
	foreground, track := theme.Primary, theme.Subtle
	if v.color != nil {
		foreground = *v.color
	}
	ring := el.Widget(core.Func(func(gtx core.C) core.D {
		size := gtx.Constraints.Min
		diameter := min(size.X, size.Y)
		if diameter <= 0 {
			return core.D{Size: size}
		}
		defer op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops).Pop()
		gtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		background := material.ProgressCircle(theme.Material, 1)
		background.Color = track
		background.Layout(gtx)
		if fraction > 0 {
			defer op.Affine(f32.Affine2D{}.Rotate(f32.Pt(float32(diameter)/2, float32(diameter)/2), phase*2*math.Pi)).Push(gtx.Ops).Pop()
			progress := material.ProgressCircle(theme.Material, fraction)
			progress.Color = foreground
			progress.Layout(gtx)
		}
		return core.D{Size: size}
	})).Absolute().Top(0).Left(0).WFull().HFull()
	box := el.Div().Role("progressbar").Name(v.label).Value(value).
		Size(el.Dp(v.size)).MaxW(el.Full).MaxH(el.Full).Center().Child(ring)
	box.Decorate(func(gtx core.C, draw func()) {
		defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
		draw()
	})
	if v.child != nil {
		box.Child(el.Div().WFull().HFull().P(theme.SpaceSm).Center().Child(v.child.Render(cx)))
	}
	return box
}
