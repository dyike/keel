package window

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestInertExcludedFromAutomationAndAccessibility(t *testing.T) {
	inert := false
	w := openTest(t, Options{Width: 300, Height: 200, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Child(el.Text("Outside"), el.Div().Inert(inert).Child(el.Div().Name("Action").OnClick(func() {}).Child(el.Text("Inside")), el.Input().Name("Editor")))
	}))})
	if len(w.snapshot()) < 3 {
		t.Fatal("missing live controls")
	}
	inert = true
	loop.SetAccessibility(true)
	defer loop.SetAccessibility(false)
	w.render()
	for _, ax := range []bool{false, true} {
		els := elementsOf(w.virt.router.AppendSemantics(nil), w.window(), ax)
		if len(els) != 1 || els[0].Name != "Outside" {
			t.Fatalf("ax=%v inert controls exposed: %+v", ax, els)
		}
	}
}

func TestLayerPresenceExitSemantics(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	now := time.Unix(2000, 0)
	show := true
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		p := cx.Presence("panel", show, el.PresenceOptions{Enter: el.Pose{Fade: 1}, Exit: el.Pose{Fade: 1}, Animation: el.Tween(time.Second, el.Linear)})
		if p.Visible {
			cx.Overlay("panel", el.Anchored("anchor", el.Div().Name("Panel").Size(el.Dp(60)).Child(el.Text("Inside"))).Presence(p))
		}
		return el.Div().Child(el.Div().ID("anchor").Size(el.Dp(20)), el.Text("Outside"))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(10 * time.Millisecond)
	h.Frame()
	show = false
	h.Frame()
	for _, ax := range []bool{false, true} {
		els := elementsOf(h.Router.AppendSemantics(nil), image.Rect(0, 0, 400, 300), ax)
		for _, e := range els {
			if e.Name == "Panel" || e.Name == "Inside" {
				t.Fatalf("ax=%v exiting layer exposed: %+v; raw=%+v", ax, els, h.Router.AppendSemantics(nil))
			}
		}
	}
}

func TestSharedFlightExitSemantics(t *testing.T) {
	now := time.Unix(2600, 0)
	detail := false
	root := el.Root(el.ViewFunc(func(*el.Context) el.Element {
		owner, x := "a", float32(0)
		if detail {
			owner, x = "b", 200
		}
		return el.Div().Items(el.Start).Child(el.Text("Outside"), el.Div().ID(owner).Pl(x).Inert(detail).Items(el.Start).Child(
			el.Div().Name("Flying action").Size(el.Dp(40)).SharedLayout("exit-flight", el.Tween(time.Second, el.Linear)).Fly().OnClick(func() {}).Child(el.Text("Inside")),
		))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	detail = true
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	for _, ax := range []bool{false, true} {
		els := elementsOf(h.Router.AppendSemantics(nil), image.Rect(0, 0, 400, 300), ax)
		if len(els) != 1 || els[0].Name != "Outside" {
			t.Fatalf("ax=%v inert flight exposed: %+v", ax, els)
		}
	}
}
