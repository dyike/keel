package el

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSharedLayoutAcrossParentsRetargetInputAndAnchors(t *testing.T) {
	now := time.Unix(1600, 0)
	detail, clicks, renders := false, 0, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		renders++
		owner, x, size := "thumb", float32(20), float32(40)
		if detail {
			owner, x, size = "detail", 220, 80
		}
		card := Div().ID("card").Name("card").Size(Dp(size)).SharedLayout("photo", Tween(time.Second, Linear)).OnClick(func() { clicks++ })
		cx.Overlay("tip", Anchored("card", Div().Name("tip").Size(Dp(10))).Placement(Bottom, Start))
		return Div().Items(Start).Child(Div().ID(owner).Pl(x).Pt(20).Items(Start).Child(card))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail = true })
	h.Frame()
	if got := nodeBounds(h, "card"); got != image.Rect(20, 20, 60, 60) {
		t.Fatalf("handoff jumped: %v", got)
	}
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before {
		t.Fatal("shared motion rebuilt its view")
	}
	if got := nodeBounds(h, "card"); got != image.Rect(120, 20, 180, 80) {
		t.Fatalf("midpoint: %v", got)
	}
	if got := nodeBounds(h, "tip"); got.Min != image.Pt(120, 84) {
		t.Fatalf("anchor did not follow: %v", got)
	}
	h.Click(150, 40)
	if clicks != 1 {
		t.Fatal("moving destination did not receive input")
	}
	core.Update(func() { detail = false })
	h.Frame()
	if got := nodeBounds(h, "card"); got != image.Rect(120, 20, 180, 80) {
		t.Fatalf("reverse jumped: %v", got)
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if got := nodeBounds(h, "card"); got != image.Rect(20, 20, 60, 60) || root.motionDue {
		t.Fatalf("did not settle: %v", got)
	}
}

func TestSharedLayoutRotationAndAffineDecomposition(t *testing.T) {
	for _, angle := range []float32{-3, -.5, 0, 2, math.Pi} {
		m := f32.NewAffine2D(80, 22, 60, 0, 40, 30).Rotate(f32.Point{}, angle)
		v, ok := sharedComponents(m)
		if !ok {
			t.Fatal("valid affine transform rejected")
		}
		round := sharedTransform(v)
		for _, p := range []f32.Point{{}, {X: 1}, {Y: 1}, {X: 1, Y: 1}} {
			a, b := round.Transform(p), m.Transform(p)
			if math.Abs(float64(a.X-b.X))+math.Abs(float64(a.Y-b.Y)) > .0001 {
				t.Fatalf("affine round trip: %v, %v", a, b)
			}
		}
	}
	now := time.Unix(1700, 0)
	detail := false
	root := Root(ViewFunc(func(*Context) Element {
		owner, angle := "a", float32(170)
		if detail {
			owner, angle = "b", -170
		}
		return Div().P(100).Items(Start).Child(Div().ID(owner).Items(Start).AnimatePose(Pose{Rotate: angle}, Instant).Child(Div().Name("card").Size(Dp(40)).SharedLayout("rotated", Tween(time.Second, Linear))))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	old := root.e.sharedLayouts["rotated"].shown
	core.Update(func() { detail = true })
	h.Frame()
	new := root.e.sharedLayouts["rotated"].shown
	for _, p := range []f32.Point{{}, {X: 1, Y: 1}} {
		a, b := old.Transform(p), new.Transform(p)
		if math.Abs(float64(a.X-b.X))+math.Abs(float64(a.Y-b.Y)) > .001 {
			t.Fatal("nested rotation jumped on handoff")
		}
	}
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	v, _ := sharedComponents(root.e.sharedLayouts["rotated"].shown)
	if math.Abs(math.Abs(float64(v[2]))-math.Pi) > .001 || math.Abs(float64(v[3]-40)) > .001 {
		t.Fatalf("did not rotate the short way at constant size: %v", v)
	}
}

func TestSharedLayoutScrollResizeReducedMotionAndReclamation(t *testing.T) {
	oldReduced := theme.ReducedMotion
	defer theme.SetReducedMotion(oldReduced)
	now := time.Unix(1800, 0)
	detail, show, enabled := false, true, true
	scale := float32(1)
	viewport := image.Pt(400, 300)
	root := Root(ViewFunc(func(*Context) Element {
		owner, width := "a", float32(40)
		if detail {
			owner, width = "b", 80
		}
		list := Div().W(Dp(150)).H(Dp(120)).Items(Start).ScrollY()
		if show {
			card := Div().Name("card").W(Dp(width)).H(Dp(400)).NoShrink()
			if enabled {
				card.SharedLayout("card", Tween(time.Second, Linear))
			}
			list.Child(Div().ID(owner).Items(Start).NoShrink().Child(card))
		}
		return Div().Items(Start).Child(list)
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Constraints.Max = viewport
		gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
		root.Layout(gtx)
	})
	h.Scroll(20, 50, 30)
	if root.motionDue {
		t.Fatal("scroll started shared animation")
	}
	core.Update(func() { detail = true })
	h.Frame()
	if !root.motionDue {
		t.Fatal("handoff did not animate")
	}
	core.Update(func() { viewport.X = 420 })
	h.Frame()
	if root.motionDue || nodeBounds(h, "card").Dx() != 80 {
		t.Fatal("viewport resize did not snap")
	}
	core.Update(func() { detail = false; scale = 2 })
	h.Frame()
	if root.motionDue || nodeBounds(h, "card").Dx() != 80 {
		t.Fatal("density change did not snap")
	}
	theme.SetReducedMotion(true)
	core.Update(func() { detail = true })
	h.Frame()
	if root.motionDue {
		t.Fatal("reduced motion animated")
	}
	core.Update(func() { enabled = false })
	h.Frame()
	if len(root.e.sharedLayouts) != 0 {
		t.Fatal("omitted declaration retained geometry")
	}
	core.Update(func() { enabled = true })
	h.Frame()
	core.Update(func() { show = false })
	h.Frame()
	if root.e.sharedLayouts != nil {
		t.Fatal("removed content retained registry")
	}
}

func TestSharedLayoutPixelsMatchFullPaint(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Skip(err)
	}
	defer gpu.Release()
	now := time.Unix(1900, 0)
	detail := false
	root := Root(ViewFunc(func(*Context) Element {
		owner, x, size := "a", float32(30), float32(50)
		if detail {
			owner, x, size = "b", 230, 100
		}
		return Div().Items(Start).Child(Div().ID(owner).Pl(x).Pt(60).Items(Start).Child(Div().Size(Dp(size)).Rounded(size*.16).Bg(color.NRGBA{R: 200, A: 255}).SharedLayout("photo", Tween(time.Second, Linear))))
	}))
	root.RetainFrames()
	var shot *image.RGBA
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		shot = image.NewRGBA(image.Rect(0, 0, 400, 300))
		if err := gpu.Screenshot(shot); err != nil {
			t.Fatal(err)
		}
	})
	initial := shot
	core.Update(func() { detail = true })
	h.Frame()
	assertSharedPixelContinuity(t, initial, shot)
	now = now.Add(350 * time.Millisecond)
	h.Frame()
	mid := shot
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("retained/full pixels differ")
	}
	core.Update(func() { detail = false })
	h.Frame()
	assertSharedPixelContinuity(t, mid, shot)
}

