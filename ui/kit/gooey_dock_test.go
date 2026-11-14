package kit

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"math"
	"testing"
	"time"
)

func TestGooeyDockInteractionAndRest(t *testing.T) {
	now := time.Unix(100, 0)
	calls := 0
	v := GooeyDock(GooeyDockItem{ID: "home", Label: "Home", Icon: IconHome, Active: true, Action: func() { calls++ }}, GooeyDockItem{ID: "search", Label: "Search", Icon: IconSearch, Action: func() { calls += 10 }}, GooeyDockItem{ID: "off", Label: "Off", Disabled: true})
	root := el.Root(v)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	b := bounds(h, "Home")

	h.Move(float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2))
	for range 180 {
		now = now.Add(time.Second / 60)
		h.Frame()
	}
	if math.Abs(float64(v.states[0].size.Get()-58)) > .1 {
		t.Fatalf("peak size %v, target %v", v.states[0].size.Get(), v.states[0].target)
	}
	if v.states[1].size.Get() <= 40 {
		t.Fatal("neighbour did not magnify")
	}
	click(t, h, "Home")
	if calls != 1 {
		t.Fatalf("clicks %d", calls)
	}
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameReturn, 0)
	if calls != 11 {
		t.Fatalf("keyboard action: %d", calls)
	}
	h.Move(399, 299)
	// Clear keyboard focus before measuring idle pointer state.
	root2 := el.Root(v)
	h = uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root2.Layout(gtx) })
	for range 180 {
		now = now.Add(time.Second / 60)
		h.Frame()
	}
	if math.Abs(float64(v.states[0].size.Get()-40)) > .01 {
		t.Fatal("did not return to base")
	}
	h.Router.WakeupTime()
	h.Frame()
	if _, ok := h.Router.PeekWakeup(); ok {
		t.Fatal("idle dock requests frames")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "Home")
	if calls != 11 {
		t.Fatal("disabled action fired")
	}
}

func TestGooeyDockReferenceFalloff(t *testing.T) {
	for _, tc := range []struct{ distance, want float32 }{{0, 1}, {75, .5}, {-75, .5}, {150, 0}, {200, 0}} {
		if math.Abs(float64(gooeyFalloff(tc.distance)-tc.want)) > 1e-6 {
			t.Fatal(tc)
		}
	}
}

func TestGooeyDockCSSCompositing(t *testing.T) {
	black, white := color.NRGBA{A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if c := gooeyComposite(black, white, .45); c.R != 140 || c.A != 255 {
		t.Fatalf("CSS ink %v", c)
	}
	if c := gooeyComposite(white, black, .45); c.R != 115 || c.A != 255 {
		t.Fatalf("dark CSS ink %v", c)
	}
}

func TestGooeyDockReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	v := GooeyDock(GooeyDockItem{ID: "home", Label: "Home", Icon: IconHome})
	h := uitest.New(el.Root(v))
	h.Move(33, 73)
	h.Frame()
	if v.states[0].size.Get() != 58 {
		t.Fatal("reduced motion did not jump to peak")
	}
	h.Move(399, 299)
	h.Frame()
	if v.states[0].size.Get() != 40 {
		t.Fatal("reduced motion did not reset")
	}
	h.Router.WakeupTime()
	h.Frame()
	if _, ok := h.Router.PeekWakeup(); ok {
		t.Fatal("reduced motion requested animation")
	}
}
