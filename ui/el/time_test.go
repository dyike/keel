package el

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"testing"
	"time"
)

func TestAfterLifecycleAndInjectedTime(t *testing.T) {
	now := time.Unix(1000, 0)
	show := true
	calls := 0
	root := Root(viewFunc(func(cx *Context) Element {
		if cx.Now() != now {
			t.Fatal("wrong frame time")
		}
		box := Div()
		if show {
			cx.After("notice", time.Second, func() { calls++ })
			box.Child(Div().ID("notice").Child(Text("notice")))
		}
		return box
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if calls != 0 {
		t.Fatal("early timer")
	}
	show = false
	now = now.Add(time.Second)
	h.Frame()
	if calls != 0 {
		t.Fatal("omitted timer fired")
	}
	show = true
	h.Frame()
	now = now.Add(time.Second)
	h.Frame()
	h.Frame()
	if calls != 1 {
		t.Fatal("one-shot timer", calls)
	}
}
func TestReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(true)
	if !ReducedMotion() {
		t.Fatal("override ignored")
	}
	theme.SetReducedMotion(false)
	if ReducedMotion() {
		t.Fatal("restore ignored")
	}
}

func TestAfterCancellationDuringClickRerender(t *testing.T) {
	now := time.Unix(1000, 0)
	show := true
	calls := 0
	root := Root(viewFunc(func(cx *Context) Element {
		if show {
			cx.After("cancel", time.Second, func() { calls++ })
		}
		return Div().Name("cancel").OnClick(func() { show = false }).Child(Text("cancel"))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(time.Second)
	h.Click(10, 10)
	h.Frame()
	if calls != 0 {
		t.Fatal("omitted timer fired from discarded render")
	}
}

func TestAfterKeysSurviveConditionalInsertion(t *testing.T) {
	now := time.Unix(1000, 0)
	middle := false
	calls := map[string]int{}
	declare := func(cx *Context, name string) {
		cx.After(struct{ ID string }{name}, time.Second, func() { calls[name]++ })
	}
	root := Root(viewFunc(func(cx *Context) Element {
		declare(cx, "first")
		if middle {
			declare(cx, "middle")
		}
		declare(cx, "last")
		return Div()
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(500 * time.Millisecond)
	middle = true
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if calls["first"] != 1 || calls["last"] != 1 || calls["middle"] != 0 {
		t.Fatalf("timers shifted in wrapper: %v", calls)
	}
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	h.Frame()
	if calls["first"] != 1 || calls["last"] != 1 || calls["middle"] != 1 {
		t.Fatalf("timers not independent one-shots: %v", calls)
	}
}

func TestAfterDurationChangeRestartsCountdown(t *testing.T) {
	now := time.Unix(1000, 0)
	duration := time.Second
	calls := 0
	root := Root(viewFunc(func(cx *Context) Element { cx.After("delay", duration, func() { calls++ }); return Div() }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(500 * time.Millisecond)
	duration = 2 * time.Second
	h.Frame()
	now = now.Add(1500 * time.Millisecond)
	h.Frame()
	if calls != 0 {
		t.Fatal("old deadline fired")
	}
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	h.Frame()
	if calls != 1 {
		t.Fatal("changed duration did not fire once")
	}
}

func TestAfterInCacheBuilderIsCancelledOnCacheHit(t *testing.T) {
	now := time.Unix(1000, 0)
	calls, builds := 0, 0
	root := Root(viewFunc(func(cx *Context) Element {
		return cx.Cache("content", func() Element {
			builds++
			cx.After("cached-timer", time.Second, func() { calls++ })
			return Div().Child(Text("cached"))
		})
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	now = now.Add(time.Second)
	h.Frame()
	if builds != 1 || calls != 0 {
		t.Fatalf("cache timer should be omitted: builds=%d calls=%d", builds, calls)
	}
}

func TestAfterInvalidatesOtherWindows(t *testing.T) {
	invalidations := 0
	tag := new(int)
	loop.Lock()
	loop.Register(tag, func() { invalidations++ })
	loop.Unlock()
	defer func() { loop.Lock(); loop.Unregister(tag); loop.Unlock() }()
	root := Root(viewFunc(func(cx *Context) Element { cx.After("notify", 0, func() {}); return Div() }))
	uitest.New(root)
	if invalidations != 1 {
		t.Fatalf("timer callback bypassed core.Call: invalidations=%d", invalidations)
	}
}

func TestAfterEnabledWaitsForVisibleOwner(t *testing.T) {
	now := time.Unix(1000, 0)
	calls := 0
	visible := false
	disabled := false
	root := Root(viewFunc(func(cx *Context) Element {
		cx.AfterEnabled("owner", "owned-timer", time.Second, func() { calls++ })
		box := Div().Disabled(disabled)
		if visible {
			box.Child(Div().ID("owner").H(Dp(20)))
		}
		return box
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	now = now.Add(5 * time.Second)
	h.Frame()
	if calls != 0 {
		t.Fatal("missing owner fired")
	}
	visible = true
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	disabled = true
	h.Frame()
	now = now.Add(5 * time.Second)
	h.Frame()
	if calls != 0 {
		t.Fatal("disabled owner fired")
	}
	disabled = false
	h.Frame()
	now = now.Add(999 * time.Millisecond)
	h.Frame()
	if calls != 0 {
		t.Fatal("did not restart full delay")
	}
	now = now.Add(time.Millisecond)
	h.Frame()
	h.Frame()
	if calls != 1 {
		t.Fatalf("one-shot calls %d", calls)
	}
	disabled = true
	h.Frame()
	disabled = false
	h.Frame()
	now = now.Add(2 * time.Second)
	h.Frame()
	if calls != 1 {
		t.Fatal("completed one-shot timer rearmed after disabling")
	}
}

func TestCountdownPreservesRemainingDelayAcrossPauses(t *testing.T) {
	for _, mode := range []string{"explicit", "disabled", "modal"} {
		now := time.Unix(100, 0)
		paused, disabled, modal := false, false, false
		calls := 0
		root := Root(viewFunc(func(cx *Context) Element {
			cx.Countdown("owner", "countdown", 5*time.Second, paused, func() { calls++ })
			if modal {
				cx.Overlay("modal", Modal(Div().Size(Dp(50))))
			}
			return Div().ID("owner").Disabled(disabled).Size(Dp(100))
		}))
		h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
		now = now.Add(2 * time.Second)
		h.Frame()
		switch mode {
		case "explicit":
			paused = true
		case "disabled":
			disabled = true
		case "modal":
			modal = true
		}
		h.Frame()
		now = now.Add(20 * time.Second)
		h.Frame()
		if calls != 0 {
			t.Fatalf("%s fired paused", mode)
		}
		paused, disabled, modal = false, false, false
		h.Frame()
		now = now.Add(2900 * time.Millisecond)
		h.Frame()
		if calls != 0 {
			t.Fatalf("%s fired early", mode)
		}
		now = now.Add(100 * time.Millisecond)
		h.Frame()
		h.Frame()
		if calls != 1 {
			t.Fatalf("%s did not resume remaining time: %d", mode, calls)
		}
	}
}