// Scaling an equivalent rounded path can change edge quantization by two
// channel levels. Rebuilding at an unchanged size remains byte-exact above.
func assertSharedPixelContinuity(t *testing.T, a, b *image.RGBA) {
	t.Helper()
	for i, value := range a.Pix {
		delta := int(value) - int(b.Pix[i])
		if delta < 0 {
			delta = -delta
		}
		if delta > 2 {
			t.Fatalf("shared handoff changed pixel byte %d: %d/%d", i, value, b.Pix[i])
		}
	}
}

func TestSharedLayoutPreservesDestinationClip(t *testing.T) {
	now := time.Unix(2000, 0)
	detail, clicks := false, 0
	root := Root(ViewFunc(func(*Context) Element {
		owner, x := "a", float32(0)
		if detail {
			owner, x = "b", 180
		}
		return Div().Items(Start).Child(Div().ID(owner).Size(Dp(80)).Translate(x, 0).OnClick(func() {}).Items(Start).Child(
			Div().ID("card").Name("card").Size(Dp(40)).SharedLayout("clip", Tween(time.Second, Linear)).OnClick(func() { clicks++ }),
		))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail = true })
	h.Frame()
	h.Click(20, 20)
	if clicks != 0 {
		t.Fatal("shared content escaped destination ancestor clip")
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	h.Click(200, 20)
	if clicks != 1 {
		t.Fatal("settled content not interactive")
	}
}

func TestSharedLayoutSpringRetargetVelocity(t *testing.T) {
	now := time.Unix(2100, 0)
	detail := false
	root := Root(ViewFunc(func(*Context) Element {
		owner, x := "a", float32(0)
		if detail {
			owner, x = "b", 200
		}
		return Div().Items(Start).Child(Div().ID(owner).Pl(x).Items(Start).Child(Div().Size(Dp(40)).SharedLayout("spring", Snappy)))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail = true })
	h.Frame()
	now = now.Add(100 * time.Millisecond)
	h.Frame()
	s := root.e.sharedLayouts["spring"]
	before, velocity := s.shown, s.values[0].Velocity()
	core.Update(func() { detail = false })
	h.Frame()
	if math.Abs(float64(s.values[0].Velocity()-velocity)) > .001 || velocity == 0 {
		t.Fatalf("interruption lost spring velocity: %v -> %v", velocity, s.values[0].Velocity())
	}
	for _, p := range []f32.Point{{}, {X: 1, Y: 1}} {
		a, b := before.Transform(p), s.shown.Transform(p)
		if math.Abs(float64(a.X-b.X))+math.Abs(float64(a.Y-b.Y)) > .001 {
			t.Fatal("spring interruption jumped")
		}
	}
}

func TestSharedLayoutDuplicateIdentityRejected(t *testing.T) {
	root := Root(ViewFunc(func(*Context) Element {
		return Div().Row().Child(Div().Size(Dp(40)).SharedLayout("duplicate", Smooth), Div().Size(Dp(40)).SharedLayout("duplicate", Smooth))
	}))
	var caught any
	uitest.NewFunc(func(gtx core.C) {
		// Recover inside the frame callback so the harness can release its lock.
		defer func() { caught = recover() }()
		root.Layout(gtx)
	})
	if caught == nil {
		t.Fatal("duplicate identity accepted")
	}
}
