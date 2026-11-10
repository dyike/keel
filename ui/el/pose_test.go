package el

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestPosePaintOnlyRoutesInputAndAnchors(t *testing.T) {
	now := time.Unix(800, 0)
	var target float32
	renders, clicks := 0, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		renders++
		cx.Overlay("tip", Anchored("card", Div().Name("tip").W(Dp(30)).H(Dp(10))).Placement(Bottom, Start))
		return Div().Items(Start).Child(Div().ID("card").Name("card").Size(Dp(40)).AnimatePose(Pose{X: target}, Tween(time.Second, Linear)).OnClick(func() { clicks++ }))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { target = 100 })
	h.Frame()
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before || root.repaints != 1 {
		t.Fatalf("pose rebuilt view: renders=%d before=%d repaints=%d", renders, before, root.repaints)
	}
	if got := nodeBounds(h, "card"); got != image.Rect(50, 0, 90, 40) {
		t.Fatalf("card bounds=%v", got)
	}
	if got := nodeBounds(h, "tip"); got.Min != image.Pt(50, 44) {
		t.Fatalf("overlay did not follow pose: %v", got)
	}
	h.Click(60, 20)
	if clicks != 1 {
		t.Fatalf("posed hit area: %d clicks", clicks)
	}
	now = now.Add(time.Second)
	h.Frame()
	h.Frame()
	if root.motionDue {
		t.Fatal("settled pose requested another frame")
	}
}

func TestNestedPoseGeometryAndClipping(t *testing.T) {
	clicks := 0
	y := float32(80)
	h := uitest.New(Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(
			Div().Size(Dp(100)).Items(Start).AnimatePose(Pose{X: 100, Y: y, Scale: 2}, Smooth).Child(
				Div().ID("nested").Name("nested").W(Dp(30)).H(Dp(10)).AnimatePose(Pose{X: 10, Rotate: 90}, Smooth).OnClick(func() { clicks++ }),
			),
		)
	})))
	got := nodeBounds(h, "nested")
	if !image.Pt(100, 40).In(got) || got.Dx() > 22 || got.Dy() > 62 {
		t.Fatalf("nested transformed bounds: %v", got)
	}
	h.Click(100, 40)
	if clicks != 1 {
		t.Fatalf("nested transformed hit area: %d clicks", clicks)
	}
	// Suppress motion to move the child fully outside the window.
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(true)
	y = 0
	h.Frame()
	if got := nodeBounds(h, "nested"); !got.Empty() {
		t.Fatalf("offscreen posed child remained visible: %v", got)
	}
}

func TestPoseStateReclaimedAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	now := time.Unix(900, 0)
	show, posed := true, true
	target := float32(0)
	root := Root(ViewFunc(func(*Context) Element {
		box := Div().Items(Start)
		if show {
			card := Div().ID("card").Name("card").Size(Dp(20))
			if posed {
				card.AnimatePose(Pose{X: target}, Smooth)
			}
			box.Child(card)
		}
		return box
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { target = 80 })
	h.Frame()
	if !root.motionDue {
		t.Fatal("retarget did not start")
	}
	theme.SetReducedMotion(true)
	now = now.Add(time.Millisecond)
	h.Frame()
	if got := nodeBounds(h, "card"); got.Min.X != 80 {
		t.Fatalf("reduced motion: %v", got)
	}
	if root.motionDue {
		t.Fatal("reduced motion requested frames")
	}
	core.Update(func() { posed = false })
	h.Frame()
	for _, st := range root.store.states {
		if st.pose != nil {
			t.Fatal("omitted pose retained channels")
		}
	}
	core.Update(func() { posed = true })
	h.Frame()
	core.Update(func() { show = false })
	h.Frame()
	for _, st := range root.store.states {
		if st.pose != nil {
			t.Fatal("removed element retained channels")
		}
	}
}

func TestPoseRepaintPixelsMatchFullFrame(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Skip(err)
	}
	defer gpu.Release()
	now := time.Unix(1000, 0)
	target := Pose{}
	width := float32(70)
	renders := 0
	root := Root(ViewFunc(func(*Context) Element {
		renders++
		return Div().P(60).Items(Start).Child(Div().W(Dp(width)).H(Dp(70)).Rounded(12).Bg(color.NRGBA{R: 200, A: 255}).AnimateLayout(Tween(time.Second, Linear)).AnimatePose(target, Tween(time.Second, Linear)).Child(Text("Pose")))
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
	core.Update(func() { target = Pose{X: 80, Y: 30, Scale: 1.4, Rotate: 45, Fade: .4}; width = 120 })
	h.Frame()
	now = now.Add(350 * time.Millisecond)
	before := renders
	h.Frame()
	if renders != before {
		t.Fatal("animation rebuilt the view")
	}
	painted := shot
	loop.Changed()
	h.Frame()
	if renders != before+1 {
		t.Fatal("forced build did not run")
	}
	for i, p := range painted.Pix {
		if p != shot.Pix[i] {
			t.Fatalf("paint-only pixel differs at byte %d: %d/%d", i, p, shot.Pix[i])
		}
	}
}

func TestPoseCurveChangeAtRestDoesNotAnimate(t *testing.T) {
	a := Smooth
	root := Root(ViewFunc(func(*Context) Element { return Div().Size(Dp(40)).AnimatePose(Pose{}, a) }))
	h := uitest.New(root)
	for _, next := range []Animation{Tween(time.Second, Linear), Bouncy, Snappy} {
		a = next
		h.Frame()
		if root.motionDue {
			t.Fatal("a curve change animated an unchanged pose")
		}
	}
}

func TestPosedLayerOutsidePressMatchesVisualBounds(t *testing.T) {
	show := true
	root := Root(ViewFunc(func(cx *Context) Element {
		if show {
			cx.Overlay("modal", Modal(Div().Size(Dp(100)).AnimatePose(Pose{X: 80}, Instant)).OnDismiss(func() { show = false }))
		}
		return Div()
	}))
	h := uitest.New(root)
	h.Click(260, 130)
	if !show {
		t.Fatal("press inside visual layer dismissed it")
	}
	h.Click(170, 130)
	if show {
		t.Fatal("press in old layout box did not dismiss moved layer")
	}
}

func TestPoseEditorCaretTransform(t *testing.T) {
	value := ""
	root := Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Input().ID("edit").Name("edit").W(Dp(100)).Bind(&value).AnimatePose(Pose{X: 80, Y: 50, Scale: 1.5}, Smooth))
	}))
	h := uitest.New(root)
	bounds := nodeBounds(h, "edit")
	h.Click(float32(bounds.Min.X+10), float32(bounds.Min.Y+10))
	h.Frame()
	h.Type("中文")
	if value != "中文" {
		t.Fatalf("posed editor did not receive text: %q", value)
	}
	s := h.Router.EditorState().Selection
	// The platform must receive the same scale as the drawing/hit-test path.
	p0, p1 := s.Transform.Transform(f32.Pt(0, 0)), s.Transform.Transform(f32.Pt(10, 0))
	if d := p1.X - p0.X; d < 14.99 || d > 15.01 {
		t.Fatalf("IME selection transform lost scale: %v", s.Transform)
	}
}

func TestOffscreenPoseSemantics(t *testing.T) {
	loop.SetAccessibility(true)
	defer loop.SetAccessibility(false)
	h := uitest.New(Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Div().Size(Dp(20)).AnimatePose(Pose{X: 500}, Smooth).Child(Text("offscreen pose").Translate(10, 0)))
	})))
	if bounds := nodeBounds(h, "offscreen pose"); bounds.Min.X != 510 {
		t.Fatalf("offscreen semantics lost parent pose or child translation: %v", bounds)
	}
}

func BenchmarkPoseFrames(b *testing.B) {
	for _, retain := range []bool{false, true} {
		b.Run(fmt.Sprintf("retain=%t", retain), func(b *testing.B) {
			now := time.Unix(1100, 0)
			x := float32(0)
			root := Root(ViewFunc(func(*Context) Element {
				rows := Div().Gap(3).AnimatePose(Pose{X: x}, Tween(time.Hour, Linear))
				for range 40 {
					rows.Child(Div().Row().Child(Text("Order"), Text("Pending")))
				}
				return rows
			}))
			if retain {
				root.RetainFrames()
			}
			h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
			core.Update(func() { x = 80 })
			h.Frame()
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				now = now.Add(time.Second / 120)
				h.Frame()
			}
			b.StopTimer()
			if retain && root.repaints != b.N {
				b.Fatalf("only %d paint-only frames out of %d", root.repaints, b.N)
			}
		})
	}
}
