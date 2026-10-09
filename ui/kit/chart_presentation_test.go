package kit

import (
	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestChartOrientedBarPixelsAndFutureHits(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, side := range []BarAlignment{BarAlignmentBottom, BarAlignmentTop, BarAlignmentLeft, BarAlignmentRight} {
			v := BarChart([]string{"first", "second"}, Series{Name: "series", Values: []float64{5, 8}}).BarAlignment(side).FutureSlots(2).HoverAnimation(false)
			calls := 0
			direction := el.Right
			v.BarFill(func(d ChartBarDatum) ChartBarFill {
				calls++
				if d.Name != "series" || d.Index < 0 || d.Index > 1 || d.Value != v.series[0].Values[d.Index] {
					t.Fatal("wrong bar datum", d)
				}
				return ChartBarFill{Gradient: &ChartBarGradient{Start: color.NRGBA{R: 255, A: 255}, End: color.NRGBA{B: 255, A: 255}, Direction: direction}}
			})
			gpu, err := headless.NewWindow(200*scale, 200*scale)
			if err != nil {
				t.Fatal(err)
			}
			var img *image.RGBA
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints = layout.Exact(image.Pt(200*scale, 200*scale))
				if side == BarAlignmentBottom {
					v.draw(gtx, 0, 10, []float64{0, 10})
				} else {
					v.drawOrientedBars(gtx, 0, 10, []float64{0, 10})
				}
				if err := gpu.Frame(gtx.Ops); err != nil {
					t.Fatal(err)
				}
				img = image.NewRGBA(image.Rect(0, 0, 200*scale, 200*scale))
				if err := gpu.Screenshot(img); err != nil {
					t.Fatal(err)
				}
			})
			left, _, future := image.Pt(18, 150), image.Pt(32, 150), image.Pt(175, 50)
			switch side {
			case BarAlignmentTop:
				left.Y = 50
			case BarAlignmentLeft:
				left, future = image.Pt(20, 25), image.Pt(50, 175)
			case BarAlignmentRight:
				left, future = image.Pt(120, 25), image.Pt(150, 175)
			}
			box := image.Rect(13, 100, 37, 200)
			switch side {
			case BarAlignmentTop:
				box = image.Rect(13, 0, 37, 100)
			case BarAlignmentLeft:
				box = image.Rect(0, 13, 100, 37)
			case BarAlignmentRight:
				box = image.Rect(100, 13, 200, 37)
			}
			for _, dir := range []el.Side{el.Top, el.Bottom, el.Left, el.Right} {
				direction = dir
				h.Frame()
				start, end := image.Pt(box.Min.X+box.Dx()/4, (box.Min.Y+box.Max.Y)/2), image.Pt(box.Max.X-box.Dx()/4, (box.Min.Y+box.Max.Y)/2)
				if dir == el.Top || dir == el.Bottom {
					start, end = image.Pt((box.Min.X+box.Max.X)/2, box.Min.Y+box.Dy()/4), image.Pt((box.Min.X+box.Max.X)/2, box.Max.Y-box.Dy()/4)
				}
				if dir == el.Top || dir == el.Left {
					start, end = end, start
				}
				a, b := img.RGBAAt(start.X*scale, start.Y*scale), img.RGBAAt(end.X*scale, end.Y*scale)
				if a.R <= a.B || b.B <= b.R || calls == 0 {
					t.Fatal("screen gradient reversed", scale, side, dir, a, b)
				}
			}
			h.Move(float32(left.X*scale), float32(left.Y*scale))
			if v.hover != 0 {
				t.Fatal("oriented hover", scale, side, v.hover)
			}
			h.Move(float32(future.X*scale), float32(future.Y*scale))
			if v.hover != -1 {
				t.Fatal("empty slot has tooltip", side, v.hover)
			}
			if lo, hi := v.span(); lo != 0 || hi != 8 {
				t.Fatal("future slots changed domain", lo, hi)
			}
			gpu.Release()
		}
	}
}

func TestChartHoverMotionRetargetAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	now := time.Now()
	target := -1
	m := chartHoverMotion{}
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; m.update(gtx, 3, target, true) })
	target = 0
	h.Frame()
	if m.weight(0) != 0 {
		t.Fatal("animation jumped")
	}
	now = now.Add(75 * time.Millisecond)
	h.Frame()
	if math.Abs(float64(m.weight(0)-.875)) > .001 {
		t.Fatal("midpoint", m.values)
	}
	target = 1
	h.Frame()
	if math.Abs(float64(m.weight(0)-.875)) > .001 {
		t.Fatal("retarget discontinuity")
	}
	now = now.Add(chartHoverDuration)
	h.Frame()
	if m.weight(0) != 0 || m.weight(1) != 1 {
		t.Fatal("target not reached", m.values)
	}
	theme.SetReducedMotion(true)
	target = 2
	h.Frame()
	if m.weight(2) != 1 || m.total() != 1 || !m.start.IsZero() {
		t.Fatal("reduced motion", m.values)
	}
	target = -1
	h.Frame()
	if m.total() != 0 {
		t.Fatal("leave emphasis")
	}
}

