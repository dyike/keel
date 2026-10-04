package kit

import (
	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func renderCarouselClock(v el.View, width, scale int) (*uitest.Harness, func(time.Duration)) {
	now := time.Unix(100, 0)
	root := el.Embed(v)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
		gtx.Constraints.Max = image.Pt(width*scale, 1000*scale)
		root.Layout(gtx)
	})
	return h, func(d time.Duration) { now = now.Add(d); h.Frame() }
}

func TestCarouselNavigationMotion(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			calls := 0
			car := Carousel(text("one"), text("two"), text("three")).Gap(8).Height(200).Vertical(vertical).OnChange(func(int) { calls++ })
			h, advance := renderCarouselClock(car.Content(), 200, scale)
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			car.Next()
			h.Frame()
			if car.Value() != 1 || calls != 1 || car.motion.offset != 0 || !car.motion.running {
				t.Fatal("initial transition", car.Value(), calls, car.motion)
			}
			advance(90 * time.Millisecond)
			mid := car.motion.offset
			if mid <= 0 || mid >= 208 {
				t.Fatal("no interpolation", mid)
			}
			car.Next()
			h.Frame()
			if car.motion.offset != mid || car.motion.to != 416 {
				t.Fatal("redirect jumped", mid, car.motion)
			}
			advance(carouselTransition)
			if car.motion.offset != 416 || car.motion.running || calls != 2 {
				t.Fatal("completion", car.motion, calls)
			}
			car.Next()
			h.Frame()
			if car.motion.to != 624 || car.motion.from != 416 {
				t.Fatal("forward seam reverses", car.motion)
			}
			advance(90 * time.Millisecond)
			if car.motion.offset <= 416 || car.motion.offset >= 624 {
				t.Fatal("forward seam missing", car.motion)
			}
			advance(90 * time.Millisecond)
			if car.motion.offset != 0 || car.motion.running || bounds(h, "one").Empty() {
				t.Fatal("rebase", car.motion)
			}
			car.Previous()
			h.Frame()
			if car.motion.to != -208 {
				t.Fatal("backward seam reverses", car.motion)
			}
			advance(carouselTransition)
			if car.motion.offset != 416 || bounds(h, "three").Empty() {
				t.Fatal("backward rebase", car.motion)
			}
			// Programmatic selection cancels even when it repeats the selected index.
			car.Next()
			h.Frame()
			advance(50 * time.Millisecond)
			car.SetValue(0)
			h.Frame()
			if car.motion.offset != 0 || car.motion.running || car.motion.pending || calls != 5 {
				t.Fatal("programmatic reset", car.motion, calls)
			}
		}
	}
}

func TestCarouselTwoItemForwardMotionAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	car := Carousel(text("one"), text("two")).Gap(8)
	h, advance := renderCarouselClock(car.Content(), 200, 1)
	for i := 0; i < 4; i++ {
		h.Frame()
	}
	car.Next()
	h.Frame()
	if car.motion.to != 208 {
		t.Fatal("two-item next went backwards", car.motion)
	}
	advance(50 * time.Millisecond)
	theme.SetReducedMotion(true)
	h.Frame()
	if car.motion.offset != 208 || car.motion.running {
		t.Fatal("reduced motion did not settle", car.motion)
	}
	car.Next()
	h.Frame()
	if car.motion.offset != 0 || car.motion.running {
		t.Fatal("reduced motion animated", car.motion)
	}
	theme.SetReducedMotion(false)
	car.Next()
	h.Frame()
	advance(50 * time.Millisecond)
	car.SetDisabled(true)
	h.Frame()
	if car.motion.offset != 208 || car.motion.running {
		t.Fatal("disable did not settle", car.motion)
	}
}

