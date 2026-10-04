package kit

import (
	"image"
	"math"
	"slices"

	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Candle is one ordered category of open/high/low/close data.
type Candle struct {
	Label                  string
	Open, High, Low, Close float64
}

func (c Candle) valid() bool {
	return finiteNumber(c.Open) && finiteNumber(c.High) && finiteNumber(c.Low) && finiteNumber(c.Close) && c.Low <= min(c.Open, c.Close) && c.High >= max(c.Open, c.Close)
}

// CandlestickChartView renders OHLC data. Rising bodies are hollow, falling
// bodies solid, so direction can be read without relying only on color.
type CandlestickChartView struct{ chart *ChartView }

func CandlestickChart(data ...Candle) *CandlestickChartView {
	v := &CandlestickChartView{chart: LineChart(nil)}
	v.chart.kind = ChartCandlestick
	v.SetData(data...)
	return v
}
func (v *CandlestickChartView) Title(s string) *CandlestickChartView    { v.chart.Title(s); return v }
func (v *CandlestickChartView) Height(dp float32) *CandlestickChartView { v.chart.Height(dp); return v }
func (v *CandlestickChartView) Format(fn func(float64) string) *CandlestickChartView {
	v.chart.Format(fn)
	return v
}
func (v *CandlestickChartView) SetDisabled(on bool) { v.chart.SetDisabled(on) }

// SetData copies the data. Invalid candles become gaps, not repaired prices.
func (v *CandlestickChartView) SetData(data ...Candle) {
	labels := make([]string, len(data))
	series := make([]Series, 4)
	for i := range series {
		series[i].Values = make([]float64, len(data))
	}
	for i, c := range data {
		labels[i] = c.Label
		values := []float64{c.Open, c.High, c.Low, c.Close}
		for j := range series {
			if c.valid() {
				series[j].Values[i] = values[j]
			} else {
				series[j].Values[i] = math.NaN()
			}
		}
	}
	v.chart.SetData(labels, series...)
	v.chart.candles = slices.Clone(data)
}
func (v *CandlestickChartView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	names := []string{text.CandleOpen, text.CandleHigh, text.CandleLow, text.CandleClose}
	for i, name := range names {
		v.chart.series[i].Name = name
		if v.chart.tbl != nil {
			v.chart.tbl.cols[i+1].title = name
		}
	}
	return v.chart.Render(cx)
}

// candleBuckets emits at most one aggregate per horizontal pixel. Each bucket
// retains first open, last close, highest high and lowest low, without reordering.
func candleBuckets(data []Candle, pixels int) []Candle {
	step := max(1, int(math.Ceil(float64(len(data))/float64(max(pixels, 1)))))
	out := make([]Candle, 0, (len(data)+step-1)/step)
	for a := 0; a < len(data); a += step {
		c := Candle{Open: math.NaN()}
		for _, next := range data[a:min(a+step, len(data))] {
			if !next.valid() {
				continue
			}
			if !c.valid() {
				c = next
			} else {
				c.High = max(c.High, next.High)
				c.Low = min(c.Low, next.Low)
				c.Close = next.Close
			}
		}
		out = append(out, c)
	}
	return out
}
func (v *ChartView) drawCandles(gtx core.C, y func(float64) float32, dp func(float32) float32) {
	pixels := max(1, int(float64(gtx.Constraints.Max.X)*float64(len(v.candles))/float64(max(v.categoryCount(), 1))))
	data := candleBuckets(v.candles, pixels)
	if len(data) == 0 {
		return
	}
	step := max(1, int(math.Ceil(float64(len(v.candles))/float64(pixels))))
	band := float32(gtx.Constraints.Max.X) / float32(max(v.categoryCount(), 1))
	width := max(float32(1), min(dp(16), band*float32(step)*.65))
	for i, c := range data {
		if !c.valid() {
			continue
		}
		x := float32(i*step+min((i+1)*step, len(v.candles))) * .5 * band
		color := theme.Danger
		if c.Close >= c.Open {
			color = theme.Success
		}
		fillRect(gtx, image.Rect(int(x), int(y(c.High)), int(x)+max(1, int(dp(1))), max(int(y(c.High))+1, int(y(c.Low)))), color)
		top, bottom := min(y(c.Open), y(c.Close)), max(y(c.Open), y(c.Close))
		rect := image.Rect(int(x-width/2), int(top), max(int(x-width/2)+1, int(x+width/2)), max(int(top)+1, int(bottom)))
		if c.Close > c.Open {
			fillRect(gtx, rect, theme.Surface)
			paint.FillShape(gtx.Ops, color, clip.Stroke{Path: clip.Rect(rect).Path(), Width: dp(1)}.Op())
		} else {
			fillRect(gtx, rect, color)
		}
	}
}

// YDomain pins the price axis; AutoDomain restores data-driven bounds.
func (v *CandlestickChartView) YDomain(lo, hi float64) *CandlestickChartView {
	v.chart.YDomain(lo, hi)
	return v
}
func (v *CandlestickChartView) AutoDomain() *CandlestickChartView { v.chart.AutoDomain(); return v }
func (v *CandlestickChartView) YTickCount(n int) *CandlestickChartView {
	v.chart.YTickCount(n)
	return v
}
func (v *CandlestickChartView) XTickCount(n int) *CandlestickChartView {
	v.chart.XTickCount(n)
	return v
}
func (v *CandlestickChartView) GridColumns(n int) *CandlestickChartView {
	v.chart.GridColumns(n)
	return v
}
func (v *CandlestickChartView) GridDashed(on bool) *CandlestickChartView {
	v.chart.GridDashed(on)
	return v
}
func (v *CandlestickChartView) ReferenceLines(lines ...ChartReference) *CandlestickChartView {
	v.chart.ReferenceLines(lines...)
	return v
}
func (v *CandlestickChartView) TooltipContent(fn func(*el.Context, ChartTooltip) el.Element) *CandlestickChartView {
	v.chart.TooltipContent(fn)
	return v
}
