package kit

import (
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
	"time"
)

func TestNotifierKeyReplacementAndQueue(t *testing.T) {
	n := Notifier()
	closes := 0
	first := n.NotifyKey("upload/1", Notice{Title: "Uploading", Timeout: -1, OnClose: func() { closes++ }})
	for i := 0; i < MaxNotifications-1; i++ {
		n.Notify(Notice{Title: "Other", Timeout: -1})
	}
	queued := n.NotifyKey("upload/2", Notice{Title: "Queued", Timeout: -1})
	h := renderNotifierContent(n, 600, 1)
	if n.NotifyKey("upload/1", Notice{Title: "Completed", Timeout: -1}) != first || closes != 0 || n.Len() != MaxNotifications+1 {
		t.Fatal("key replacement changed identity or closed")
	}
	if n.NotifyKey("upload/2", Notice{Title: "Queued replacement", Timeout: -1}) != queued {
		t.Fatal("queued identity changed")
	}
	h.Frame()
	if !shown(h, "Completed") || shown(h, "Queued replacement") {
		t.Fatal("queue order changed")
	}
	n.DismissKey("upload/1")
	h.Frame()
	if !shown(h, "Queued replacement") {
		t.Fatal("queued replacement not released")
	}
	if !n.Update(queued, Notice{Title: "Updated by ID", Timeout: -1}) || !n.DismissKey("upload/2") {
		t.Fatal("Update lost key")
	}
	if n.DismissKey("upload/2") || n.DismissKey("") {
		t.Fatal("unknown key removed")
	}
	if n.NotifyKey("upload/2", Notice{}) == queued {
		t.Fatal("dismissed ID reused")
	}
	a, b := n.NotifyKey("", Notice{}), n.NotifyKey("", Notice{})
	if a == b {
		t.Fatal("anonymous notices coalesced")
	}
	other := Notifier()
	other.NotifyKey("upload/2", Notice{Title: "Other host"})
	if other.Len() != 1 || other.items[0].Title != "Other host" {
		t.Fatal("key escaped host")
	}
}

func TestNotifierKeyReplacementRestartsCountdown(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	n.NotifyKey("task", Notice{Title: "Working", Timeout: 5 * time.Second})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	c.advance(h, 4*time.Second)
	closes := 0
	n.NotifyKey("task", Notice{Title: "Done", Timeout: 5 * time.Second, OnClose: func() { closes++ }})
	h.Frame()
	c.advance(h, 2*time.Second)
	if n.Len() != 1 || closes != 0 {
		t.Fatal("old timer survived replacement")
	}
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 || closes != 1 {
		t.Fatal("new countdown not used")
	}
}

func TestNotifierClearSnapshotAndReentrancy(t *testing.T) {
	n := Notifier()
	order := []int{}
	first := 0
	first = n.NotifyKey("same", Notice{OnClose: func() {
		if n.Len() != 0 || n.Update(first, Notice{}) {
			t.Fatal("callbacks preceded removal")
		}
		n.Dismiss(first)
		order = append(order, 1)
		n.NotifyKey("same", Notice{Title: "New", Timeout: -1})
	}})
	n.Notify(Notice{OnClose: func() { order = append(order, 2) }})
	if count := n.Clear(); count != 2 {
		t.Fatal("clear count", count)
	}
	if !reflect.DeepEqual(order, []int{1, 2}) || n.Len() != 1 || n.items[0].Title != "New" || n.items[0].id == first {
		t.Fatal("clear snapshot/reentrancy", order)
	}
	if n.Clear() != 1 || n.Clear() != 0 {
		t.Fatal("empty clear")
	}
}

func TestNotifierClearCancelsVisibleAndQueuedTimers(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	closes := 0
	for i := 0; i < MaxNotifications+2; i++ {
		n.Notify(Notice{Title: "Old", Timeout: time.Second, OnClose: func() { closes++ }})
	}
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	n.Clear()
	n.Notify(Notice{Title: "New", Timeout: -1})
	h.Frame()
	c.advance(h, 5*time.Second)
	if closes != MaxNotifications+2 || n.Len() != 1 || !shown(h, "New") {
		t.Fatal("clear left old timer/callback", closes)
	}
}
