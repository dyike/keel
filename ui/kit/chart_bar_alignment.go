package kit

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image"
)

// BarAlignment selects the side from which positive bars grow.
type BarAlignment uint8

const (
	BarAlignmentBottom BarAlignment = iota
	BarAlignmentTop
	BarAlignmentLeft
	BarAlignmentRight
)

func (v *ChartView) BarAlignment(side BarAlignment) *ChartView {
	if side <= BarAlignmentRight && v.options.barAlignment != side {
		v.options.barAlignment = side
		v.hover = -1
		v.hoverMotion = chartHoverMotion{}
	}
	return v
}
func (v *ChartView) horizontalBars() bool {
	return v.options.barAlignment == BarAlignmentLeft || v.options.barAlignment == BarAlignmentRight
}
func (v *ChartView) barValueFraction(value, lo, hi float64) float32 {
	f := float32(axisFraction(value, lo, hi))
	if v.options.barAlignment == BarAlignmentRight {
		return 1 - f
	}
	return f
}
func (v *ChartView) drawOrientedBars(gtx core.C, lo, hi float64, ticks []float64) core.D {
	size := gtx.Constraints.Max
	px := gtx.Metric.PxPerDp
	if px <= 0 {
		px = 1
	}
	if width := float32(size.X) / px; gtx.Enabled() && width != v.options.orientedWidth {
		v.options.orientedWidth = width
		gtx.Execute(op.InvalidateCmd{})
	}
	canonical := size
	if v.horizontalBars() {
		canonical = image.Pt(size.Y, size.X)
	}
	transform := v.barTransform(canonical)
	defer op.Affine(transform).Push(gtx.Ops).Pop()
	gtx.Constraints = layout.Exact(canonical)
	v.draw(gtx, lo, hi, ticks)
	return core.D{Size: size}
}

// barTransform maps the canonical bottom-up plot into screen coordinates.
func (v *ChartView) barTransform(canonical image.Point) f32.Affine2D {
	switch v.options.barAlignment {
	case BarAlignmentTop:
		return f32.NewAffine2D(1, 0, 0, 0, -1, float32(canonical.Y))
	case BarAlignmentLeft:
		return f32.NewAffine2D(0, -1, float32(canonical.Y), 1, 0, 0)
	case BarAlignmentRight:
		return f32.NewAffine2D(0, 1, 0, 1, 0, 0)
	default:
		return f32.Affine2D{}
	}
}

