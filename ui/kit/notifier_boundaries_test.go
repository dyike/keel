package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

func TestNotifierHoverPreservesRemainingTime(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	n.Notify(Notice{Title: "Notice", Timeout: 5 * time.Second})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	c.advance(h, 2*time.Second)
	x, y := center(bounds(h, "Notice"))
	h.Move(x, y)
	c.advance(h, 20*time.Second)
	if n.Len() != 1 {
		t.Fatal("hover timeout")
	}
	h.Move(1, 299)
	c.advance(h, 2900*time.Millisecond)
	if n.Len() != 1 {
		t.Fatal("resumed early")
	}
	c.advance(h, 100*time.Millisecond)
	h.Frame()
	if n.Len() != 0 {
		t.Fatal("hover restarted full timeout")
	}
}
func TestNotifierBackgroundUpdateResetsAndKeepsOrder(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	id := n.Notify(Notice{Title: "Uploading", Timeout: 5 * time.Second})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	c.advance(h, 4*time.Second)
	done := make(chan struct{})
	go func() {
		core.Update(func() { n.Update(id, Notice{Title: "Completed", Tone: ToneSuccess, Timeout: 5 * time.Second}) })
		close(done)
	}()
	<-done
	h.Frame()
	if n.Len() != 1 || !shown(h, "Completed") || shown(h, "Uploading") {
		t.Fatal("background update")
	}
	c.advance(h, 2*time.Second)
	if n.Len() != 1 {
		t.Fatal("old timer dismissed replacement")
	}
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 || n.Update(id, Notice{Title: "gone"}) {
		t.Fatal("updated timeout or unknown id")
	}
}
func TestNotifierDisabledOwnerPausesWithoutReset(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	n.Notify(Notice{Title: "Notice", Timeout: 5 * time.Second})
	disabled := false
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(n.Render(cx)) })
	c.advance(h, 2*time.Second)
	disabled = true
	h.Frame()
	c.advance(h, 20*time.Second)
	if n.Len() != 1 {
		t.Fatal("disabled timeout")
	}
	disabled = false
	h.Frame()
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 {
		t.Fatal("remaining timeout not restored")
	}
}

func TestNotifierKeyboardFocusPausesCountdown(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	n.Notify(Notice{Title: "Notice", Timeout: 5 * time.Second})
	var cx *el.Context
	h := c.harness(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().Child(n.Render(ctx)) })
	c.advance(h, 2*time.Second)
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Frame()
	c.advance(h, 20*time.Second)
	if n.Len() != 1 {
		t.Fatal("focused close button timed out")
	}
	cx.Focus("")
	h.Frame()
	h.Frame()
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 {
		t.Fatal("focus pause lost remaining delay")
	}
}
