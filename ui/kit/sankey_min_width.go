package kit

// MinLinkWidth sets a visual minimum width in dp for positive flows (0–64).
// Zero preserves proportional sizing. Nodes grow to fit their incident links;
// when the column cannot hold all minima, the minimum is reduced uniformly.
// Displayed values and tooltips retain the original data, not adjusted widths.
func (v *SankeyChartView) MinLinkWidth(dp float32) *SankeyChartView {
	if dp >= 0 && dp <= 64 && finiteNumber(float64(dp)) {
		v.minLinkWidth = dp
	}
	return v
}

func (v *SankeyChartView) minimumFlowLayout(weights []float64, unit float64, columns [][]int, incoming, outgoing [][]int, h, gap, zeroH float32) ([]float64, float64, float64) {
	floor := float64(v.minLinkWidth)
	// A node must accommodate the larger number of incoming/outgoing ports.
	for _, column := range columns {
		ports, zeros := 0, 0
		for _, node := range column {
			a, b := 0, 0
			for _, li := range incoming[node] {
				if v.links[li].Value > 0 {
					a++
				}
			}
			for _, li := range outgoing[node] {
				if v.links[li].Value > 0 {
					b++
				}
			}
			ports += max(a, b)
			if v.values[node] == 0 {
				zeros++
			}
		}
		if ports > 0 {
			room := max(0, float64(h)-float64(gap)*float64(max(0, len(column)-1))-float64(zeros)*float64(zeroH))
			floor = min(floor, room/float64(ports))
		}
	}
	heights := make([]float64, len(weights))
	inSize, outSize := make([]float64, len(weights)), make([]float64, len(weights))
	measure := func(scale float64) bool {
		clear(inSize)
		clear(outSize)
		for _, l := range v.links {
			if l.Value <= 0 {
				continue
			}
			outSize[l.Source] += max(floor, weights[l.Source]*scale*(l.Value/v.values[l.Source]))
			inSize[l.Target] += max(floor, weights[l.Target]*scale*(l.Value/v.values[l.Target]))
		}
		for i, w := range weights {
			heights[i] = max(w*scale, inSize[i], outSize[i])
			if v.values[i] == 0 {
				heights[i] = float64(zeroH)
			}
		}
		for _, column := range columns {
			total := float64(gap) * float64(max(0, len(column)-1))
			for _, i := range column {
				total += heights[i]
			}
			if total > float64(h)+1e-6 {
				return false
			}
		}
		return true
	}
	lo, hi := 0.0, unit
	for range 48 {
		mid := (lo + hi) / 2
		if measure(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	measure(lo)
	return heights, lo, floor
}