func (v *ChartView) orientedBars(cx *el.Context, root *el.DivEl) el.Element {
	ticks := v.axisTicks()
	lo, hi := ticks[0], ticks[len(ticks)-1]
	gutter := v.chartGutter()
	horizontal := v.horizontalBars()
	if horizontal && v.options.gutter == nil {
		gutter.Left = 80
		gutter.Bottom = 24
	}
	axis := el.Div().W(el.Dp(gutter.Left)).H(el.Dp(v.height)).NoShrink()
	plot := el.Div().Grow().W(el.Dp(0)).H(el.Dp(v.height)).Items(el.Stretch).Child(el.Widget(core.Func(func(gtx core.C) core.D { return v.drawOrientedBars(gtx, lo, hi, ticks) })).HFull().WFull())
	if horizontal {
		for _, index := range v.xTickIndices(v.height) {
			top := (float32(index)+.5)*v.height/float32(max(1, v.categoryCount())) - 8
			axis.Child(el.Div().Absolute().Top(top).Right(8).MaxW(el.Dp(max(0, gutter.Left-8))).Child(el.Text(v.labels[index]).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1)))
		}
	} else {
		for _, tick := range ticks {
			top := float32(axisFraction(tick, lo, hi))*v.height - 8
			label := el.Div().Absolute().Top(top).Child(el.Text(v.format(tick)).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1))
			if v.options.yLabelsInside {
				plot.Child(label.Left(6))
			} else {
				axis.Child(label.Right(8).MaxW(el.Dp(max(0, gutter.Left-8))))
			}
		}
	}
	width := v.options.orientedWidth
	if width <= 0 {
		w, _ := cx.ViewportSize()
		width = max(1, w-gutter.Left-gutter.Right-24)
	}
	bottom := el.Div().Grow().W(el.Dp(0)).H(el.Dp(gutter.Bottom))
	if horizontal {
		for _, tick := range ticks {
			left := v.barValueFraction(tick, lo, hi) * width
			label := el.Div().Absolute().Left(max(0, min(width-48, left-24))).W(el.Dp(48)).Items(el.Center).Child(el.Text(v.format(tick)).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1))
			if v.options.yLabelsInside {
				plot.Child(label.Top(max(0, v.height-18)))
			} else {
				bottom.Child(label)
			}
		}
	} else {
		for _, index := range v.xTickIndices(width) {
			left := (float32(index)+.5)*width/float32(max(1, v.categoryCount())) - 28
			bottom.Child(el.Div().Absolute().Left(left).W(el.Dp(56)).Items(el.Center).Child(el.Text(v.labels[index]).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1)))
		}
	}
	for _, line := range v.options.references {
		if line.Label == "" || line.Value < lo || line.Value > hi {
			continue
		}
		label := el.Div().Absolute().Child(el.Text(line.Label).TextSize(theme.TextXs).TextColor(line.Color))
		if horizontal {
			label.Left(v.barValueFraction(line.Value, lo, hi)*width + 4).Top(4)
		} else {
			label.Left(4).Top(float32(axisFraction(line.Value, lo, hi))*v.height - 16)
		}
		plot.Child(label)
	}
	if v.hover >= 0 && v.hover < len(v.labels) {
		datum := v.tooltipData()
		tip := el.Div().Absolute().Left(8).Top(8).W(el.Dp(min(168, width))).P(theme.SpaceMd).Gap(theme.SpaceXs).Rounded(theme.RadiusMd).Bg(theme.Surface).Border(1, theme.Border)
		if horizontal {
			tip.Top(max(0, min(v.height-60, (float32(v.hover)+.5)*v.height/float32(max(1, v.categoryCount()))+8)))
		}
		if v.options.tooltip != nil {
			tip.Child(v.options.tooltip(cx, datum))
		} else {
			tip.Child(el.Text(datum.Label).Bold().TextSize(theme.TextSm))
			for _, value := range datum.Values {
				tip.Child(el.Text(value.Name + "  " + value.Text).TextSize(theme.TextSm))
			}
		}
		plot.Child(tip)
	}
	root.Child(el.Div().Row().Mt(gutter.Top).Items(el.Start).Child(axis, plot, el.Div().W(el.Dp(gutter.Right)).NoShrink()))
	if gutter.Bottom > 0 {
		root.Child(el.Div().Row().Child(el.Div().W(el.Dp(gutter.Left)).NoShrink(), bottom, el.Div().W(el.Dp(gutter.Right)).NoShrink()))
	}
	return root
}

// Convert a screen-space gradient direction to the canonical bottom-up plot.
func (v *ChartView) barGradientDirection(direction el.Side) el.Side {
	if direction > el.Right {
		direction = el.Bottom
	}
	switch v.options.barAlignment {
	case BarAlignmentTop:
		if direction == el.Top {
			return el.Bottom
		}
		if direction == el.Bottom {
			return el.Top
		}
	case BarAlignmentLeft:
		switch direction {
		case el.Top:
			return el.Left
		case el.Bottom:
			return el.Right
		case el.Left:
			return el.Bottom
		case el.Right:
			return el.Top
		}
	case BarAlignmentRight:
		switch direction {
		case el.Top:
			return el.Left
		case el.Bottom:
			return el.Right
		case el.Left:
			return el.Top
		case el.Right:
			return el.Bottom
		}
	}
	return direction
}
