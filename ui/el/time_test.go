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
