package kit

// FutureSlots reserves n empty category positions after the data, from 0 to
// 100000. Existing points, bars and candles use the remaining portion of the
// plot. Empty slots have no labels, tooltip, data-table rows or effect on YDomain.
// This is a categorical spacing option; it does not generate dates or values.
func (v *ChartView) FutureSlots(n int) *ChartView {
	if n >= 0 && n <= 100000 && (n != v.options.futureSlots || v.options.pointCount != 0) {
		v.options.futureSlots = n
		v.options.pointCount = 0
		v.hoverMotion = chartHoverMotion{}
		v.hover = -1
	}
	return v
}
func (v *ChartView) categoryCount() int {
	return max(len(v.labels)+v.options.futureSlots, v.options.pointCount)
}
func (v *CandlestickChartView) FutureSlots(n int) *CandlestickChartView {
	v.chart.FutureSlots(n)
	return v
}

// PointCount pins the categorical axis to at least n positions (0–100000).
// Appending data fills reserved positions without moving existing categories.
// More data than n expands the axis. Zero restores automatic spacing.
// PointCount and FutureSlots replace one another. Radar ignores both.
func (v *ChartView) PointCount(n int) *ChartView {
	if n >= 0 && n <= 100000 && (n != v.options.pointCount || v.options.futureSlots != 0) {
		v.options.pointCount = n
		v.options.futureSlots = 0
		v.hover = -1
		v.hoverMotion = chartHoverMotion{}
	}
	return v
}
func (v *CandlestickChartView) PointCount(n int) *CandlestickChartView {
	v.chart.PointCount(n)
	return v
}
