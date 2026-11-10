package el

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestLayoutMotionReorderSizeAndInput(t *testing.T) {
	now := time.Unix(1200, 0)
	swap, clicks, renders := false, 0, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		renders++
		a := Div().ID("a").Name("a").Size(Dp(40)).AnimateLayout(Tween(time.Second, Linear)).OnClick(func() { clicks++ })
		width := float32(40)
		if swap {
			width = 80
		}
		b := Div().ID("b").Name("b").W(Dp(width)).H(Dp(40)).AnimateLayout(Tween(time.Second, Linear))
		cx.Overlay("tip", Anchored("a", Div().Name("tip").Size(Dp(10))).Placement(Bottom, Start))
		if swap {
			return Div().Row().Items(Start).Child(b, a)
		}
		return Div().Row().Items(Start).Child(a, b)
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { swap = true })
	h.Frame()
	if got := nodeBounds(h, "a"); got != image.Rect(0, 0, 40, 40) {
		t.Fatalf("jump on reorder: %v", got)
	}
	if got := nodeBounds(h, "b"); got != image.Rect(40, 0, 80, 40) {
		t.Fatalf("jump on size change: %v", got)
	}
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before || root.repaints != 1 {
		t.Fatal("layout animation rebuilt view")
	}
	if got := nodeBounds(h, "a"); got != image.Rect(40, 0, 80, 40) {
		t.Fatalf("midpoint a: %v", got)
	}
	if got := nodeBounds(h, "b"); got != image.Rect(20, 0, 80, 40) {
		t.Fatalf("midpoint b: %v", got)
	}
	if got := nodeBounds(h, "tip"); got.Min != image.Pt(40, 44) {
		t.Fatalf("anchor: %v", got)
	}
	h.Click(60, 20)
	if clicks != 1 {
		t.Fatal("moving item did not receive click")
	}
	// Reverse in flight; the current position is continuous.
	core.Update(func() { swap = false })
	h.Frame()
	if got := nodeBounds(h, "a"); got.Min.X != 40 {
		t.Fatalf("retarget jumped: %v", got)
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if got := nodeBounds(h, "a"); got.Min.X != 0 || root.motionDue {
		t.Fatalf("did not settle: %v", got)
	}
}

func TestLayoutMotionScrollDensityAndReclamation(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	now := time.Unix(1300, 0)
	width, scale := float32(40), float32(1)
	show, animate := true, true
	root := Root(ViewFunc(func(*Context) Element {
		list := Div().ID("scroll").W(Dp(160)).H(Dp(100)).ScrollY().Items(Start)
		if show {
			box := Div().ID("box").Name("box").W(Dp(width)).H(Dp(500)).NoShrink()
			if animate {
				box.AnimateLayout(Smooth)
			}
			list.Child(box)
		}
		return Div().Items(Start).Child(list)
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
		root.Layout(gtx)
	})
	h.Scroll(20, 50, 40)
	if root.motionDue {
		t.Fatal("scroll started layout animation")
	}
	core.Update(func() { width = 80 })
	h.Frame()
	if !root.motionDue {
		t.Fatal("resize did not animate")
	}
	core.Update(func() { scale = 2 })
	h.Frame()
	if got := nodeBounds(h, "box"); got.Dx() != 160 || root.motionDue {
		t.Fatalf("density change did not snap: %v", got)
	}
	theme.SetReducedMotion(true)
	core.Update(func() { width = 60 })
	h.Frame()
	if got := nodeBounds(h, "box"); got.Dx() != 120 || root.motionDue {
		t.Fatalf("reduced motion: %v", got)
	}
	core.Update(func() { animate = false })
	h.Frame()
	for _, st := range root.store.states {
		if st.layoutMotion != nil {
			t.Fatal("omitted animation retained state")
		}
	}
	core.Update(func() { animate = true })
	h.Frame()
	core.Update(func() { show = false })
	h.Frame()
	for _, st := range root.store.states {
		if st.layoutMotion != nil {
			t.Fatal("removed element retained state")
		}
	}
}

func TestLayoutMotionComposesWithPose(t *testing.T) {
	now := time.Unix(1400, 0)
	width := float32(40)
	root := Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Div().Name("box").ID("box").W(Dp(width)).H(Dp(40)).AnimateLayout(Tween(time.Second, Linear)).AnimatePose(Pose{X: 10, Scale: 1.5}, Instant))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { width = 80 })
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	got := nodeBounds(h, "box")
	// Gio rounds the two half-pixel edges outward: -7.5 and 82.5.
	if got != image.Rect(-8, -10, 83, 50) {
		t.Fatalf("combined layout/pose bounds: %v", got)
	}
}

func TestLayoutMotionEnteringClipRequestsFrames(t *testing.T) {
	now := time.Unix(1450, 0)
	spacer := float32(160)
	root := Root(ViewFunc(func(*Context) Element {
		return Div().W(Dp(100)).H(Dp(100)).ScrollY().Items(Start).Child(Div().H(Dp(spacer)).NoShrink(), Div().ID("incoming").Name("incoming").Size(Dp(30)).NoShrink().AnimateLayout(Tween(time.Second, Linear)))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { spacer = 0 })
	h.Frame()
	if !root.motionDue {
		t.Fatal("offscreen layout arrival requested no frame")
	}
	now = now.Add(700 * time.Millisecond)
	h.Frame()
	if got := nodeBounds(h, "incoming"); got.Empty() {
		t.Fatal("arrival remained outside clip")
	}
}
