package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestCarouselContinuousLoopSeam(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			clicks, changes := 0, 0
			renders := [3]int{}
			slides := make([]el.View, 3)
			for i := range slides {
				slides[i] = el.ViewFunc(func(cx *el.Context) el.Element {
					renders[i]++
					name := []string{"first", "second", "last"}[i]
					return el.Div().Grow().Name(name).Child(el.Div().W(el.Dp(40)).H(el.Dp(24)).Name("action " + name).OnClick(func() { clicks++ }))
				})
			}
			car := Carousel(slides...).Gap(8).Height(200).Vertical(vertical).OnChange(func(int) { changes++ })
			h, advance := renderCarouselClock(el.ViewFunc(func(c *el.Context) el.Element { return car.Content().Render(c) }), 200, scale)
			for i := 0; i < 5; i++ {
				h.Frame()
			}
			// Each original view is rendered exactly once per build, including at the seam.
			before := renders
			h.Frame()
			for i := range renders {
				if renders[i]-before[i] != 1 {
					t.Fatal("duplicate mount", renders, before)
				}
			}
			point := func(along float32) f32.Point {
				if vertical {
					return f32.Pt(60*float32(scale), along*float32(scale))
				}
				return f32.Pt(along*float32(scale), 60*float32(scale))
			}
			h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: point(30)})
			h.Frame()
			h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: point(170)})
			h.Frame()
			h.Frame()
			if !car.drag.active || car.drag.offset != -140 {
				t.Fatal("backward edge did not continue", car.drag)
			}
			last, first := bounds(h, "last"), bounds(h, "first")
			if last.Empty() || first.Empty() {
				t.Fatal("seam has missing items", vertical, scale, last, first)
			}
			if vertical {
				if last.Max.Y+8*scale != first.Min.Y {
					t.Fatal("vertical seam spacing", last, first)
				}
			} else if last.Max.X+8*scale != first.Min.X {
				t.Fatal("horizontal seam spacing", last, first)
			}
			h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: point(170)})
			h.Frame()
			h.Frame()
			if car.Value() != 2 || changes != 1 || clicks != 0 {
				t.Fatal("release selection", car.Value(), changes, clicks)
			}
			advance(carouselTransition)
			click(t, h, "action last")
			if clicks != 1 {
				t.Fatal("translated hit area", clicks)
			}
			// Disabling looping rebases without changing selection or emitting a callback.
			car.Loop(false)
			h.Frame()
			h.Frame()
			if car.Value() != 2 || changes != 1 || bounds(h, "last").Empty() {
				t.Fatal("loop toggle lost item")
			}
		}
	}
}

func TestCarouselContinuousScrollCrossesBothEdges(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		car := Carousel(text("first"), text("second"), text("last")).Gap(8).Height(200).Vertical(vertical)
		c := &clock{now: time.Unix(100, 0)}
		h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(200)).Child(car.Content().Render(cx)) })
		for i := 0; i < 4; i++ {
			h.Frame()
		}
		scroll := func(delta float32) {
			d := f32.Pt(delta, 0)
			if vertical {
				d = f32.Pt(0, delta)
			}
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(40, 40), Scroll: d})
			h.Frame()
			c.advance(h, 150*time.Millisecond)
			h.Frame()
		}
		scroll(-300)
		if car.Value() != 2 {
			t.Fatal("backward wrap", vertical, car.Value(), car.scroll)
		}
		scroll(300)
		if car.Value() != 0 {
			t.Fatal("forward wrap", vertical, car.Value(), car.scroll)
		}
	}
}

func TestCarouselCircularGeometry(t *testing.T) {
	g := carouselGeometry{points: []float32{0, 208, 416}, maximum: 624, cycle: 624}
	for _, tc := range []struct {
		offset float32
		want   int
	}{{-170, 2}, {600, 0}, {800, 1}, {-800, 2}, {624*100 + 200, 1}} {
		if got := g.nearest(tc.offset, 0); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	// Grabbing an in-flight scroll continues at the visible location.
	v := Carousel(text("a"), text("b"), text("c"))
	v.scroll.active = true
	v.scroll.offset = -80
	v.handleDrag(el.DragEvent{Kind: el.DragStart, X: 30}, &g)
	v.handleDrag(el.DragEvent{Kind: el.DragMove, X: 40}, &g)
	if v.drag.offset != -90 || v.scroll.active {
		t.Fatal("gesture interruption jumped", v.drag, v.scroll)
	}
}

func TestCarouselShortLoopTrackFallsBack(t *testing.T) {
	car := Carousel(text("one"), text("two")).ItemsPerView(2).Gap(8)
	var cx *el.Context
	h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return car.Content().Render(c) }), 200, 1)
	for i := 0; i < 4; i++ {
		h.Frame()
	}
	offset, view, total := cx.ScrollStateX(autoID("carousel", car) + "/stage")
	if offset != 0 || total > view+1 {
		t.Fatal("short track gained loop runway", offset, view, total)
	}
	if bounds(h, "one").Empty() || bounds(h, "two").Empty() {
		t.Fatal("missing short-track items")
	}
	car.Next()
	car.Next()
	h.Frame()
	if car.Value() != 0 {
		t.Fatal("fallback lost index loop")
	}
}
