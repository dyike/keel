package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/ui/el"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestChartPinnedDomainTicksStylesAndTooltip(t *testing.T) {
	v := LineChart([]string{"A", "B", "C", "D", "E"}, Series{Name: "S", Values: []float64{1, 2, 3, 4, 5}}).YDomain(-math.MaxFloat64, math.MaxFloat64).YTickCount(5).XTickCount(3)
	ticks := v.axisTicks()
	if len(ticks) != 5 || ticks[0] != -math.MaxFloat64 || ticks[4] != math.MaxFloat64 || ticks[2] != 0 {
		t.Fatal(ticks)
	}
	for _, x := range ticks {
		if !finiteNumber(x) {
			t.Fatal("nonfinite tick")
		}
	}
	if !reflect.DeepEqual(v.xTickIndices(300), []int{0, 2, 4}) {
		t.Fatal("endpoints not spread")
	}
	v.YDomain(5, 5).YTickCount(1)
	if !reflect.DeepEqual(ticks, v.axisTicks()) {
		t.Fatal("invalid option changed domain")
	}
	col := color.NRGBA{R: 200, A: 255}
	v.SeriesStyle(0, ChartSeriesStyle{Stroke: &col})
	col.R = 0
	if v.seriesColor(0).R != 200 {
		t.Fatal("borrowed style color")
	}
	v.hover = 1
	v.plotW = 300
	calls := 0
	v.TooltipContent(func(_ *el.Context, d ChartTooltip) el.Element {
		calls++
		if d.Index != 1 || d.Values[0].Value != 2 {
			t.Fatal(d)
		}
		d.Values[0].Value = 99
		return el.Text("custom tooltip")
	})
	h := renderView(v, 500, 1)
	if !shown(h, "custom tooltip") || calls == 0 || v.series[0].Values[1] != 2 {
		t.Fatal("custom tooltip missing or borrowed values")
	}
	v.SetDisabled(true)
	h.Frame()
	if shown(h, "custom tooltip") {
		t.Fatal("disabled chart retained hover")
	}
	v.AutoDomain().YTickCount(0)
	a := v.axisTicks()
	if a[0] > 1 || a[len(a)-1] < 5 {
		t.Fatal("auto range not restored")
	}
}
func TestChartCurveDoesNotOvershootOrMoveVertices(t *testing.T) {
	pts := []f32.Point{{X: 0, Y: 20}, {X: 20, Y: 0}, {X: 40, Y: 30}}
	v := LineChart(nil).Curve(ChartCurveSmooth)
	curve := v.curvePoints(pts)
	if curve[0] != pts[0] || curve[8] != pts[1] || curve[16] != pts[2] {
		t.Fatal("lost endpoints")
	}
	for i, p := range curve {
		if p.Y < 0 || p.Y > 30 || (i > 0 && p.X < curve[i-1].X) {
			t.Fatal("overshoot", p)
		}
	}
	v.Curve(ChartCurveStepAfter)
	if !reflect.DeepEqual(v.curvePoints(pts), []f32.Point{f32.Pt(0, 20), f32.Pt(20, 20), f32.Pt(20, 0), f32.Pt(40, 0), f32.Pt(40, 30)}) {
		t.Fatal("step path")
	}
}
func TestRadarScalingOwnershipLegendAndTable(t *testing.T) {
	values := []float64{20, 40, 80}
	v := RadarChart([]string{"A", "B", "C"}, Series{Name: "One", Values: values}, Series{Name: "Two", Values: []float64{10, 30, 60}}).Title("Radar")
	values[2] = 1000
	if v.radarScale() != 80 {
		t.Fatal("data borrowed")
	}
	v.RadarMax(100)
	if v.radarScale() != 100 {
		t.Fatal("fixed ring")
	}
	v.RadarMax(0)
	for _, scale := range []int{1, 2} {
		h := renderView(v, 500, scale)
		h.Frame()
		if !shown(h, "Radar") || !shown(h, "A") {
			t.Fatal("radar semantic labels missing")
		}
		click(t, h, "One")
		if v.radarScale() != 60 {
			t.Fatal("legend not applied")
		}
		click(t, h, "查看数据表")
		if !shown(h, "C | 80 | 60") {
			t.Fatal("radar raw table missing")
		}
		v.table = false
		v.hidden[0] = false
	}
	_, data := v.Data()
	data[0].Values[0] = 999
	if v.series[0].Values[0] != 20 {
		t.Fatal("Data borrowed")
	}
}
func TestSankeyRejectsInvalidDataAtomically(t *testing.T) {
	nodes := []SankeyNode{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	links := []SankeyLink{{0, 1, 60}, {0, 2, 40}}
	v := SankeyChart(nodes, links)
	if v.Error() != nil || v.values[0] != 100 {
		t.Fatal("valid graph", v.Error())
	}
	nodes[0].Name = "changed"
	links[0].Value = 999
	for _, bad := range [][]SankeyLink{{{0, 4, 1}}, {{0, 0, 1}}, {{0, 1, -1}}, {{0, 1, math.NaN()}}, {{0, 1, math.Inf(1)}}, {{0, 1, 1}, {1, 0, 1}}, {{0, 1, math.MaxFloat64}, {0, 2, math.MaxFloat64}}} {
		if v.SetData(v.nodes, bad) == nil {
			t.Fatal("invalid accepted", bad)
		}
		if v.nodes[0].Name != "A" || v.links[0].Value != 60 {
			t.Fatal("invalid graph modified data")
		}
	}
	bad := SankeyChart(nodes, []SankeyLink{{0, 9, 1}})
	if bad.Error() == nil {
		t.Fatal("constructor error missing")
	}
	if bad.SetData(nodes, nil) != nil || bad.Error() != nil {
		t.Fatal("valid replacement did not recover")
	}
}
func TestSankeyGeometryBoundsAndRibbonConservation(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	for run := 0; run < 50; run++ {
		n := 3 + rng.Intn(20)
		nodes := make([]SankeyNode, n)
		links := []SankeyLink{}
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if rng.Intn(4) == 0 {
					links = append(links, SankeyLink{i, j, float64(1 + rng.Intn(100))})
				}
			}
		}
		v := SankeyChart(nodes, links)
		for _, alignment := range []SankeyAlign{SankeyAlignJustify, SankeyAlignLeft, SankeyAlignRight, SankeyAlignCenter} {
			for _, scale := range []SankeyValueScale{SankeyValueScaleLinear, SankeyValueScaleSqrt} {
				v.NodeAlign(alignment).ValueScale(scale)
				g := v.layout(400, 180)
				for i, a := range g.nodes {
					if a.x < -0.001 || a.y < -0.001 || a.x+a.w > 400.001 || a.y+a.h > 180.001 {
						t.Fatalf("node out of bounds run=%d align=%d scale=%d %+v", run, alignment, scale, a)
					}
					for j, b := range g.nodes {
						if j > i && a.x == b.x && min(a.y+a.h, b.y+b.h)-max(a.y, b.y) > .001 {
							t.Fatal("nodes overlap", a, b)
						}
					}
				}
				outgoing, incoming := make([]float32, n), make([]float32, n)
				for i, l := range links {
					a, b := g.nodes[l.Source], g.nodes[l.Target]
					band := g.links[i]
					if a.x >= b.x {
						t.Fatal("backward flow")
					}
					if band[0] < a.y-.001 || band[0]+band[1] > a.y+a.h+.001 || band[2] < b.y-.001 || band[2]+band[3] > b.y+b.h+.001 {
						t.Fatal("ribbon exceeds node")
					}
					outgoing[l.Source] += band[1]
					incoming[l.Target] += band[3]
				}
				for i := range nodes {
					if v.values[i] > 0 && math.Abs(float64(max(incoming[i], outgoing[i])-g.nodes[i].h)) > .002 {
						t.Fatal("ribbons do not fill throughput")
					}
				}
			}
		}
	}
	// Dense isolated nodes must fit even after reserving gaps.
	v := SankeyChart(make([]SankeyNode, 1000), nil)
	for _, r := range v.layout(20, 30).nodes {
		if r.y < -0.001 || r.y+r.h > 30.001 {
			t.Fatal("dense zero-flow overflow", r)
		}
	}
}
