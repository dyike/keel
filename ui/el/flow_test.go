package el

import (
	"image"
	"testing"

	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestWrapBreaksLinesAndKeepsHitCoordinates(t *testing.T) {
	for _, scale := range []int{1, 2} {
		width := float32(230)
		calls := 0
		var boxes []*DivEl
		root := Root(viewFunc(func(cx *Context) Element {
			boxes = nil
			flow := Div().Wrap().W(Dp(width)).Gap(10).P(5)
			for i := 0; i < 4; i++ {
				box := Div().W(Dp(100)).H(Dp(float32(20 + i*10))).NoShrink()
				if i == 2 {
					box.OnClick(func() { calls++ })
				}
				boxes = append(boxes, box)
				flow.Child(box)
			}
			return Div().Items(Start).Child(flow)
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 400*scale)
			root.Layout(gtx)
		})
		want := []image.Point{{5, 5}, {115, 5}, {5, 45}, {115, 45}}
		for i, p := range want {
			if boxes[i].n.pos != p.Mul(scale) {
				t.Fatalf("scale %d child %d: %v", scale, i, boxes[i].n.pos)
			}
		}
		h.Click(float32(20*scale), float32(55*scale))
		if calls != 1 {
			t.Fatal("wrapped hit mismatch")
		}
		width = 120
		h.Frame()
		if boxes[1].n.pos != image.Pt(5, 35).Mul(scale) || boxes[3].n.pos != image.Pt(5, 125).Mul(scale) {
			t.Fatalf("narrow lines: %v %v", boxes[1].n.pos, boxes[3].n.pos)
		}
	}
}

func TestWrapPerLineFlexAndAlignment(t *testing.T) {
	a := Div().W(Dp(120)).H(Dp(20)).NoShrink()
	b := Div().Flex(1).MinW(Dp(40)).H(Dp(40))
	c := Div().W(Dp(100)).H(Dp(20)).NoShrink()
	flow := Div().Wrap().W(Dp(200)).Gap(10).Items(Center).Justify(End).Child(a, b, c)
	render(t, Div().Items(Start).Child(flow))
	if a.n.pos.Y != 10 || b.n.size.X != 70 || c.n.pos != image.Pt(100, 50) {
		t.Fatalf("a %v b %v c %v", rect(a), rect(b), rect(c))
	}
}

func TestGridTracksMarginsMinimumsAndRows(t *testing.T) {
	a := Div().H(Dp(20)).Mx(5)
	b := Div().H(Dp(40))
	c := Div().H(Dp(10))
	d := Div().H(Dp(30))
	grid := Div().Grid(3).W(Dp(320)).Gap(10).P(5).Items(Center).Child(a, b, c, d)
	render(t, Div().Items(Start).Child(grid))
	if a.n.size.X != 86 || a.n.pos != image.Pt(10, 15) || b.n.pos != image.Pt(111, 5) || c.n.pos != image.Pt(218, 20) || d.n.pos != image.Pt(5, 55) {
		t.Fatalf("grid: %v %v %v %v", rect(a), rect(b), rect(c), rect(d))
	}
	// A wide minimum takes from the other track before the grid overflows.
	x, y := Div().MinW(Dp(140)).H(Dp(10)), Div().H(Dp(10))
	render(t, Div().Items(Start).Child(Div().Grid(2).W(Dp(200)).Gap(10).Child(x, y)))
	if x.n.size.X != 140 || y.n.size.X != 50 || y.n.pos.X != 150 {
		t.Fatalf("min tracks %v %v", rect(x), rect(y))
	}
	y.MinW(Dp(100))
	render(t, Div().Items(Start).Child(Div().Grid(2).W(Dp(200)).Gap(10).Child(x, y)))
	if x.n.size.X != 140 || y.n.size.X != 100 {
		t.Fatal("grid violated minimum width")
	}
}

func TestGridWrapHiddenAbsoluteAndModeChange(t *testing.T) {
	for _, grid := range []bool{false, true} {
		a, b := Div().Size(Dp(40)), Div().Size(Dp(40))
		hidden := Div().W(Dp(300)).Hidden(true)
		absolute := Div().Absolute().Right(0).Top(0).Size(Dp(10))
		flow := Div().Wrap().W(Dp(100)).Gap(10)
		if grid {
			flow.Grid(2)
		}
		flow.Child(a, hidden, b, absolute)
		render(t, Div().Items(Start).Child(flow))
		if a.n.pos.X != 0 || b.n.pos.Y != 0 || absolute.n.pos.X != 90 {
			t.Fatalf("grid %v: %v %v %v", grid, rect(a), rect(b), rect(absolute))
		}
		flow.Col()
		render(t, Div().Items(Start).Child(flow))
		if b.n.pos.Y != 50 {
			t.Fatal("Col did not reset flow mode")
		}
	}
}

func TestGridColumnSpansAndRowBreaks(t *testing.T) {
	a, b, c, d := Div().H(Dp(20)).ColSpan(2), Div().H(Dp(40)), Div().H(Dp(10)).ColSpan(2), Div().H(Dp(30)).ColSpan(2)
	grid := Div().Grid(3).W(Dp(320)).Gap(10).Child(a, b, c, d)
	render(t, Div().Items(Start).Child(grid))
	if a.n.size.X != 210 || b.n.pos != image.Pt(220, 0) || c.n.pos != image.Pt(0, 50) || d.n.pos != image.Pt(0, 70) {
		t.Fatalf("spans: %v %v %v %v", rect(a), rect(b), rect(c), rect(d))
	}
	d.ColSpan(99)
	render(t, Div().Items(Start).Child(grid))
	if d.n.size.X != 320 {
		t.Fatal("span not clamped to columns")
	}
	a.ColSpan(0)
	render(t, Div().Items(Start).Child(grid))
	if a.n.size.X != 100 || b.n.pos.X != 110 {
		t.Fatal("zero span is not one")
	}
}

func TestGridSpanningMinimumAndStretch(t *testing.T) {
	a := Div().ColSpan(2).MinW(Dp(250)).Mx(5).Child(Div().H(Dp(10)))
	b := Div().H(Dp(40))
	grid := Div().Grid(3).W(Dp(300)).Gap(10).Child(a, b)
	render(t, Div().Items(Start).Child(grid))
	if a.n.size.X != 250 || a.n.size.Y != 40 || b.n.size.X != 30 || b.n.pos.X != 270 {
		t.Fatalf("minimum/stretch: %v %v", rect(a), rect(b))
	}
	a.Hidden(true)
	render(t, Div().Items(Start).Child(grid))
	if b.n.pos.X != 0 {
		t.Fatal("hidden span reserved tracks")
	}
}

func TestWrapFitHugsContentAndHonorsWidth(t *testing.T) {
	a, b := Div().Size(Dp(60)), Div().Size(Dp(40))
	flow := Div().WrapFit().Gap(4).Child(a, b)
	render(t, Div().Items(Start).Child(flow))
	if flow.n.size.X != 104 || b.n.pos.X != 64 {
		t.Fatal("fit width", rect(flow), rect(b))
	}
	flow.MaxW(Dp(80))
	render(t, Div().Items(Start).Child(flow))
	if flow.n.size.X != 60 || b.n.pos.Y != 64 {
		t.Fatal("fit wrap", rect(flow), rect(b))
	}
	flow.MaxW(Full).W(Dp(200)).Justify(End)
	render(t, Div().Items(Start).Child(flow))
	if flow.n.size.X != 200 || a.n.pos.X != 96 {
		t.Fatal("explicit width alignment", rect(flow), rect(a))
	}
	flow.W(Auto).Wrap()
	render(t, Div().Items(Start).Child(Div().W(Dp(200)).Items(Start).Child(flow)))
	if flow.n.size.X != 200 {
		t.Fatal("Wrap did not clear fit", rect(flow))
	}
}
