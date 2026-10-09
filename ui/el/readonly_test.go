package el

import (
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDisabledParentRendersRetainedState(t *testing.T) {
	disabled := false
	builds := 0
	root := Embed(ViewFunc(func(cx *Context) Element {
		return Div().Child(cx.Cache("content", func() Element {
			builds++
			return Div().Child(Input().ID("name").Placeholder("名字"),
				Div().ID("scroll").W(Dp(120)).H(Dp(50)).ScrollY().Child(Text("顶部"), Div().H(Dp(200)), Text("底部")))
		}))
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		if disabled {
			gtx = gtx.Disabled()
		}
		root.Layout(gtx)
	})
	h.Click(center(nodeBounds(h, "名字")))
	h.Type("abc")
	if got := description(h, "名字"); got != "abc" {
		t.Fatalf("initial value: %q", got)
	}
	var scroll *elemState
	for _, s := range root.store.states {
		if s.id == "scroll" {
			scroll = s
		}
	}
	if scroll == nil {
		t.Fatal("missing scroll state")
	}
	scroll.scrollY = 60
	h.Frame()
	beforeScroll, beforeFrame := scroll.scrollY, root.store.frame
	disabled = true
	for i := 0; i < 3; i++ {
		h.Frame()
		if got := description(h, "名字"); got != "abc" {
			t.Fatalf("disabled frame %d value: %q", i, got)
		}
	}
	if builds != 1 || scroll.scrollY != beforeScroll || root.store.frame != beforeFrame {
		t.Fatalf("readonly changed state: builds=%d scroll=%d frame=%d", builds, scroll.scrollY, root.store.frame)
	}
	disabled = false
	h.Frame()
	if got := description(h, "名字"); got != "abc" || builds != 1 {
		t.Fatalf("restored value=%q builds=%d", got, builds)
	}
}

func TestReadOnlyFramesPreserveOverlayFocusCacheAndTimers(t *testing.T) {
	readonly := false
	show := true
	request := true
	calls, builds := 0, 0
	now := time.Unix(1000, 0)
	root := Root(ViewFunc(func(cx *Context) Element {
		if show {
			cx.Overlay("modal", Modal(Div().Child(Input().ID("editor").Name("editor"))).OnDismiss(func() { calls++ }))
			cx.After("timer", time.Second, func() { calls++ })
			if request {
				cx.Focus("editor")
				request = false
			}
		}
		return Div().When(show, func(d *DivEl) { d.Child(cx.Cache("retained", func() Element { builds++; return Text("cached") })) })
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		if readonly {
			// Neither omissions nor replacement declarations in discarded/disabled
			// layouts may change a live layer, focus history, timer, or cache entry.
			show = false
			for i := 0; i < 3; i++ {
				var scratch op.Ops
				g := gtx
				g.Ops = &scratch
				g.Source = input.Source{}
				root.Layout(g)
			}
			show = true
		}
		root.Layout(gtx)
	})
	h.Frame()
	layer := root.layers["modal"]
	timer := root.timers["timer"]
	saved := layer.returnFocus
	readonly = true
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if root.layers["modal"] != layer || !layer.active || layer.returnFocus != saved || root.timers["timer"] != timer || calls != 0 || builds != 1 {
		t.Fatal("readonly frame changed live lifecycle")
	}
	if !(&Context{root: root}).Focused("editor") {
		t.Fatal("readonly frames lost focus")
	}
	h.Type("abc")
	if got := description(h, "editor"); got != "abc" {
		t.Fatalf("focus no longer edits: %q", got)
	}
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if calls != 1 {
		t.Fatalf("timer deadline changed: %d", calls)
	}
	h.Key(key.NameEscape, 0)
	if calls != 2 {
		t.Fatalf("overlay callback broken: %d", calls)
	}
}

func TestReadOnlyTimerDeclarationDoesNotResetDeadline(t *testing.T) {
	readonly := false
	duration := time.Second
	calls := 0
	now := time.Unix(1000, 0)
	root := Embed(ViewFunc(func(cx *Context) Element { cx.After("timer", duration, func() { calls++ }); return Div() }))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		if readonly {
			gtx = gtx.Disabled()
		}
		root.Layout(gtx)
	})
	timer := root.timers["timer"]
	readonly = true
	duration = time.Hour
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.timers["timer"] != timer || calls != 0 {
		t.Fatal("readonly declaration reset or fired timer")
	}
	readonly = false
	duration = time.Second
	h.Frame()
	if calls != 1 {
		t.Fatal("original deadline not retained")
	}
}
