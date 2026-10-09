package main

import (
	"fmt"
	"image"
	"math"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/plot"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("plot_primitives", "data", func() core.Widget {
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).W(el.Dp(640)).MaxW(el.Full).Gap(12).Child(
				el.Text(demoText("Drawing primitives: stacked bars, trend lines, and donut charts", "绘图基础件：堆叠柱、趋势线与环图")),
				el.Widget(core.Func(func(gtx core.C) core.D {
					size := gtx.Constraints.Max
					dp := func(v float32) float32 { return float32(gtx.Dp(unit.Dp(v))) }
					c := plot.Canvas{Context: gtx, Bounds: image.Rectangle{Max: size}}
					cats := []string{"Mon", "Tue", "Wed", "Thu"}
					x := plot.NewBand(cats, [2]float64{float64(dp(40)), float64(size.X) - float64(dp(16))}).Padding(.25, .1)
					y := plot.ScaleLinear{Domain: [2]float64{-10, 40}, Range: [2]float64{float64(size.Y) - float64(dp(30)), float64(dp(10))}}
					stacks, _ := plot.Stack([][]float64{{14, 20, 12, 25}, {8, 10, -6, 9}})
					var ticks []plot.AxisText
					for i, cat := range cats {
						left, _ := x.Map(cat)
						ticks = append(ticks, plot.AxisText{Position: float32(left + x.Bandwidth()/2), Text: cat})
						for s, values := range stacks {
							lo, _ := y.Map(values[i].Low)
							hi, _ := y.Map(values[i].High)
							plot.Bar{Rect: plot.Rectangle{Min: f32.Pt(float32(left), float32(lo)), Max: f32.Pt(float32(left+x.Bandwidth()), float32(hi))}, Fill: theme.Chart[s], Radius: dp(3)}.Paint(c)
						}
					}
					var yTicks []plot.AxisText
					for _, value := range []float64{-10, 0, 10, 20, 30, 40} {
						pos, _ := y.Map(value)
						yTicks = append(yTicks, plot.AxisText{Position: float32(pos), Text: fmt.Sprint(value)})
					}
					plot.Axis{Side: plot.AxisLeft, At: dp(40), From: dp(10), To: float32(size.Y) - dp(30), TickSize: dp(4), Width: dp(1), Ticks: yTicks, Stroke: theme.Border, TextColor: theme.Muted}.Paint(c)
					zero, _ := y.Map(0)
					plot.Axis{Side: plot.AxisBottom, At: float32(size.Y) - dp(26), From: dp(40), To: float32(size.X) - dp(16), TickSize: dp(4), Width: dp(1), Ticks: ticks, Stroke: theme.Border, TextColor: theme.Muted}.Paint(c)
					plot.CrossLine{At: f32.Pt(0, float32(zero)), Horizontal: true, Stroke: theme.Border, Width: dp(1)}.Paint(c)
					var points []f32.Point
					for i, v := range []float64{16, 30, 17, 36} {
						xx, _ := x.Map(cats[i])
						yy, _ := y.Map(v)
						points = append(points, f32.Pt(float32(xx+x.Bandwidth()/2), float32(yy)))
					}
					plot.Line{Points: points, Stroke: theme.Chart[2], Width: dp(2), DotRadius: dp(4), Curve: plot.CurveSmooth}.Paint(c)
					return core.D{Size: size}
				})).H(el.Dp(240)),
				el.Text(demoText("Blue and orange form separate positive and negative stacks; green is another trend series. Values below zero are negative.", "蓝色、橙色为独立正负堆叠，绿色为另一组趋势数据；零线以下表示负值。")),
				el.Text("Mon：14 + 8；Tue：20 + 10；Wed：12 − 6；Thu：25 + 9"),
				el.Widget(core.Func(func(gtx core.C) core.D {
					size := gtx.Constraints.Max
					dp := func(v float32) float32 { return float32(gtx.Dp(unit.Dp(v))) }
					c := plot.Canvas{Context: gtx, Bounds: image.Rectangle{Max: size}}
					arcs, _ := plot.Pie([]float64{40, 35, 25}, -math.Pi/2, 2*math.Pi, .04)
					for _, a := range arcs {
						plot.Arc{Center: f32.Pt(float32(size.X)/2, float32(size.Y)/2), Inner: dp(32), Outer: dp(60), Start: a.Start, End: a.End, Fill: theme.Chart[a.Index]}.Paint(c)
					}
					return core.D{Size: size}
				})).H(el.Dp(140)),
				el.Text(demoText("Donut: 40 / 35 / 25. The application supplies data descriptions and interactions for low-level drawing.", "环图：40 / 35 / 25。低层绘图由应用提供数据描述与交互。")),
			)
		}))
	})
}
