package kit

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestChartPinnedPointCount(t *testing.T) {
	v := LineChart([]string{"a", "b"}, Series{Values: []float64{1, 2}}).PointCount(7).XTickCount(3)
	if v.categoryCount() != 7 || !reflect.DeepEqual(v.xTickIndices(400), []int{0}) {
		t.Fatal("initial positions", v.xTickIndices(400))
	}
	v.SetData([]string{"a", "b", "c", "d"}, Series{Values: []float64{1, 2, 3, 4}})
	if v.categoryCount() != 7 || !reflect.DeepEqual(v.xTickIndices(400), []int{0, 3}) {
		t.Fatal("growing axis moved", v.categoryCount(), v.xTickIndices(400))
	}
	v.PointCount(2)
	if v.categoryCount() != 4 {
		t.Fatal("point count hid actual data")
	}
	v.PointCount(-1)
	if v.options.pointCount != 2 {
		t.Fatal("invalid count mutated")
	}
	v.FutureSlots(2)
	if v.categoryCount() != 6 || v.options.pointCount != 0 {
		t.Fatal("future slots didn't replace count")
	}
	v.PointCount(8)
	if v.categoryCount() != 8 || v.options.futureSlots != 0 {
		t.Fatal("count didn't replace future slots")
	}
	v.PointCount(0)
	if v.categoryCount() != 4 {
		t.Fatal("auto count not restored")
	}
	c := CandlestickChart(Candle{Label: "a", Open: 1, High: 3, Low: 0, Close: 2}).PointCount(9)
	if c.chart.categoryCount() != 9 {
		t.Fatal("candle count not forwarded")
	}
}

func TestChartGradientStopClipping(t *testing.T) {
	red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 255}
	raw := []ChartColorStop{{2, blue}, {-1, red}}
	got := clippedChartStops(raw)
	if len(got) != 2 || got[0].Position != 0 || got[1].Position != 1 || got[0].Color.R <= got[0].Color.B || got[1].Color.B <= got[1].Color.R {
		t.Fatal("outside stop interpolation", got)
	}
	if raw[0].Position != 2 {
		t.Fatal("mutated caller stops")
	}
	got = clippedChartStops([]ChartColorStop{{.5, red}, {.5, blue}})
	if got[0].Color != blue || got[len(got)-1].Color != blue {
		t.Fatal("duplicate precedence", got)
	}
	if clippedChartStops([]ChartColorStop{{math.NaN(), red}}) != nil {
		t.Fatal("accepted nonfinite stop")
	}
	r := ChartBarRange{Base: -math.MaxFloat64, Tip: math.MaxFloat64}
	if math.Abs(r.ChartToBar(0)-.5) > 1e-12 {
		t.Fatal("extreme mapping overflow")
	}
	r = ChartBarRange{Base: 3, Tip: 7}
	if r.ChartToBar(3) != 0 || r.ChartToBar(7) != 1 || r.ChartToBar(0) >= 0 {
		t.Fatal("stacked segment mapping")
	}
}

