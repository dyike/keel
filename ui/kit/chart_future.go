package kit

// FutureSlots reserves n empty category positions after the data, from 0 to
// 100000. Existing points, bars and candles use the remaining portion of the
// plot. Empty slots have no labels, tooltip, data-table rows or effect on YDomain.
// This is a categorical spacing option; it does not generate dates or values.
func (v *ChartView) FutureSlots(n int) *ChartView {
	if n >= 0 && n <= 100000 && n != v.options.futureSlots {
		v.options.futureSlots = n
		v.hover = -1
	}
	return v
}
func (v *ChartView) categoryCount() int { return len(v.labels) + v.options.futureSlots }
func (v *CandlestickChartView) FutureSlots(n int) *CandlestickChartView {
	v.chart.FutureSlots(n)
	return v
}
