package main

import (
	"image"
	"math"

	"gioui.org/f32"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/plot"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("plot_primitives", "data", func() core.Widget {
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(640)).MaxW(el.Full).Gap(12).Child(
				el.Text("公共绘图基础件：正负堆叠与趋势线"),
				el.Widget(core.Func(func(gtx core.C) core.D {
					size := gtx.Constraints.Max
					c := plot.Canvas{Context: gtx, Bounds: image.Rectangle{Max: size}}
					cats := []string{"Mon", "Tue", "Wed", "Thu"}
					x := plot.NewBand(cats, [2]float64{40, float64(size.X - 16)}).Padding(.25, .1)
					y := plot.ScaleLinear{Domain: [2]float64{-10, 40}, Range: [2]float64{float64(size.Y - 30), 10}}
					stacks, _ := plot.Stack([][]float64{{14, 20, 12, 25}, {8, 10, -6, 9}})
					var ticks []plot.AxisText
					for i, cat := range cats {
						left, _ := x.Map(cat)
						ticks = append(ticks, plot.AxisText{Position: float32(left + x.Bandwidth()/2), Text: cat})
						for s, values := range stacks {
							lo, _ := y.Map(values[i].Low)
							hi, _ := y.Map(values[i].High)
							plot.Bar{Rect: plot.Rectangle{Min: f32.Pt(float32(left), float32(lo)), Max: f32.Pt(float32(left+x.Bandwidth()), float32(hi))}, Fill: theme.Chart[s], Radius: 3}.Paint(c)
						}
					}
					zero, _ := y.Map(0)
					plot.Axis{Side: plot.AxisBottom, At: float32(size.Y - 26), From: 40, To: float32(size.X - 16), TickSize: 4, Width: 1, Ticks: ticks, Stroke: theme.Border, TextColor: theme.Muted}.Paint(c)
					plot.CrossLine{At: f32.Pt(0, float32(zero)), Horizontal: true, Stroke: theme.Border, Width: 1}.Paint(c)
					var points []f32.Point
					for i, v := range []float64{16, 30, 17, 36} {
						xx, _ := x.Map(cats[i])
						yy, _ := y.Map(v)
						points = append(points, f32.Pt(float32(xx+x.Bandwidth()/2), float32(yy)))
					}
					plot.Line{Points: points, Stroke: theme.Chart[2], Width: 2, DotRadius: 4, Curve: plot.CurveSmooth}.Paint(c)
					return core.D{Size: size}
				})).H(el.Dp(240)),
				el.Text("Mon：14 + 8；Tue：20 + 10；Wed：12 − 6；Thu：25 + 9"),
				el.Widget(core.Func(func(gtx core.C) core.D {
					size := gtx.Constraints.Max
					c := plot.Canvas{Context: gtx, Bounds: image.Rectangle{Max: size}}
					arcs, _ := plot.Pie([]float64{40, 35, 25}, -math.Pi/2, 2*math.Pi, .04)
					for _, a := range arcs {
						plot.Arc{Center: f32.Pt(float32(size.X)/2, 70), Inner: 32, Outer: 60, Start: a.Start, End: a.End, Fill: theme.Chart[a.Index]}.Paint(c)
					}
					return core.D{Size: size}
				})).H(el.Dp(140)),
				el.Text("环图：40 / 35 / 25。低层绘图由应用提供数据描述与交互。"),
			)
		}))
	})
}
