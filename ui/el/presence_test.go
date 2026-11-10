package el

import (
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestPresenceLifecycleReentryAndRelease(t *testing.T) {
	now := time.Unix(1500, 0)
	show, declare := true, true
	var frame PresenceFrame
	root := Root(ViewFunc(func(cx *Context) Element {
		body := Div().Items(Start)
		if declare {
			frame = cx.Presence("card", show, PresenceOptions{Enter: Pose{X: 100, Fade: 1}, Exit: Pose{X: -100, Fade: 1}, Animation: Tween(time.Second, Linear)})
			if frame.Visible {
				body.Child(Div().Name("card").Size(Dp(40)).Presence(frame))
			}
		}
		return body
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	if !frame.Visible || frame.Settled || frame.Pose.X != 100 {
		t.Fatalf("entry: %+v", frame)
	}
	now = now.Add(time.Second)
	h.Frame()
	if !frame.Settled || !frame.Arrived || frame.Pose.X != 0 {
		t.Fatalf("arrived: %+v", frame)
	}
	show = false
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if !frame.Visible || !frame.Exiting || frame.Pose.X != -50 {
		t.Fatalf("exit: %+v", frame)
	}
	show = true
	h.Frame()
	if frame.Exiting || frame.Pose.X != -50 {
		t.Fatalf("reentry jumped: %+v", frame)
	}
	now = now.Add(time.Second)
	h.Frame()
	if frame.Pose.X != 0 || !frame.Settled {
		t.Fatalf("reentry settle: %+v", frame)
	}
	show = false
	h.Frame()
	now = now.Add(time.Second)
	h.Frame()
	if frame.Visible || !frame.Arrived || len(root.presences) != 0 {
		t.Fatalf("exit retained state: %+v count=%d", frame, len(root.presences))
	}
	show = true
	h.Frame()
	declare = false
	h.Frame()
	if len(root.presences) != 0 {
		t.Fatal("omitted presence retained state")
	}
}

func TestInertStopsInputFocusAndOwnedLayers(t *testing.T) {
	inert, clicks, keys := false, 0, 0
	value := ""
	root := Root(ViewFunc(func(cx *Context) Element {
		cx.Overlay("tip", Anchored("button", Div().Name("tip").Size(Dp(10))))
		return Div().Items(Start).Inert(inert).Child(
			Div().ID("button").Name("button").Size(Dp(40)).OnClick(func() { clicks++ }).OnKey(func(KeyEvent) bool { keys++; return true }),
			Input().ID("edit").Name("edit").W(Dp(120)).Bind(&value),
		)
	}))
	h := uitest.New(root)
	h.Click(20, 20)
	if clicks != 1 {
		t.Fatal("live button did not click")
	}
	b := nodeBounds(h, "edit")
	h.Click(float32(b.Min.X+10), float32(b.Min.Y+10))
	h.Frame()
	h.Type("A")
	if value != "A" {
		t.Fatal("live input did not type")
	}
	inert = true
	h.Frame()
	h.Click(20, 20)
	h.Key(key.NameReturn, 0)
	h.Type("B")
	if clicks != 1 || keys != 0 || value != "A" {
		t.Fatalf("inert input: clicks=%d keys=%d value=%q", clicks, keys, value)
	}
	if len(root.activeLayers) != 0 {
		t.Fatal("inert owner retained overlay")
	}
	for _, st := range root.store.states {
		if st.editor != nil && h.Router.Source().Focused(st.editor) {
			t.Fatal("inert editor retained focus")
		}
	}
	inert = false
	h.Frame()
	h.Click(20, 20)
	if clicks != 2 {
		t.Fatal("first click after reentry lost")
	}
}

func TestPresenceReducedMotionAndMeasurement(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	now := time.Unix(1600, 0)
	show, disabled := true, false
	var frame PresenceFrame
	root := Root(ViewFunc(func(cx *Context) Element {
		frame = cx.Presence("card", show, PresenceOptions{Enter: Pose{Fade: 1}, Exit: Pose{Fade: 1}, Animation: Tween(time.Second, Linear)})
		return Div()
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		if disabled {
			gtx = gtx.Disabled()
		}
		root.Layout(gtx)
	})
	state := root.presences["card"]
	now = now.Add(500 * time.Millisecond)
	disabled = true
	h.Frame()
	if state.values[6].Get() != 1 {
		t.Fatal("measurement advanced live presence")
	}
	disabled = false
	h.Frame()
	if frame.Pose.Fade != .5 {
		t.Fatalf("live midpoint: %+v", frame)
	}
	theme.SetReducedMotion(true)
	h.Frame()
	if frame.Pose.Fade != 0 || !frame.Settled {
		t.Fatalf("reduced entry: %+v", frame)
	}
	show = false
	h.Frame()
	if frame.Visible || len(root.presences) != 0 {
		t.Fatal("reduced exit retained state")
	}
}

func TestModalPresenceReentryClick(t *testing.T) {
	now := time.Unix(1700, 0)
	show, clicks := false, 0
	root := Root(ViewFunc(func(cx *Context) Element {
		p := cx.Presence("modal", show, PresenceOptions{Enter: Pose{Fade: 1}, Exit: Pose{Fade: 1}, Animation: Tween(time.Second, Linear)})
		if p.Visible {
			cx.Overlay("modal", Modal(Div().ID("panel").Size(Dp(100)).Child(Div().ID("ok").Name("OK").Size(Dp(40)).OnClick(func() { clicks++ }))).OnDismiss(func() { show = false }).Presence(p))
		}
		return Div()
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	for i := range 3 {
		show = true
		h.Frame()
		b := nodeBounds(h, "OK")
		h.Click(float32(b.Min.X+10), float32(b.Min.Y+10))
		if clicks != i+1 {
			t.Fatalf("reentry %d lost click at %v", i, b)
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		if show {
			t.Fatal("Esc did not close")
		}
	}
}