func TestChartMultiStopGradientPixels(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, side := range []BarAlignment{BarAlignmentBottom, BarAlignmentTop, BarAlignmentLeft, BarAlignmentRight} {
			for _, value := range []float64{10, -10} {
				v := BarChart([]string{"one"}, Series{Values: []float64{value}}).BarAlignment(side).HoverAnimation(false)
				v.BarGradient(func(d ChartBarDatum, r ChartBarRange) []ChartColorStop {
					if r.Base != 0 || r.Tip != value || d.Value != value {
						t.Fatal("wrong original endpoints", r, d)
					}
					return []ChartColorStop{{0, color.NRGBA{R: 255, A: 255}}, {.5, color.NRGBA{G: 255, A: 255}}, {1, color.NRGBA{B: 255, A: 255}}}
				})
				gpu, err := headless.NewWindow(200*scale, 200*scale)
				if err != nil {
					t.Fatal(err)
				}
				var pixels *image.RGBA
				uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints = layout.Exact(image.Pt(200*scale, 200*scale))
					paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					lo, hi := 0.0, 10.0
					if value < 0 {
						lo, hi = -10, 0
					}
					if side == BarAlignmentBottom {
						v.draw(gtx, lo, hi, []float64{lo, hi})
					} else {
						v.drawOrientedBars(gtx, lo, hi, []float64{lo, hi})
					}
					if err := gpu.Frame(gtx.Ops); err != nil {
						t.Fatal(err)
					}
					pixels = image.NewRGBA(image.Rect(0, 0, 200*scale, 200*scale))
					if err := gpu.Screenshot(pixels); err != nil {
						t.Fatal(err)
					}
				})
				canonical := image.Pt(200*scale, 200*scale)
				transform := v.barTransform(canonical)
				for _, fraction := range []float32{.1, .5, .9} {
					y := (1 - fraction) * 200 * float32(scale)
					if value < 0 {
						y = fraction * 200 * float32(scale)
					}
					pt := transform.Transform(f32.Pt(100*float32(scale), y)).Round()
					c := pixels.RGBAAt(pt.X, pt.Y)
					if fraction == .1 && (c.R <= c.G || c.R <= c.B) || fraction == .5 && (c.G <= c.R || c.G <= c.B) || fraction == .9 && (c.B <= c.R || c.B <= c.G) {
						t.Fatal("gradient orientation", scale, side, value, fraction, c)
					}
				}
				gpu.Release()
			}
		}
	}
}

func TestChartGradientMappingAcrossBars(t *testing.T) {
	v := BarChart([]string{"short", "tall"}, Series{Values: []float64{10, 20}}).HoverAnimation(false)
	v.BarGradient(func(_ ChartBarDatum, r ChartBarRange) []ChartColorStop {
		return []ChartColorStop{{r.ChartToBar(r.Min), color.NRGBA{R: 255, A: 128}}, {r.ChartToBar(r.Max), color.NRGBA{B: 255, A: 255}}}
	})
	gpu, err := headless.NewWindow(200, 200)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	var img *image.RGBA
	uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Exact(image.Pt(200, 200))
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		v.draw(gtx, 0, 20, []float64{0, 20})
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		img = image.NewRGBA(image.Rect(0, 0, 200, 200))
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
	})
	a, b := img.RGBAAt(50, 150), img.RGBAAt(150, 150)
	for i, x := range []uint8{a.R, a.G, a.B, a.A} {
		y := []uint8{b.R, b.G, b.B, b.A}[i]
		if math.Abs(float64(x)-float64(y)) > 2 {
			t.Fatal("same chart value has different color", a, b)
		}
	}
	v.BarFill(func(d ChartBarDatum) ChartBarFill { return ChartBarFill{Color: d.Color} })
	if v.options.barGradient != nil {
		t.Fatal("BarFill didn't replace gradient")
	}
	v.BarGradient(nil)
	if v.options.barFill != nil {
		t.Fatal("gradient reset didn't restore series color")
	}
	stacked := BarChart([]string{"a"}, Series{Values: []float64{3}}, Series{Values: []float64{4}}, Series{Values: []float64{-2}}, Series{Values: []float64{-5}}).Stacked()
	seen := map[int]ChartBarRange{}
	stacked.BarGradient(func(d ChartBarDatum, r ChartBarRange) []ChartColorStop {
		seen[d.Series] = r
		return []ChartColorStop{{0, d.Color}}
	})
	uitest.NewFunc(func(gtx core.C) { stacked.draw(gtx, -7, 7, []float64{-7, 7}) })
	if seen[1].Base != 3 || seen[1].Tip != 7 || seen[3].Base != -2 || seen[3].Tip != -7 {
		t.Fatal("stacked endpoints lost", seen)
	}
}
