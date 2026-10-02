package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestChartMissingOwnershipLegendAndRename(t *testing.T) {
	labels := []string{"a", "b", "c"}
	values := []float64{-4, math.NaN(), 8}
	c := LineChart(labels, Series{Name: "A", Values: values}, Series{Name: "B", Values: []float64{100}})
	labels[0] = "changed"
	values[0] = 999
	if c.labels[0] != "a" || c.value(c.series[0], 0) != -4 {
		t.Fatal("data aliases caller")
	}
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(400)).Child(c.Render(cx)) }), 400, 1)
	click(t, h, "B")
	if lo, hi := c.span(); lo != -4 || hi != 8 {
		t.Fatalf("hidden/missing span %v..%v", lo, hi)
	}
	if c.valueText(c.series[0], 1) != "—" || c.valueText(c.series[1], 2) != "—" {
		t.Fatal("missing shown as zero")
	}
	c.SetDisabled(true)
	h.Frame()
	click(t, h, "B")
	if !c.hidden[1] {
		t.Fatal("disabled legend toggled")
	}
	c.SetDisabled(false)
	c.SetData([]string{"x"}, Series{Name: "Renamed", Values: []float64{2}})
	c.table = true
	h.Frame()
	if !shown(h, "Renamed") || !shown(h, "x | 2") {
		t.Fatal("table retained old schema")
	}
}

func TestChartDenseSamplesPreserveSpikesAndGaps(t *testing.T) {
	values := make([]float64, 100000)
	values[4567] = 100
	values[90000] = -100
	values[50000] = math.NaN()
	segments := chartSegments(values, len(values), .01, func(x float64) float32 { return float32(x) }, 1000)
	count, high, low := 0, false, false
	if len(segments) != 2 {
		t.Fatal("missing point joined")
	}
	for _, s := range segments {
		for _, p := range s {
			count++
			high = high || p.Y == 100
			low = low || p.Y == -100
		}
	}
	if count > 4100 || !high || !low {
		t.Fatalf("dense samples count=%d peaks=%v %v", count, high, low)
	}
}

func TestChartExtremeTicksAndStackedData(t *testing.T) {
	for _, bounds := range [][2]float64{{-math.MaxFloat64, math.MaxFloat64}, {math.MaxFloat64, math.MaxFloat64}, {0, math.SmallestNonzeroFloat64}} {
		ticks := niceTicks(bounds[0], bounds[1], 4)
		if len(ticks) < 2 || len(ticks) > 52 {
			t.Fatal(ticks)
		}
		for _, x := range ticks {
			if !finiteNumber(x) {
				t.Fatal(ticks)
			}
		}
		if !finiteNumber(axisFraction(ticks[0], ticks[0], ticks[len(ticks)-1])) {
			t.Fatal("invalid projection")
		}
	}
	c := BarChart([]string{"x"}, Series{Values: []float64{math.MaxFloat64}}, Series{Values: []float64{math.MaxFloat64}}, Series{Values: []float64{math.Inf(-1)}}).Stacked()
	if lo, hi := c.span(); lo != 0 || hi != math.MaxFloat64 {
		t.Fatal(lo, hi)
	}
	renderView(c, 300, 1).Frame()
}

func BenchmarkChartDenseSamples(b *testing.B) {
	values := make([]float64, 100000)
	for i := range values {
		values[i] = math.Sin(float64(i))
	}
	b.ReportAllocs()
	for b.Loop() {
		chartSegments(values, len(values), .01, func(x float64) float32 { return float32(x) }, 1000)
	}
}

func TestAreaChartZeroBaselineAndGaps(t *testing.T) {
	c := AreaChart([]string{"a", "b", "c"}, Series{Name: "area", Values: []float64{3, math.NaN(), 5}})
	if lo, hi := c.span(); lo != 0 || hi != 5 {
		t.Fatal(lo, hi)
	}
	renderView(c, 300, 1).Frame()
	c.SetData([]string{"a", "b"}, Series{Name: "area", Values: []float64{-5, -3}})
	if lo, hi := c.span(); lo != -5 || hi != 0 {
		t.Fatal(lo, hi)
	}
	renderView(c, 300, 2).Frame()
}
