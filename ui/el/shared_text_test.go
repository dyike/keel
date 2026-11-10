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

func TestMatchTextNestedHandoffAndReversal(t *testing.T) {
	gpu, err := headless.NewWindow(600, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	now := time.Unix(2500, 0)
	detail, match := false, true
	root := Root(ViewFunc(func(*Context) Element {
		owner, value, w, h, size, x := "a", "中文 Keel", float32(300), float32(80), float32(20), float32(20)
		if detail {
			owner, value, w, h, size, x = "b", "Keel 与中文", 400, 130, 26, 150
		}
		panel := Div().ID("panel").W(Dp(w)).H(Dp(h)).Items(Start).Bg(theme.Surface).SharedLayout("panel", Tween(time.Second, Linear)).Fly()
		if match {
			panel.MatchText()
		}
		panel.Child(Text(value).ID("title").W(Dp(w)).H(Dp(50)).TextSize(size))
		return Div().Size(Full).Bg(theme.Bg).Items(Start).Child(Div().ID(owner).Translate(x, 30).Items(Start).Child(panel))
	}))
	root.RetainFrames()
	var shot *image.RGBA
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Constraints.Max = image.Pt(600, 300)
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		shot = image.NewRGBA(image.Rect(0, 0, 600, 300))
		if err := gpu.Screenshot(shot); err != nil {
			t.Fatal(err)
		}
	})
	initial := shot
	core.Update(func() { detail = true })
	h.Frame()
	assertSharedPixelContinuity(t, initial, shot)
	found := false
	for id, s := range root.e.sharedLayouts {
		if _, ok := id.(sharedTextIdentity); ok {
			found = true
			if s.content == nil || len(s.content.tracks) == 0 {
				t.Fatal("automatic text did not morph")
			}
		}
	}
	if !found {
		t.Fatal("descendant not matched")
	}
	now = now.Add(400 * time.Millisecond)
	h.Frame()
	mid := shot
	loop.Changed()
	h.Frame()
	if !bytes.Equal(mid.Pix, shot.Pix) {
		t.Fatal("retained/full automatic match differs")
	}
	core.Update(func() { detail = false })
	h.Frame()
	assertSharedPixelContinuity(t, mid, shot)
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue {
		t.Fatal("automatic matching did not settle")
	}
	core.Update(func() { match = false })
	h.Frame()
	if len(root.e.sharedLayouts) != 1 || len(root.e.contentStates) != 0 {
		t.Fatal("omitted MatchText retained automatic child state")
	}
}

func TestMatchTextCachedChildrenAndExplicitScope(t *testing.T) {
	match := true
	root := Root(ViewFunc(func(cx *Context) Element {
		child := cx.Cache("label", func() Element { return Text("cached").ID("title") })
		panel := Div().SharedLayout("panel", Snappy)
		if match {
			panel.MatchText()
		}
		return panel.Child(child, Text("explicit").SharedLayout("explicit", Snappy))
	}))
	h := uitest.New(root)
	if len(root.e.sharedLayouts) != 3 {
		t.Fatalf("expected panel/implicit/explicit, got %d", len(root.e.sharedLayouts))
	}
	core.Update(func() { match = false })
	h.Frame()
	if len(root.e.sharedLayouts) != 2 || len(root.e.contentStates) != 0 {
		t.Fatal("cached automatic declarations survived opt-out")
	}
}
