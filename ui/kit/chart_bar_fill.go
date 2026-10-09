package kit

import (
	"image"
	"image/color"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// ChartBarDatum identifies one original data value, including a stacked segment.
type ChartBarDatum struct {
	Series, Index int
	Name, Label   string
	Value         float64
	Stacked       bool
	Color         color.NRGBA
}

// ChartBarGradient interpolates Start to End toward Direction (default Bottom).
// Top/Bottom run vertically; Left/Right run horizontally across each bar's box.
type ChartBarGradient struct {
	Start, End color.NRGBA
	Direction  el.Side
}

// ChartBarFill uses Color unless Gradient is non-nil. Alpha is preserved.
type ChartBarFill struct {
	Color    color.NRGBA
	Gradient *ChartBarGradient
}

// BarFill customizes each bar or stacked segment during drawing. The callback
// must be pure: it may run during measurement as well as painting. Nil restores
// series colors. It does not alter legends, hit testing, values or data tables.
func (v *ChartView) BarFill(fn func(ChartBarDatum) ChartBarFill) *ChartView {
	v.options.barFill = fn
	v.options.barGradient = nil
	return v
}

func (v *ChartView) paintBar(gtx core.C, shape clip.RRect, series, index int, context ChartBarRange) {
	color := v.seriesColor(series)
	if v.options.barGradient != nil {
		s := v.series[series]
		datum := ChartBarDatum{Series: series, Index: index, Name: s.Name, Label: v.labels[index], Value: v.value(s, index), Stacked: v.stacked, Color: color}
		if v.paintGradientStops(gtx, shape, context, v.options.barGradient(datum, context)) {
			return
		}
	}
	if v.options.barFill == nil {
		paint.FillShape(gtx.Ops, color, shape.Op(gtx.Ops))
		return
	}
	s := v.series[series]
	fill := v.options.barFill(ChartBarDatum{Series: series, Index: index, Name: s.Name, Label: v.labels[index], Value: v.value(s, index), Stacked: v.stacked, Color: color})
	if fill.Gradient == nil {
		paint.FillShape(gtx.Ops, fill.Color, shape.Op(gtx.Ops))
		return
	}
	a, b := barGradientPoints(shape.Rect, v.barGradientDirection(fill.Gradient.Direction))
	area := shape.Op(gtx.Ops).Push(gtx.Ops)
	// Keep the rotated clip, but paint the gradient in screen coordinates:
	// Gio's gradient brush does not follow a rotated/reflected paint transform.
	transform := v.barTransform(gtx.Constraints.Max)
	a, b = transform.Transform(a), transform.Transform(b)
	brush := op.Affine(transform.Invert()).Push(gtx.Ops)
	paint.LinearGradientOp{Stop1: a, Stop2: b, Color1: fill.Gradient.Start, Color2: fill.Gradient.End}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	brush.Pop()
	area.Pop()
}
func barGradientPoints(rect image.Rectangle, direction el.Side) (f32.Point, f32.Point) {
	x, y := float32(rect.Min.X+rect.Max.X)/2, float32(rect.Min.Y+rect.Max.Y)/2
	switch direction {
	case el.Top:
		return f32.Pt(x, float32(rect.Max.Y)), f32.Pt(x, float32(rect.Min.Y))
	case el.Left:
		return f32.Pt(float32(rect.Max.X), y), f32.Pt(float32(rect.Min.X), y)
	case el.Right:
		return f32.Pt(float32(rect.Min.X), y), f32.Pt(float32(rect.Max.X), y)
	default:
		return f32.Pt(x, float32(rect.Min.Y)), f32.Pt(x, float32(rect.Max.Y))
	}
}
