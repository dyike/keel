package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestNotifierCardClickAndChildIsolation(t *testing.T) {
	for _, scale := range []int{1, 2} {
		n := Notifier()
		cards, actions, closes := 0, 0, 0
		disabled := false
		n.Notify(Notice{Title: "Clickable", Body: "Details", Timeout: -1,
			OnClick: func() { cards++ }, OnClose: func() { closes++ },
			Action:  Button("Action", func() { actions++ }),
			Content: Button("Content action", func() { actions++ }),
		})
		h := renderNotifierContent(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(n.Render(cx)) }), 500, scale)
		h.Frame()
		// The title is noninteractive and belongs to the card's activation area.
		b := bounds(h, "Clickable")
		x, y := float32(b.Min.X+50*scale), float32(b.Min.Y+20*scale)
		h.Click(x, y)
		h.Frame()
		if cards != 1 || n.Len() != 1 {
			t.Fatal("card activation")
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if cards != 2 {
			t.Fatal("keyboard activation")
		}
		click(t, h, "Action")
		h.Frame()
		click(t, h, "Content action")
		h.Frame()
		if actions != 2 || cards != 2 {
			t.Fatal("child click bubbled", cards, actions)
		}
		disabled = true
		h.Frame()
		h.Click(x, y)
		h.Frame()
		if cards != 2 {
			t.Fatal("disabled card activated")
		}
		disabled = false
		h.Frame()
		click(t, h, "关闭 Clickable")
		h.Frame()
		if cards != 2 || closes != 1 || n.Len() != 0 {
			t.Fatal("close click bubbled or missed callback")
		}
	}
}

func TestNotifierCloseReentrantAndReplacement(t *testing.T) {
	n := Notifier()
	old, newCalls := 0, 0
	id := n.Notify(Notice{Title: "Old", OnClose: func() { old++ }})
	n.Update(id, Notice{Title: "New", OnClose: func() { newCalls++ }})
	if old != 0 || newCalls != 0 {
		t.Fatal("update closed notice")
	}
	n.Dismiss(id)
	n.Dismiss(id)
	if old != 0 || newCalls != 1 {
		t.Fatal("replacement close callbacks")
	}
	var nested int
	nested = n.Notify(Notice{OnClose: func() {
		if n.Len() != 0 || n.Update(nested, Notice{}) {
			t.Fatal("callback ran before removal")
		}
		n.Dismiss(nested)
		n.Notify(Notice{Title: "From callback", Timeout: -1})
	}})
	n.Dismiss(nested)
	if n.Len() != 1 || n.items[0].Title != "From callback" {
		t.Fatal("reentrant mutation lost")
	}
	for i := 0; i < MaxNotifications; i++ {
		n.Notify(Notice{Timeout: -1})
	}
	queued := n.Notify(Notice{OnClose: func() { newCalls++ }})
	n.Dismiss(queued)
	if newCalls != 2 {
		t.Fatal("queued dismiss callback")
	}
}

func TestNotifierTimeoutCloseCallback(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	calls := 0
	id := n.Notify(Notice{Title: "Expires", Timeout: time.Second, OnClose: func() { calls++; n.Notify(Notice{Title: "Replacement", Timeout: -1}) }})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	c.advance(h, time.Second)
	h.Frame()
	n.Dismiss(id)
	c.advance(h, 5*time.Second)
	if calls != 1 || n.Len() != 1 || !shown(h, "Replacement") {
		t.Fatal("timeout callback mutation", calls, n.Len())
	}
}
