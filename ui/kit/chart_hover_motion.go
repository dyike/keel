package kit

import (
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"time"
)

// chartHoverMotion blends emphasis weights. A rapid change begins at the
// currently painted weights instead of restarting from the previous endpoint.
type chartHoverMotion struct {
	target       int
	values, from []float32
	start        time.Time
	ready        bool
}

const chartHoverDuration = 150 * time.Millisecond

func (m *chartHoverMotion) update(gtx core.C, count, target int, animate bool) {
	if !gtx.Enabled() {
		return
	}
	if !m.ready || len(m.values) != count {
		m.values = make([]float32, count)
		m.from = make([]float32, count)
		m.target = -1
		m.ready = true
		m.start = time.Time{}
	}
	if !animate || theme.ReducedMotion {
		clear(m.values)
		if target >= 0 && target < count {
			m.values[target] = 1
		}
		m.target = target
		m.start = time.Time{}
		return
	}
	if !m.start.IsZero() {
		t := max(0, min(1, float32(gtx.Now.Sub(m.start))/float32(chartHoverDuration)))
		ease := 1 - (1-t)*(1-t)*(1-t)
		for i, from := range m.from {
			to := float32(0)
			if i == m.target {
				to = 1
			}
			m.values[i] = from + (to-from)*ease
		}
		if t >= 1 {
			m.start = time.Time{}
		}
	}
	if target != m.target {
		copy(m.from, m.values)
		m.target, m.start = target, gtx.Now
	}
	if !m.start.IsZero() {
		gtx.Execute(op.InvalidateCmd{})
	}
}
func (m *chartHoverMotion) weight(index int) float32 {
	if index < 0 || index >= len(m.values) {
		return 0
	}
	return m.values[index]
}
func (m *chartHoverMotion) total() float32 {
	total := float32(0)
	for _, value := range m.values {
		total += value
	}
	return min(1, total)
}

// HoverAnimation enables 150ms hover emphasis transitions, on by default.
// Reduced motion always shows the target emphasis immediately.
func (v *PieChartView) HoverAnimation(on bool) *PieChartView { v.noHoverAnimation = !on; return v }
func (v *SankeyChartView) HoverAnimation(on bool) *SankeyChartView {
	v.noHoverAnimation = !on
	return v
}

func (v *ChartView) HoverAnimation(on bool) *ChartView { v.noHoverAnimation = !on; return v }
func (v *CandlestickChartView) HoverAnimation(on bool) *CandlestickChartView {
	v.chart.HoverAnimation(on)
	return v
}
func chartEmphasis(c color.NRGBA, weight float32) color.NRGBA {
	c.A = uint8(float32(c.A) * max(0, min(1, weight)))
	return c
}