func TestCarouselMotionGestureInterruptionAndResize(t *testing.T) {
	car := Carousel(text("one"), text("two"), text("three")).Gap(8).Loop(false).Height(200)
	width := float32(200)
	h, advance := renderCarouselClock(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(el.Div().W(el.Dp(width)).Child(car.Content().Render(cx)))
	}), 300, 1)
	for i := 0; i < 4; i++ {
		h.Frame()
	}
	car.Next()
	h.Frame()
	advance(50 * time.Millisecond)
	mid := car.motion.offset
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 60)})
	h.Frame()
	if car.drag.origin != mid || car.motion.running {
		t.Fatal("drag did not take over animation", mid, car.drag, car.motion)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(50, 60)})
	h.Frame()
	dragged := car.drag.offset
	if dragged != mid-20 {
		t.Fatal("drag jumped", mid, dragged)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(50, 60)})
	h.Frame()
	h.Frame()
	if !car.motion.running || car.motion.from != dragged {
		t.Fatal("release did not animate from drag", dragged, car.motion)
	}
	advance(50 * time.Millisecond)
	mid = car.motion.offset
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(60, 60), Scroll: f32.Pt(30, 0)})
	h.Frame()
	if car.scroll.offset != mid+30 || car.motion.running {
		t.Fatal("scroll did not take over animation", mid, car.scroll, car.motion)
	}
	advance(150 * time.Millisecond)
	h.Frame()
	if !car.motion.running {
		t.Fatal("idle snap did not animate", car.motion)
	}
	width = 240
	h.Frame()
	h.Frame()
	h.Frame()
	if car.motion.running || car.motion.offset != car.motion.geometry.points[car.current] || car.motion.geometry.viewport != 240 {
		t.Fatal("resize retained stale geometry", car.motion)
	}
}

func TestCarouselLoopMotionPixels(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			red := color.NRGBA{R: 255, A: 255}
			blue := color.NRGBA{B: 255, A: 255}
			slide := func(c color.NRGBA) el.View {
				return el.ViewFunc(func(*el.Context) el.Element { return el.Div().Grow().Bg(c) })
			}
			car := Carousel(slide(red), slide(color.NRGBA{G: 255, A: 255}), slide(blue)).Height(200).Gap(8).Vertical(vertical)
			car.SetValue(2)
			now := time.Unix(100, 0)
			gpu, err := headless.NewWindow(200*scale, 200*scale)
			if err != nil {
				t.Fatal(err)
			}
			root := el.Embed(car.Content())
			img := image.NewRGBA(image.Rect(0, 0, 200*scale, 200*scale))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints = layout.Exact(image.Pt(200*scale, 200*scale))
				root.Layout(gtx)
				if err := gpu.Frame(gtx.Ops); err != nil {
					t.Fatal(err)
				}
				if err := gpu.Screenshot(img); err != nil {
					t.Fatal(err)
				}
			})
			for i := 0; i < 4; i++ {
				h.Frame()
			}
			pixel := func(along int) color.RGBA {
				if vertical {
					return img.RGBAAt(100*scale, along*scale)
				}
				return img.RGBAAt(along*scale, 100*scale)
			}
			if pixel(100) != (color.RGBA{B: 255, A: 255}) {
				t.Fatal("initial last slide", pixel(100))
			}
			car.Next()
			h.Frame()
			now = now.Add(90 * time.Millisecond)
			h.Frame()
			if pixel(10) != (color.RGBA{B: 255, A: 255}) || pixel(100) != (color.RGBA{R: 255, A: 255}) {
				t.Fatal("seam animation pixels", scale, vertical, pixel(10), pixel(100))
			}
			now = now.Add(90 * time.Millisecond)
			h.Frame()
			if pixel(10) != (color.RGBA{R: 255, A: 255}) || pixel(190) != (color.RGBA{R: 255, A: 255}) {
				t.Fatal("final rebase pixels", pixel(10), pixel(190))
			}
			gpu.Release()
		}
	}
}

func TestCarouselSameIndexSnapAndCancellationAnimate(t *testing.T) {
	calls := 0
	car := Carousel(text("one"), text("two")).Loop(false).OnChange(func(int) { calls++ })
	h, advance := renderCarouselClock(car.Content(), 200, 1)
	for i := 0; i < 4; i++ {
		h.Frame()
	}
	for _, canceled := range []bool{false, true} {
		g := car.motion.geometry
		car.handleDrag(el.DragEvent{Kind: el.DragStart, X: 100}, g)
		car.handleDrag(el.DragEvent{Kind: el.DragMove, X: 50}, g)
		h.Frame()
		car.handleDrag(el.DragEvent{Kind: el.DragEnd, X: 50, Canceled: canceled}, g)
		h.Frame()
		if car.Value() != 0 || calls != 0 || !car.motion.running || car.motion.from != 50 || car.motion.to != 0 {
			t.Fatal("same index/cancel snap", canceled, car.motion, calls)
		}
		advance(carouselTransition)
		if car.motion.offset != 0 || car.motion.running {
			t.Fatal("snap did not settle", car.motion)
		}
	}
	car.Draggable(false).Scrollable(false)
	car.Next()
	h.Frame()
	if !car.motion.running {
		t.Fatal("disabling gestures disabled navigation animation")
	}
}