func TestSankeyMinimumFlowFitsNodes(t *testing.T) {
	rng := rand.New(rand.NewSource(91))
	for run := 0; run < 40; run++ {
		nodes := make([]SankeyNode, 12)
		links := []SankeyLink{}
		for i := 0; i < 11; i++ {
			for j := i + 1; j < 12; j++ {
				if rng.Intn(3) == 0 {
					links = append(links, SankeyLink{i, j, math.Pow(10, float64(rng.Intn(6)))})
				}
			}
		}
		v := SankeyChart(nodes, links).MinLinkWidth(8)
		for _, height := range []float32{20, 200} {
			for _, scale := range []SankeyValueScale{SankeyValueScaleLinear, SankeyValueScaleSqrt} {
				v.ValueScale(scale)
				g := v.layout(400, height)
				for _, n := range g.nodes {
					if !finiteNumber(float64(n.h)) || n.y < -0.002 || n.y+n.h > height+.002 {
						t.Fatal("node bounds", n, height)
					}
				}
				for i, l := range links {
					r := g.links[i]
					a, b := g.nodes[l.Source], g.nodes[l.Target]
					if r[1] <= 0 || r[3] <= 0 || r[0] < a.y-.002 || r[0]+r[1] > a.y+a.h+.002 || r[2] < b.y-.002 || r[2]+r[3] > b.y+b.h+.002 {
						t.Fatal("link outside port", r, a, b)
					}
				}
			}
		}
	}
}

func TestPieCustomTooltipFollowsPointerAndLegend(t *testing.T) {
	for _, scale := range []int{1, 2} {
		p := PieChart(PieSlice{"A", 3}, PieSlice{"B", 1}).TooltipContent(func(_ *el.Context, d PieChartTooltip) el.Element {
			if d.Index != 0 || d.Slice.Name != "A" {
				t.Fatal(d)
			}
			return el.Text("sector details")
		})
		h := renderView(p, 400, scale)
		h.Move(float32(280*scale), float32(150*scale))
		h.Frame()
		if !shown(h, "sector details") {
			t.Fatal("floating tooltip missing")
		}
		tip, _ := semanticNode(h, "tooltip")
		if tip.Desc.Bounds.Min.X < 0 || tip.Desc.Bounds.Max.X > 400*scale {
			t.Fatal("tooltip escaped viewport", tip.Desc.Bounds)
		}
		click(t, h, "B")
		h.Move(float32(280*scale), float32(150*scale))
		h.Frame()
		if p.fractions()[0] != 1 {
			t.Fatal("hidden slice retained share")
		}
		p.SetDisabled(true)
		h.Frame()
		if shown(h, "sector details") {
			t.Fatal("disabled tooltip remained")
		}
	}
}

func TestChartGutterLayoutAndValidation(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := LineChart([]string{"category A", "category B"}, Series{Values: []float64{2, 4}}).Gutter(ChartGutter{Left: 20, Bottom: 24})
		h := renderView(v, 400, scale)
		h.Frame()
		initial := v.plotW
		v.Gutter(ChartGutter{Left: 80, Bottom: 24})
		h.Frame()
		h.Frame()
		if math.Abs(float64(initial-v.plotW-60)) > 1 {
			t.Fatal("gutter did not reserve width", scale, initial, v.plotW)
		}
		v.Gutter(ChartGutter{Left: float32(math.NaN())})
		if v.options.gutter.Left != 80 {
			t.Fatal("invalid gutter accepted")
		}
		v.Gutter(ChartGutter{Left: 80})
		h.Frame()
		if shown(h, "category A") {
			t.Fatal("zero bottom retained labels")
		}
		v.AutoGutter().YLabelsInside(true)
		h.Frame()
		h.Frame()
		if v.chartGutter().Left != 0 || v.plotW <= initial {
			t.Fatal("inside labels retain outside column")
		}
	}
}

func TestCandleFutureSlotsKeepRawDataAndHover(t *testing.T) {
	c := CandlestickChart(Candle{"one", 8, 10, 0, 2}, Candle{"two", 2, 10, 0, 8}).FutureSlots(2).HoverAnimation(false)
	v := c.chart
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Exact(image.Pt(200, 100))
		v.draw(gtx, 0, 10, []float64{0, 10})
	})
	h.Move(25, 50)
	if v.hover != 0 {
		t.Fatal("first candle hit", v.hover)
	}
	h.Move(75, 50)
	if v.hover != 1 {
		t.Fatal("second candle hit", v.hover)
	}
	h.Move(175, 50)
	if v.hover != -1 {
		t.Fatal("future candle hit", v.hover)
	}
	if len(v.candles) != 2 || len(v.labels) != 2 {
		t.Fatal("future generated data")
	}
	c.FutureSlots(-1)
	if v.categoryCount() != 4 {
		t.Fatal("invalid future count accepted")
	}
	c.FutureSlots(0)
	h.Frame()
	h.Move(175, 50)
	if v.hover != 1 {
		t.Fatal("reset spacing", v.hover)
	}
}

func TestStackedBarFillRetainsOriginalNegativeValues(t *testing.T) {
	for _, side := range []BarAlignment{BarAlignmentBottom, BarAlignmentTop, BarAlignmentLeft, BarAlignmentRight} {
		seen := map[float64]bool{}
		v := BarChart([]string{"value"}, Series{Name: "positive", Values: []float64{4}}, Series{Name: "negative", Values: []float64{-3}}).Stacked().BarAlignment(side)
		v.BarFill(func(d ChartBarDatum) ChartBarFill {
			if !d.Stacked || d.Label != "value" {
				t.Fatal(d)
			}
			seen[d.Value] = true
			return ChartBarFill{Color: d.Color}
		})
		renderView(v, 400, 1).Frame()
		if !seen[4] || !seen[-3] {
			t.Fatal("missing stack segment", side, seen)
		}
	}
	v := SankeyChart(make([]SankeyNode, 1000), nil).MinLinkWidth(64)
	for _, n := range v.layout(20, 30).nodes {
		if !finiteNumber(float64(n.h)) || n.y+n.h > 30.001 {
			t.Fatal("empty dense graph", n)
		}
	}
}
