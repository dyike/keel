package kit

// ChartGutter reserves space around the Cartesian plot in dp. Height still
// describes the plot itself. Bottom contains the categorical labels; zero hides
// those labels. The default is Left:52, Bottom:18, Top:0, Right:0.
type ChartGutter struct{ Left, Right, Top, Bottom float32 }

// Gutter sets all four plot margins. Negative, non-finite or over-4096 values
// reject the entire configuration. It does not affect radar or pie charts.
func (v *ChartView) Gutter(g ChartGutter) *ChartView {
	for _, n := range []float32{g.Left, g.Right, g.Top, g.Bottom} {
		if n < 0 || n > 4096 || !finiteNumber(float64(n)) {
			return v
		}
	}
	v.options.gutter = &g
	return v
}
func (v *ChartView) AutoGutter() *ChartView { v.options.gutter = nil; return v }

// YLabelsInside places y-axis labels at the left edge inside the plot. Without
// explicit Gutter, this also removes the outside left label column. Labels can
// cover data; reference labels and tooltips are drawn over them.
func (v *ChartView) YLabelsInside(on bool) *ChartView { v.options.yLabelsInside = on; return v }
func (v *ChartView) chartGutter() ChartGutter {
	if v.options.gutter != nil {
		return *v.options.gutter
	}
	g := ChartGutter{Left: axisWidth, Bottom: 18}
	if v.options.yLabelsInside {
		g.Left = 0
	}
	return g
}

// Gutter configures margins for the OHLC plot.
func (v *CandlestickChartView) Gutter(g ChartGutter) *CandlestickChartView {
	v.chart.Gutter(g)
	return v
}
func (v *CandlestickChartView) AutoGutter() *CandlestickChartView { v.chart.AutoGutter(); return v }
func (v *CandlestickChartView) YLabelsInside(on bool) *CandlestickChartView {
	v.chart.YLabelsInside(on)
	return v
}
