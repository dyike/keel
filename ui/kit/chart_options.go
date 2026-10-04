package kit

import (
	"image/color"
	"math"
	"slices"

	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ChartCurve selects interpolation between finite data points.
type ChartCurve uint8

const (
	ChartCurveLinear ChartCurve = iota
	ChartCurveStepAfter
	ChartCurveSmooth
)

// ChartSeriesStyle overrides a series' theme colors and line presentation.
// Nil colors use the palette; zero Width uses 2dp. Fill alpha is preserved.
type ChartSeriesStyle struct {
	Stroke, Fill *color.NRGBA
	Width        float32
	Dots         bool
}
type ChartReference struct {
	Value float64
	Color color.NRGBA
	Label string
}
type ChartTooltipValue struct {
	Series int
	Name   string
	Value  float64
	Text   string
	Color  color.NRGBA
}
type ChartTooltip struct {
	Index  int
	Label  string
	Values []ChartTooltipValue
}
type chartOptions struct {
	barAlignment   BarAlignment
	orientedWidth  float32
	barFill        func(ChartBarDatum) ChartBarFill
	futureSlots    int
	pointCount     int
	barGradient    func(ChartBarDatum, ChartBarRange) []ChartColorStop
	gutter         *ChartGutter
	yLabelsInside  bool
	domain         *[2]float64
	yTicks, xTicks int
	gridColumns    int
	gridDashed     bool
	references     []ChartReference
	curve          ChartCurve
	styles         map[int]ChartSeriesStyle
	tooltip        func(*el.Context, ChartTooltip) el.Element
	radarMax       float64
	radarLevels    int
	radarRadius    float32
	radarLabel     func(*el.Context, int, string) el.Element
}

// YDomain pins the exact domain. Invalid or equal endpoints leave it unchanged.
func (v *ChartView) YDomain(lo, hi float64) *ChartView {
	if finiteNumber(lo) && finiteNumber(hi) && lo < hi {
		v.options.domain = &[2]float64{lo, hi}
	}
	return v
}
func (v *ChartView) AutoDomain() *ChartView { v.options.domain = nil; return v }

// YTickCount uses evenly spaced ticks including both endpoints; 0 restores nice ticks.
func (v *ChartView) YTickCount(n int) *ChartView {
	if n == 0 || n >= 2 && n <= 50 {
		v.options.yTicks = n
	}
	return v
}

// XTickCount spreads labels from first to last; 0 restores width-based thinning.
func (v *ChartView) XTickCount(n int) *ChartView {
	if n >= 0 && n <= 100 {
		v.options.xTicks = n
	}
	return v
}
func (v *ChartView) GridColumns(n int) *ChartView {
	if n >= 0 && n <= 100 {
		v.options.gridColumns = n
	}
	return v
}
func (v *ChartView) GridDashed(on bool) *ChartView { v.options.gridDashed = on; return v }

// ReferenceLines replaces horizontal reference guides; non-finite values are ignored.
func (v *ChartView) ReferenceLines(lines ...ChartReference) *ChartView {
	v.options.references = nil
	for _, l := range lines {
		if finiteNumber(l.Value) {
			v.options.references = append(v.options.references, l)
		}
	}
	return v
}
func (v *ChartView) Curve(curve ChartCurve) *ChartView {
	if curve <= ChartCurveSmooth {
		v.options.curve = curve
	}
	return v
}
func (v *ChartView) SeriesStyle(index int, style ChartSeriesStyle) *ChartView {
	if index < 0 || !finiteNumber(float64(style.Width)) || style.Width < 0 || style.Width > 32 {
		return v
	}
	if v.options.styles == nil {
		v.options.styles = map[int]ChartSeriesStyle{}
	}
	if style.Stroke != nil {
		c := *style.Stroke
		style.Stroke = &c
	}
	if style.Fill != nil {
		c := *style.Fill
		style.Fill = &c
	}
	v.options.styles[index] = style
	return v
}

// TooltipContent replaces the tooltip body; nil restores the default. Values are copied.
func (v *ChartView) TooltipContent(fn func(*el.Context, ChartTooltip) el.Element) *ChartView {
	v.options.tooltip = fn
	return v
}
func (v *ChartView) seriesColor(i int) color.NRGBA {
	if s := v.options.styles[i]; s.Stroke != nil {
		return *s.Stroke
	}
	return theme.Chart[i%len(theme.Chart)]
}
func (v *ChartView) seriesWidth(i int) float32 {
	if w := v.options.styles[i].Width; w > 0 {
		return w
	}
	return 2
}
func (v *ChartView) tooltipData() ChartTooltip {
	d := ChartTooltip{Index: v.hover, Label: v.labels[v.hover]}
	for i, s := range v.series {
		if !v.hidden[i] {
			d.Values = append(d.Values, ChartTooltipValue{i, s.Name, v.value(s, v.hover), v.valueText(s, v.hover), v.seriesColor(i)})
		}
	}
	return d
}
func (v *ChartView) axisTicks() []float64 {
	lo, hi := v.span()
	ticks := niceTicks(lo, hi, 4)
	lo, hi = ticks[0], ticks[len(ticks)-1]
	if v.options.domain != nil {
		lo, hi = v.options.domain[0], v.options.domain[1]
	}
	n := v.options.yTicks
	if n == 0 && v.options.domain == nil {
		return ticks
	}
	if n == 0 {
		n = 5
	}
	ticks = make([]float64, n)
	for i := range ticks {
		t := float64(i) / float64(n-1)
		ticks[i] = lo*(1-t) + hi*t
	}
	ticks[0], ticks[n-1] = lo, hi
	return ticks
}
func (v *ChartView) xTickIndices(width float32) []int {
	n := len(v.labels)
	if n == 0 {
		return nil
	}
	out := []int{}
	if v.options.xTicks > 0 {
		slots := n
		if v.options.pointCount > 0 {
			slots = v.categoryCount()
		}
		count := min(slots, v.options.xTicks)
		if count == 1 {
			return []int{0}
		}
		for i := 0; i < count; i++ {
			index := int(math.Round(float64(i) * float64(slots-1) / float64(count-1)))
			if index < n {
				out = append(out, index)
			}
		}
		return out
	}
	every := max(1, int(math.Ceil(56*float64(v.categoryCount())/float64(max(width, 1)))))
	for i := 0; i < n; i += every {
		out = append(out, i)
	}
	return out
}
func (v *ChartView) curvePoints(points []f32.Point) []f32.Point {
	if v.options.curve == ChartCurveLinear || len(points) < 2 {
		return points
	}
	out := []f32.Point{points[0]}
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		if v.options.curve == ChartCurveStepAfter {
			out = append(out, f32.Pt(b.X, a.Y), b)
			continue
		}
		for j := 1; j <= 8; j++ {
			t := float32(j) / 8
			s := t * t * (3 - 2*t)
			out = append(out, f32.Pt(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*s))
		}
	}
	return out
}

// RadarMax pins the outer ring's positive value; zero restores automatic scaling.
func (v *ChartView) RadarMax(value float64) *ChartView {
	if finiteNumber(value) && value >= 0 {
		v.options.radarMax = value
	}
	return v
}
func (v *ChartView) GridLevels(n int) *ChartView {
	if n >= 1 && n <= 20 {
		v.options.radarLevels = n
	}
	return v
}

// OuterRadius is the radar radius in dp; zero fits the available chart bounds.
func (v *ChartView) OuterRadius(dp float32) *ChartView {
	if finiteNumber(float64(dp)) && dp >= 0 {
		v.options.radarRadius = dp
	}
	return v
}
func (v *ChartView) RadarLabel(fn func(*el.Context, int, string) el.Element) *ChartView {
	v.options.radarLabel = fn
	return v
}

// Data returns owned copies of the chart's categories and series.
func (v *ChartView) Data() ([]string, []Series) {
	ss := slices.Clone(v.series)
	for i := range ss {
		ss[i].Values = slices.Clone(ss[i].Values)
	}
	return slices.Clone(v.labels), ss
}
