package el

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestScrollXWheelRevealShrinkAndHit(t *testing.T) {
	for _, scale := range []int{1, 2} {
		width := float32(360)
		calls := 0
		var cx *Context
		root := Root(viewFunc(func(c *Context) Element {
			cx = c
			return Div().Items(Start).Child(Div().ID("horizontal").W(Dp(120)).H(Dp(60)).ScrollX().Child(
				Div().Row().W(Dp(width)).Child(Div().W(Dp(240)).NoShrink(), Div().W(Dp(120)).H(Dp(40)).OnClick(func() { calls++ }).Child(Text("end")))))
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
			root.Layout(gtx)
		})
		h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(30*float32(scale), 20*float32(scale)), Scroll: f32.Pt(240*float32(scale), 0)})
		h.Frame()
		h.Frame()
		x, view, total := cx.ScrollStateX("horizontal")
		if x != 240 || view != 120 || total != 360 {
			t.Fatalf("scale %d: %g %g %g", scale, x, view, total)
		}
		h.Click(float32(30*scale), float32(20*scale))
		h.Frame()
		if calls != 1 {
			t.Fatal("scrolled content hit coordinates incorrect")
		}
		cx.ScrollIntoViewX("horizontal", 0, 20)
		h.Frame()
		if x, _, _ := cx.ScrollStateX("horizontal"); x != 0 {
			t.Fatal("reveal failed")
		}
		cx.ScrollIntoViewX("horizontal", 300, 320)
		h.Frame()
		width = 100
		h.Frame()
		h.Frame()
		if x, _, _ := cx.ScrollStateX("horizontal"); x != 0 {
			t.Fatalf("shrink retained offset %g", x)
		}
	}
}

func TestScrollXAndYIndependent(t *testing.T) {
	var cx *Context
	h := uitest.New(Root(viewFunc(func(c *Context) Element {
		cx = c
		return Div().Items(Start).Child(
			Div().ID("both").W(Dp(100)).H(Dp(100)).ScrollX().ScrollY().Child(Div().W(Dp(400)).H(Dp(400))))
	})))
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(30, 30), Scroll: f32.Pt(40, 60)})
	h.Frame()
	h.Frame()
	x, _, _ := cx.ScrollStateX("both")
	y, _, _ := cx.ScrollState("both")
	if x != 40 || y != 60 {
		t.Fatalf("axes: %g %g", x, y)
	}
}

func TestScrollXNestedAndReadOnly(t *testing.T) {
	readonly := false
	var cx *Context
	root := Root(viewFunc(func(c *Context) Element {
		cx = c
		return Div().Items(Start).Child(Div().W(Dp(120)).H(Dp(100)).ID("vertical").ScrollY().Child(
			Div().ID("horizontal").H(Dp(60)).ScrollX().Child(Div().W(Dp(400)).H(Dp(50))), Div().H(Dp(400))))
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		if readonly {
			gtx = gtx.Disabled()
		}
		root.Layout(gtx)
	})
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(20, 20), Scroll: f32.Pt(80, 0)})
	h.Frame()
	h.Frame()
	x, _, _ := cx.ScrollStateX("horizontal")
	y, _, _ := cx.ScrollState("vertical")
	if x != 80 || y != 0 {
		t.Fatalf("nested axes %g %g", x, y)
	}
	readonly = true
	for range 3 {
		h.Frame()
	}
	readonly = false
	h.Frame()
	if x, _, _ := cx.ScrollStateX("horizontal"); x != 80 {
		t.Fatal("readonly reset scroll")
	}
	h.Scroll(20, 20, 40)
	h.Frame()
	if y, _, _ := cx.ScrollState("vertical"); y != 40 {
		t.Fatalf("vertical parent did not receive wheel: %g", y)
	}
}

func TestScrollXStretchesNarrowColumns(t *testing.T) {
	child := Div().MinW(Dp(120)).H(Dp(20))
	box := Div().W(Dp(300)).ScrollX().Child(child)
	render(t, Div().Items(Start).Child(box))
	if child.n.size.X != 300 || box.n.contentW != 300 {
		t.Fatalf("child %v content %d", child.n.size, box.n.contentW)
	}
	child.MinW(Dp(500))
	render(t, Div().Items(Start).Child(box))
	if child.n.size.X != 500 || box.n.contentW != 500 {
		t.Fatal("wide child shrank")
	}
}

func TestFlexRedistributesMinimumSizes(t *testing.T) {
	for _, row := range []bool{true, false} {
		a, b := Div().Flex(1), Div().Flex(2)
		box := Div().W(Dp(100)).H(Dp(100))
		if row {
			box.Row()
			a.MinW(Dp(40))
			b.MinW(Dp(40))
		} else {
			a.MinH(Dp(40))
			b.MinH(Dp(40))
		}
		box.Child(a, b)
		render(t, Div().Items(Start).Child(box))
		if mainOf(a.n.size, row) != 40 || mainOf(b.n.size, row) != 60 {
			t.Fatalf("row %v: %v %v", row, a.n.size, b.n.size)
		}
	}
}
