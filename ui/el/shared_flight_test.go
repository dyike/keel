package el

import (
	"bytes"
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

func TestSharedFlightEscapesClipOpacityAndLaterSibling(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Skip(err)
	}
	defer gpu.Release()
	now := time.Unix(2200, 0)
	detail, clicks, renders := false, 0, 0
	ownOpacity := float32(1)
	exiting := false
	palette := theme.Current()
	palette.Primary = color.NRGBA{B: 255, A: 255}
	root := Root(ViewFunc(func(cx *Context) Element {
		renders++
		owner, x := "thumb", float32(20)
		if detail {
			owner, x = "detail", 260
		}
		parent := cx.Themed(palette, ViewFunc(func(*Context) Element {
			return Div().ID(owner).W(Dp(80)).H(Dp(80)).Translate(x, 20).Items(Start).Opacity(.1).Inert(exiting).OnClick(func() {}).Child(
				Div().ID("photo").Name("photo").Size(Dp(40)).Opacity(ownOpacity).SharedLayout("flight", Tween(time.Second, Linear)).Fly().OnClick(func() { clicks++ }).Decorate(func(_ core.C, draw func()) {
					if theme.Primary != palette.Primary {
						t.Fatal("flight lost ancestor theme")
					}
					draw()
				}).Bg(palette.Primary),
			)
		}))
		cx.Overlay("tip", Anchored("photo", Div().Name("tip").Size(Dp(10))).Placement(Bottom, Start))
		return Div().Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).Items(Start).Child(parent,
			Div().Absolute().Left(100).Top(0).W(Dp(140)).H(Dp(100)).Bg(color.NRGBA{G: 255, A: 255}),
		)
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
	core.Update(func() { detail = true })
	h.Frame()
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before {
		t.Fatal("flight rebuilt the view")
	}
	if b := nodeBounds(h, "photo"); b != image.Rect(140, 20, 180, 60) {
		t.Fatalf("flight bounds: %v", b)
	}
	if p := shot.RGBAAt(160, 40); p != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("clip/opacity/sibling hid flight: %v", p)
	}
	if b := nodeBounds(h, "tip"); b.Min != image.Pt(140, 64) {
		t.Fatalf("flight anchor: %v", b)
	}
	mid := shot
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("flight retained/full pixels differ")
	}
	core.Update(func() { ownOpacity = .5 })
	h.Frame()
	if p := shot.RGBAAt(160, 40); p.B < 100 || p.G == 0 {
		t.Fatalf("own opacity was not preserved: %v", p)
	}
	core.Update(func() { ownOpacity = 1 })
	h.Frame()
	h.Click(160, 40)
	if clicks != 1 {
		t.Fatal("flying element outside ancestor clip did not receive click")
	}
	if len(root.e.flights) != 0 {
		t.Fatal("flight queue survived painting")
	}
	for _, f := range root.e.flights[:cap(root.e.flights)] {
		if f.n != nil {
			t.Fatal("queue retained arena node")
		}
	}
	core.Update(func() { exiting = true })
	h.Frame()
	if p := shot.RGBAAt(160, 40); p.B > 150 || p.G < 200 {
		t.Fatalf("flight ignored ancestor exit fade: %v", p)
	}
	core.Update(func() { exiting = false })
	h.Frame()
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue {
		t.Fatal("settled flight scheduled frames")
	}
	if p := shot.RGBAAt(280, 40); p.B != 255 || p.R == 0 {
		t.Fatalf("settled element did not return to ancestor opacity: %v", p)
	}
}

func TestSharedFlightInsideModalDoesNotDismissOutsidePanel(t *testing.T) {
	now := time.Unix(2300, 0)
	open, clicks := false, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		photo := Div().ID("photo").Name("photo").Size(Dp(40)).SharedLayout("photo", Tween(time.Second, Linear)).Fly().OnClick(func() { clicks++ })
		if open {
			cx.Overlay("modal", Modal(Div().ID("detail").Size(Dp(80)).Items(Start).Child(photo)).OnDismiss(func() { open = false }))
			return Div()
		}
		return Div().P(20).Items(Start).Child(photo)
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { open = true })
	h.Frame()
	if b := nodeBounds(h, "photo"); b != image.Rect(20, 20, 60, 60) {
		t.Fatalf("modal handoff: %v", b)
	}
	h.Click(40, 40)
	if !open || clicks != 1 {
		t.Fatalf("flight click dismissed modal: open=%v clicks=%d", open, clicks)
	}
	h.Click(360, 250)
	if open {
		t.Fatal("outside press no longer dismisses modal")
	}
}

func TestFlightHitUsesRotatedRoundedShape(t *testing.T) {
	m := f32.AffineId().Rotate(f32.Pt(20, 20), .7).Offset(f32.Pt(100, 40))
	a := flightHit{m.Invert(), image.Pt(40, 40), corners{10, 10, 10, 10}}
	for _, p := range []f32.Point{{X: 20, Y: 20}, {X: 4, Y: 10}, {X: 35, Y: 20}} {
		if !a.contains(m.Transform(p)) {
			t.Fatalf("lost inside point: %v", p)
		}
	}
	for _, p := range []f32.Point{{}, {X: 1, Y: 1}, {X: 40, Y: 20}, {X: -1, Y: 20}} {
		if a.contains(m.Transform(p)) {
			t.Fatalf("accepted outside point: %v", p)
		}
	}
}

func TestSharedFlightPreservesOwnOpacityAndLaterOverlay(t *testing.T) {
	now := time.Unix(2400, 0)
	detail, top, clicks := false, false, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		owner, x := "a", float32(0)
		if detail {
			owner, x = "b", 200
		}
		if top {
			cx.Overlay("top", Modal(Div().Size(Dp(400))).OnDismiss(func() {}))
		}
		return Div().Items(Start).Child(Div().ID(owner).Pl(x).Items(Start).Child(Div().ID("photo").Name("photo").Size(Dp(40)).Opacity(.5).SharedLayout("flight", Tween(time.Second, Linear)).Fly().OnClick(func() { clicks++ })))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail = true })
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	h.Click(120, 20)
	if clicks != 1 {
		t.Fatal("own opacity blocked flight input")
	}
	core.Update(func() { top = true })
	h.Frame()
	h.Click(120, 20)
	if clicks != 1 {
		t.Fatal("flight intercepted later modal")
	}
}

func TestSharedFlightKeepsMovingWhileAncestorIsInert(t *testing.T) {
	now := time.Unix(2500, 0)
	detail, clicks := false, 0
	var shown f32.Affine2D
	var root *RootWidget
	root = Root(ViewFunc(func(*Context) Element {
		owner, x := "a", float32(0)
		if detail {
			owner, x = "b", 200
		}
		return Div().Items(Start).Child(Div().ID(owner).Pl(x).Items(Start).Inert(detail).Opacity(.5).Child(
			Div().ID("photo").Name("photo").Size(Dp(40)).SharedLayout("inert-flight", Tween(time.Second, Linear)).Fly().OnClick(func() { clicks++ }).Decorate(func(_ core.C, draw func()) { shown = root.e.poseWorld; draw() }),
		))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail = true })
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if p := shown.Transform(f32.Pt(200, 0)); p.X != 100 {
		t.Fatalf("inert ancestor snapped flight: %v", p)
	}
	h.Click(120, 20)
	if clicks != 0 {
		t.Fatal("inert flight exposed interaction")
	}
	if !root.motionDue {
		t.Fatal("inert flight stopped scheduling")
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue {
		t.Fatal("inert flight did not settle")
	}
}
