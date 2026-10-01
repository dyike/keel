package el

import (
	"github.com/dyike/keel/ui/core"
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
			s := cx.Scope("notice")
			s.After(time.Second, func() { calls++ })
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
		t.Fatal("removed owner fired")
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
func TestAfterAbsentScopeOwner(t *testing.T) {
	n := 0
	r := Root(viewFunc(func(cx *Context) Element { cx.Scope("absent").After(0, func() { n++ }); return Div() }))
	uitest.New(r)
	if n != 0 {
		t.Fatal("timer for absent element fired")
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
			cx.After(time.Second, func() { calls++ })
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
