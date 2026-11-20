package el

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/unit"
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
		// Rows center by default: each box is centered in its line's height.
		want := []image.Point{{5, 10}, {115, 5}, {5, 50}, {115, 45}}
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

func TestAlignSelfInFlexWrapAndGrid(t *testing.T) {
	// Row with Items(Start): children override with Center, End, Stretch.
	a := Div().W(Dp(30)).H(Dp(20))
	b := Div().W(Dp(30)).H(Dp(20)).Self(Center)
	c := Div().W(Dp(30)).H(Dp(20)).Self(End)
	d := Div().W(Dp(30)).Self(Stretch)
	row := Div().Row().W(Dp(200)).H(Dp(100)).Items(Start).Child(a, b, c, d)
	render(t, Div().Items(Start).Child(row))
	if a.n.pos.Y != 0 || b.n.pos.Y != 40 || c.n.pos.Y != 80 || d.n.size.Y != 100 || d.n.pos.Y != 0 {
		t.Fatalf("row self: a=%v b=%v c=%v d=%v", rect(a), rect(b), rect(c), rect(d))
	}

	// Col (defaults to Stretch): child overrides with Center/End and does not stretch.
	ca := Div().H(Dp(20))
	cb := Div().W(Dp(40)).H(Dp(20)).Self(Center)
	cc := Div().W(Dp(40)).H(Dp(20)).Self(End)
	col := Div().Col().W(Dp(200)).Child(ca, cb, cc)
	render(t, Div().Items(Start).Child(col))
	if ca.n.size.X != 200 || cb.n.size.X != 40 || cb.n.pos.X != 80 || cc.n.pos.X != 160 {
		t.Fatalf("col self: ca=%v cb=%v cc=%v", rect(ca), rect(cb), rect(cc))
	}

	// Wrap with Items(Start): second child overrides with End.
	wa := Div().W(Dp(50)).H(Dp(40))
	wb := Div().W(Dp(50)).H(Dp(20)).Self(End)
	wrap := Div().Wrap().W(Dp(200)).Items(Start).Child(wa, wb)
	render(t, Div().Items(Start).Child(wrap))
	if wa.n.pos.Y != 0 || wb.n.pos.Y != 20 {
		t.Fatalf("wrap self: wa=%v wb=%v", rect(wa), rect(wb))
	}

	// Grid with Items(Start): one child stretches, another aligns End.
	ga := Div().H(Dp(60))
	gb := Div().Self(Stretch)
	gc := Div().H(Dp(20)).Self(End)
	grid := Div().Grid(3).W(Dp(300)).Items(Start).Child(ga, gb, gc)
	render(t, Div().Items(Start).Child(grid))
	if gb.n.size.Y != 60 || gc.n.pos.Y != 40 {
		t.Fatalf("grid self: ga=%v gb=%v gc=%v", rect(ga), rect(gb), rect(gc))
	}
}

func TestGridColsTracksAndGaps(t *testing.T) {
	// Fixed + Fr(1) + Fr(2) with separate ColGap and RowGap.
	a := Div().H(Dp(20))
	b := Div().H(Dp(20))
	c := Div().H(Dp(20))
	d := Div().H(Dp(30))
	grid := Div().GridCols(Dp(100), Fr(1), Fr(2)).W(Dp(420)).ColGap(10).RowGap(20).Child(a, b, c, d)
	render(t, Div().Items(Start).Child(grid))
	// Available for tracks: 420 - 2*10 = 400. Fixed 100 -> 300 left -> Fr(1)=100, Fr(2)=200.
	if a.n.size.X != 100 || b.n.size.X != 100 || c.n.size.X != 200 {
		t.Fatalf("track widths: a=%v b=%v c=%v", rect(a), rect(b), rect(c))
	}
	if b.n.pos.X != 110 || c.n.pos.X != 220 || d.n.pos != image.Pt(0, 40) {
		t.Fatalf("track positions: a=%v b=%v c=%v d=%v", rect(a), rect(b), rect(c), rect(d))
	}

	// Auto track sizes to intrinsic content width while Fr(1) takes the remainder.
	x := Div().W(Dp(60)).H(Dp(20))
	y := Div().H(Dp(20))
	autoGrid := Div().GridCols(Auto, Fr(1)).W(Dp(200)).ColGap(10).Child(x, y)
	render(t, Div().Items(Start).Child(autoGrid))
	if x.n.size.X != 60 || y.n.size.X != 130 || y.n.pos.X != 70 {
		t.Fatalf("auto+fr grid: x=%v y=%v", rect(x), rect(y))
	}
}

func TestGridRowSpanPlacementAndHeight(t *testing.T) {
	// 2-column grid:
	// [ A (RowSpan 2) ] [ B (h=30) ]
	// [ A (continued) ] [ C (h=40) ]
	// [ D (h=20)      ] [ E (h=20) ]
	a := Div().RowSpan(2)
	b := Div().H(Dp(30))
	c := Div().H(Dp(40))
	d := Div().H(Dp(20))
	e := Div().H(Dp(20))
	grid := Div().Grid(2).W(Dp(210)).ColGap(10).RowGap(10).Child(a, b, c, d, e)
	render(t, Div().Items(Start).Child(grid))
	// Row 0 height = 30, Row 1 height = 40 -> A spans 30 + 10 + 40 = 80.
	if a.n.pos != image.Pt(0, 0) || a.n.size != image.Pt(100, 80) {
		t.Fatalf("a=%v", rect(a))
	}
	if b.n.pos != image.Pt(110, 0) || c.n.pos != image.Pt(110, 40) {
		t.Fatalf("b=%v c=%v", rect(b), rect(c))
	}
	if d.n.pos != image.Pt(0, 90) || e.n.pos != image.Pt(110, 90) {
		t.Fatalf("d=%v e=%v", rect(d), rect(e))
	}

	// Tall RowSpan item expands its spanned rows when taller than single-row siblings.
	tall := Div().RowSpan(2).H(Dp(110))
	r0 := Div().H(Dp(20))
	r1 := Div().H(Dp(20))
	g2 := Div().Grid(2).W(Dp(210)).Gap(10).Child(tall, r0, r1)
	render(t, Div().Items(Start).Child(g2))
	// Single-row heights 20+20+10=50, deficit 60 -> +30 each -> rows become 50 and 50.
	if r1.n.pos.Y != 60 || g2.n.size.Y != 110 {
		t.Fatalf("tall rowspan: g2=%v r0=%v r1=%v", rect(g2), rect(r0), rect(r1))
	}
}
