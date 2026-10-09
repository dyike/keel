package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestPlotFiniteDataOwnershipAndZoomLimits(t *testing.T) {
	points := []PlotPoint{{math.NaN(), 2}, {1, 2}, {3, 4}, {1, math.Inf(1)}}
	p := Plot(PlotSeries{Name: "data", Points: points}).Lines()
	points[1] = PlotPoint{999, 999}
	p.Reset()
	if p.x0 >= 1 || p.x1 <= 3 || p.x1 > 4 {
		t.Fatal("bad fit", p.x0, p.x1)
	}
	for range 10000 {
		p.zoom(.8, .5, .5)
	}
	check := func() {
		t.Helper()
		if !validPlotRange(p.x0, p.x1) || !validPlotRange(p.y0, p.y1) {
			t.Fatal("invalid view", p.x0, p.x1, p.y0, p.y1)
		}
	}
	check()
	for range 10000 {
		p.zoom(100, 1, 0)
	}
	check()
	for range 100 {
		p.pan(1e300, 1e300)
	}
	check()
	before := [4]float64{p.x0, p.x1, p.y0, p.y1}
	p.SetView(0, math.Inf(1), 0, 1)
	if before != [4]float64{p.x0, p.x1, p.y0, p.y1} {
		t.Fatal("invalid SetView changed state")
	}
	p.SetSeries(PlotSeries{Points: []PlotPoint{{-math.MaxFloat64, -math.MaxFloat64}, {math.MaxFloat64, math.MaxFloat64}}})
	p.Reset()
	check()
	renderView(p, 300, 1).Frame()
	p.SetSeries(PlotSeries{Points: []PlotPoint{{math.NaN(), 0}}})
	h := renderView(p, 220, 2)
	if !shown(h, "暂无数据") {
		t.Fatal("invalid-only data not empty")
	}
}

func TestPlotDisabledKeyboardAndNarrowLayout(t *testing.T) {
	p := Plot(PlotSeries{Points: []PlotPoint{{0, 0}, {10, 10}}})
	h := page(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(180)).Child(p.Render(cx)) }))
	click(t, h, "复位")
	before := [4]float64{p.x0, p.x1, p.y0, p.y1}
	h.Key("+", 0)
	if p.x1-p.x0 >= before[1]-before[0] {
		t.Fatal("keyboard zoom")
	}
	p.SetDisabled(true)
	h.Frame()
	before = [4]float64{p.x0, p.x1, p.y0, p.y1}
	h.Key(key.NameRightArrow, 0)
	h.Scroll(100, 140, -1000)
	if before != [4]float64{p.x0, p.x1, p.y0, p.y1} {
		t.Fatal("disabled view moved")
	}
}

func TestPlotCanceledDragRestoresView(t *testing.T) {
	p := Plot(PlotSeries{Points: []PlotPoint{{0, 0}, {10, 10}}})
	h := page(p)
	before := [4]float64{p.x0, p.x1, p.y0, p.y1}
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(180, 150)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(240, 160)})
	h.Frame()
	if !p.dragging || before == [4]float64{p.x0, p.x1, p.y0, p.y1} {
		t.Fatal("drag did not pan")
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
	h.Frame()
	if p.dragging || before != [4]float64{p.x0, p.x1, p.y0, p.y1} {
		t.Fatal("canceled drag changed view")
	}
}

func TestPlotHorizontalTicksPresentOnFirstFrame(t *testing.T) {
	p := Plot(PlotSeries{Points: []PlotPoint{{0, 100}, {10, 200}}})
	h := renderView(p, 400, 1)
	if !shown(h, "5") || !shown(h, "10") {
		t.Fatal("horizontal tick labels absent on first frame")
	}
}
