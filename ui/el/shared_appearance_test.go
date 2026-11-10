package el

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSharedAppearanceHandoffReverseAndRest(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	now := time.Unix(2200, 0)
	detail, enabled := false, true
	renders, clicks := 0, 0
	root := Root(ViewFunc(func(*Context) Element {
		renders++
		owner, bg, radius, alpha := "a", color.NRGBA{R: 240, A: 255}, float32(4), float32(1)
		if detail {
			owner, bg, radius, alpha = "b", color.NRGBA{B: 240, A: 255}, 24, .5
		}
		card := Div().Name("look").Size(Dp(80)).Rounded(radius).Bg(bg).Border(2, bg).Opacity(alpha).OnClick(func() { clicks++ })
		if enabled {
			card.SharedLayout("look", Tween(time.Second, Linear)).Fly()
		}
		return Div().Size(Full).Bg(theme.Bg).Items(Start).Child(Div().ID(owner).P(20).Items(Start).Child(card))
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
	if !root.motionDue {
		t.Fatal("look-only handoff did not animate")
	}
	assertSharedPixelContinuity(t, initial, shot)
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before {
		t.Fatal("appearance rebuilt view")
	}
	s := root.e.sharedLayouts["look"]
	if s.look.background != (color.NRGBA{R: 120, B: 120, A: 255}) || s.look.opacity != .75 {
		t.Fatalf("incorrect halfway appearance: %+v", s.look)
	}
	mid := shot
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("retained/full appearance differs")
	}
	h.Click(60, 60)
	if clicks != 1 {
		t.Fatal("appearance-only flight lost input")
	}
	core.Update(func() { detail = false })
	h.Frame()
	assertSharedPixelContinuity(t, mid, shot)
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue || s.lookActive {
		t.Fatal("appearance did not settle")
	}
	settled := shot
	core.Update(func() { enabled = false })
	h.Frame()
	if !bytes.Equal(settled.Pix, shot.Pix) {
		t.Fatal("settled appearance differs from ordinary paint")
	}
	if root.e.sharedLayouts != nil {
		t.Fatal("appearance registry retained after omission")
	}
}

func TestSharedGradientShadowContinuityAndRemoval(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	now := time.Unix(2600, 0)
	stage := 0
	enabled := true
	root := Root(ViewFunc(func(*Context) Element {
		g := theme.Gradient{From: color.NRGBA{R: 220, A: 255}, To: color.NRGBA{B: 210, A: 255}, Angle: 45}
		elevation := theme.ElevationSm
		if stage == 1 {
			g.Angle = 225
			elevation = theme.ElevationLg
		}
		card := Div().Size(Dp(80)).Rounded(12).BgGradient(g)
		if stage != 2 {
			card.Shadow(elevation)
		} else {
			card.Bg(theme.Surface)
		}
		if enabled {
			card.SharedLayout("look", Tween(time.Second, Linear))
		}
		return Div().Size(Full).Bg(theme.Bg).Items(Start).Child(Div().ID(string(rune('a' + stage))).P(60).Items(Start).Child(card))
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
	core.Update(func() { stage = 1 })
	h.Frame()
	assertSharedPixelContinuity(t, initial, shot)
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	mid := shot
	s := root.e.sharedLayouts["look"]
	if s.look.gradient.start == s.look.gradient.end {
		t.Fatal("opposite gradient angles collapsed to a point")
	}
	if s.look.shadowBlur <= 3.0/80 || s.look.shadowBlur >= 32.0/80 {
		t.Fatal("shadow blur did not interpolate")
	}
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("gradient retained/full differs")
	}
	core.Update(func() { stage = 2 })
	h.Frame()
	assertSharedPixelContinuity(t, mid, shot)
	now = now.Add(2 * time.Second)
	h.Frame()
	settled := shot
	if root.motionDue {
		t.Fatal("gradient/shadow removal did not settle")
	}
	core.Update(func() { enabled = false })
	h.Frame()
	if !bytes.Equal(settled.Pix, shot.Pix) {
		t.Fatal("settled gradient/shadow differs")
	}
}

func TestSharedScaledGradientShadowHandoff(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	now := time.Unix(2800, 0)
	detail := false
	root := Root(ViewFunc(func(*Context) Element {
		owner, size := "a", float32(80)
		if detail {
			owner, size = "b", 160
		}
		return Div().Size(Full).Bg(theme.Bg).Items(Start).Child(Div().ID(owner).P(60).Items(Start).Child(
			Div().Size(Dp(size)).Rounded(8).Shadow(theme.ElevationMd).
				BgGradient(theme.Gradient{From: theme.Primary, To: theme.Success, Angle: 30}).SharedLayout("look", Tween(time.Second, Linear)).Fly(),
		))
	}))
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
}
