package el

import (
	"bytes"
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSharedContentUnequalAspectAndFontHandoff(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	now := time.Unix(2300, 0)
	detail, enabled := false, true
	renders := 0
	root := Root(ViewFunc(func(*Context) Element {
		renders++
		owner, value, width, height, size, x := "a", "Café 中文 12", float32(180), float32(45), float32(18), float32(10)
		if detail {
			owner, value, width, height, size, x = "b", "中文 Café 21", 300, 90, 24, 60
		}
		label := Text(value).ID("title").W(Dp(width)).H(Dp(height)).TextSize(size)
		if enabled {
			label.ContentTransition(ContentTransition{Morph: true, Animation: Tween(time.Second, Linear)}).SharedLayout("title", Tween(time.Second, Linear)).Fly()
		}
		return Div().Size(Full).Bg(theme.Bg).Items(Start).Child(Div().ID(owner).Translate(x, 30).Items(Start).Child(label))
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
	s := root.e.sharedLayouts["title"]
	if s.content == nil || len(s.content.tracks) == 0 || s.contentFrom != nil {
		t.Fatal("content handoff missing or pending glyphs retained")
	}
	before := renders
	now = now.Add(400 * time.Millisecond)
	h.Frame()
	if renders != before {
		t.Fatal("shared text rebuilt view")
	}
	mid := shot
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("retained/full shared text differs")
	}
	core.Update(func() { detail = false })
	h.Frame()
	assertSharedPixelContinuity(t, mid, shot)
	for _, n := range h.Router.AppendSemantics(nil) {
		if n.Desc.Label == "中文 Café 21" {
			t.Fatal("old shared text leaked into semantics")
		}
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue || len(s.content.tracks) != 0 {
		t.Fatal("shared text did not settle")
	}
	settled := shot
	core.Update(func() { enabled = false })
	h.Frame()
	if !bytes.Equal(settled.Pix, shot.Pix) {
		t.Fatal("settled shared text differs from ordinary paint")
	}
	if root.e.sharedLayouts != nil {
		t.Fatal("shared content retained after omission")
	}
}

func TestSharedContentInertReducedMotionAndRemoval(t *testing.T) {
	oldReduced := theme.ReducedMotion
	defer theme.SetReducedMotion(oldReduced)
	now := time.Unix(2400, 0)
	detail, inert, show := false, false, true
	root := Root(ViewFunc(func(*Context) Element {
		if !show {
			return Div()
		}
		owner, value := "a", "First title"
		if detail {
			owner, value = "b", "Second title"
		}
		return Div().Items(Start).Child(Div().ID(owner).Inert(inert).Items(Start).Child(
			Text(value).W(Dp(200)).H(Dp(50)).ContentTransition(ContentTransition{Morph: true, Animation: Tween(time.Second, Linear)}).SharedLayout("text", Tween(time.Second, Linear)),
		))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { detail, inert = true, true })
	h.Frame()
	if s := root.e.sharedLayouts["text"]; s.content == nil || len(s.content.tracks) == 0 || !root.motionDue {
		t.Fatal("inert handoff stopped text animation")
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue {
		t.Fatal("inert shared text did not settle")
	}
	theme.SetReducedMotion(true)
	core.Update(func() { detail = false })
	h.Frame()
	if s := root.e.sharedLayouts["text"]; s.content != nil && len(s.content.tracks) != 0 || root.motionDue {
		t.Fatal("reduced motion retained shared glyph animation")
	}
	core.Update(func() { show = false })
	h.Frame()
	if root.e.sharedLayouts != nil || len(root.e.contentStates) != 0 {
		t.Fatal("removed shared text retained content")
	}
}
