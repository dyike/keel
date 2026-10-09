package kit

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/ui/core"
)

// ChartColorStop locates a color along a bar: zero is its base, one its tip.
// Positions outside [0,1] are interpolated at the visible boundaries.
type ChartColorStop struct {
	Position float64
	Color    color.NRGBA
}

// ChartBarRange contains the displayed value domain and this segment's endpoints.
// Stacked segments use accumulated Base/Tip; ChartBarDatum.Value stays original.
type ChartBarRange struct{ Min, Max, Base, Tip float64 }

// ChartToBar maps a value to a base-to-tip gradient position, without clamping.
func (r ChartBarRange) ChartToBar(value float64) float64 {
	scale := max(math.Abs(r.Base), math.Abs(r.Tip), math.Abs(value))
	if scale == 0 || r.Base == r.Tip {
		return 0
	}
	return (value/scale - r.Base/scale) / (r.Tip/scale - r.Base/scale)
}

// BarGradient sets a pure per-segment gradient callback and clears BarFill.
// Nil restores series colors. Stops are copied, sorted and clipped; invalid or
// empty input falls back to the series color. Equal positions use the last color.
// Gradients run from the base to the tip for both positive and negative bars.
func (v *ChartView) BarGradient(fn func(ChartBarDatum, ChartBarRange) []ChartColorStop) *ChartView {
	v.options.barGradient = fn
	v.options.barFill = nil
	return v
}

func clippedChartStops(input []ChartColorStop) []ChartColorStop {
	if len(input) == 0 {
		return nil
	}
	stops := append([]ChartColorStop(nil), input...)
	for _, s := range stops {
		if !finiteNumber(s.Position) {
			return nil
		}
	}
	sort.SliceStable(stops, func(i, j int) bool { return stops[i].Position < stops[j].Position })
	n := 0
	for _, s := range stops {
		if n > 0 && stops[n-1].Position == s.Position {
			stops[n-1] = s
		} else {
			stops[n] = s
			n++
		}
	}
	stops = stops[:n]
	sample := func(x float64) color.NRGBA {
		if x <= stops[0].Position {
			return stops[0].Color
		}
		for i := 1; i < len(stops); i++ {
			if x <= stops[i].Position {
				a, b := stops[i-1], stops[i]
				scale := max(math.Abs(a.Position), math.Abs(b.Position), 1)
				t := (x/scale - a.Position/scale) / (b.Position/scale - a.Position/scale)
				return blendChartColor(a.Color, b.Color, t)
			}
		}
		return stops[len(stops)-1].Color
	}
	out := []ChartColorStop{{0, sample(0)}}
	for _, s := range stops {
		if s.Position > 0 && s.Position < 1 {
			out = append(out, s)
		}
	}
	return append(out, ChartColorStop{1, sample(1)})
}

// Interpolate in linear light with premultiplied alpha, matching Gio's brush.
func blendChartColor(a, b color.NRGBA, t float64) color.NRGBA {
	linear := func(c uint8) float64 {
		x := float64(c) / 255
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	alpha := (float64(a.A)*(1-t) + float64(b.A)*t) / 255
	if alpha <= 0 {
		return color.NRGBA{}
	}
	channel := func(x, y uint8) uint8 {
		v := (linear(x)*float64(a.A)/255*(1-t) + linear(y)*float64(b.A)/255*t) / alpha
		if v <= .0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - .055
		}
		return uint8(math.Round(max(0, min(1, v)) * 255))
	}
	return color.NRGBA{R: channel(a.R, b.R), G: channel(a.G, b.G), B: channel(a.B, b.B), A: uint8(math.Round(alpha * 255))}
}

func (v *ChartView) paintGradientStops(gtx core.C, shape clip.RRect, r ChartBarRange, input []ChartColorStop) bool {
	stops := clippedChartStops(input)
	if len(stops) == 0 {
		return false
	}
	transform := v.barTransform(gtx.Constraints.Max)
	rect := shape.Rect
	x := float32(rect.Min.X+rect.Max.X) / 2
	base, tip := float32(rect.Max.Y), float32(rect.Min.Y)
	if r.Tip < r.Base {
		base, tip = tip, base
	}
	a, b := transform.Transform(f32.Pt(x, base)), transform.Transform(f32.Pt(x, tip))
	area := shape.Op(gtx.Ops).Push(gtx.Ops)
	defer area.Pop()
	brush := op.Affine(transform.Invert()).Push(gtx.Ops)
	defer brush.Pop()
	// Round shared strip boundaries once so adjacent stops leave no seams.
	horizontal := a.X != b.X
	axis := func(t float64) int {
		if horizontal {
			return int(math.Round(float64(a.X) + (float64(b.X)-float64(a.X))*t))
		}
		return int(math.Round(float64(a.Y) + (float64(b.Y)-float64(a.Y))*t))
	}
	p0, p1 := transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Min.Y))), transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Max.Y)))
	bounds := image.Rect(int(min(p0.X, p1.X)), int(min(p0.Y, p1.Y)), int(max(p0.X, p1.X)), int(max(p0.Y, p1.Y)))
	for i := 1; i < len(stops); i++ {
		lo, hi := stops[i-1], stops[i]
		u, w := axis(lo.Position), axis(hi.Position)
		if u == w {
			continue
		}
		band := bounds
		if horizontal {
			band.Min.X = min(u, w)
			band.Max.X = max(u, w)
		} else {
			band.Min.Y = min(u, w)
			band.Max.Y = max(u, w)
		}
		first, last := a, a
		if horizontal {
			first.X = float32(u)
			last.X = float32(w)
		} else {
			first.Y = float32(u)
			last.Y = float32(w)
		}
		stripe := clip.Rect(band).Push(gtx.Ops)
		paint.LinearGradientOp{Stop1: first, Stop2: last, Color1: lo.Color, Color2: hi.Color}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stripe.Pop()
	}
	return true
}
