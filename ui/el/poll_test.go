package el

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
)

type pollWindow struct {
	core.WindowControls
	closed atomic.Bool
}

func (w *pollWindow) Closed() bool { return w.closed.Load() }

// Poll checks between frames, and draws only when a check changed something.
func TestPollDrawsOnlyOnChange(t *testing.T) {
	win := &pollWindow{}
	var checks, draws atomic.Int32
	var change atomic.Bool
	declare := true
	root := Root(viewFunc(func(cx *Context) Element {
		if declare {
			cx.Poll("check", 5*time.Millisecond, func() bool {
				checks.Add(1)
				return change.Load()
			})
		}
		return Div()
	}))
	key := new(int)
	loop.Register(key, func() { draws.Add(1) })
	defer loop.Unregister(key)
	h := uitest.NewFunc(func(gtx core.C) {
		defer core.SetCurrentWindow(win)()
		root.Layout(gtx)
	})
	h.Frame()

	waitFor := func(what string, cond func() bool) {
		for deadline := time.Now().Add(2 * time.Second); !cond(); {
			if time.Now().After(deadline) {
				t.Fatal(what)
			}
			time.Sleep(time.Millisecond)
		}
	}
	waitFor("no checks", func() bool { return checks.Load() >= 3 })
	if n := draws.Load(); n != 0 {
		t.Fatalf("checks that changed nothing drew %d frames", n)
	}
	change.Store(true)
	waitFor("a change drew no frame", func() bool { return draws.Load() > 0 })
	change.Store(false)

	// A frame that omits the poll stops it.
	declare = false
	h.Frame()
	time.Sleep(20 * time.Millisecond)
	n := checks.Load()
	time.Sleep(30 * time.Millisecond)
	if checks.Load() != n {
		t.Fatal("omitted poll kept checking")
	}

	// Closing the window stops it too.
	declare = true
	h.Frame()
	waitFor("restarted poll does not check", func() bool { return checks.Load() > n })
	win.closed.Store(true)
	time.Sleep(20 * time.Millisecond)
	n = checks.Load()
	time.Sleep(30 * time.Millisecond)
	if checks.Load() != n {
		t.Fatal("poll outlived its window")
	}
}

// Without a window, as in a screenshot, Poll does nothing.
func TestPollWithoutWindow(t *testing.T) {
	var checks atomic.Int32
	root := Root(viewFunc(func(cx *Context) Element {
		cx.Poll("check", time.Millisecond, func() bool { checks.Add(1); return false })
		return Div()
	}))
	h := uitest.NewFunc(func(gtx core.C) { root.Layout(gtx) })
	h.Frame()
	time.Sleep(20 * time.Millisecond)
	if checks.Load() != 0 {
		t.Fatal("poll ran without a window")
	}
}
