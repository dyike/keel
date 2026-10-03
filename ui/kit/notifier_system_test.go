package kit

import (
	"errors"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

type noticeBackendRequest struct {
	id, title, body string
	remove          bool
	done            func(error)
}
type noticeBackendProbe struct{ requests chan noticeBackendRequest }

func newNoticeBackendProbe() *noticeBackendProbe {
	return &noticeBackendProbe{requests: make(chan noticeBackendRequest, 20)}
}
func (b *noticeBackendProbe) Post(id, title, body string, done func(error)) {
	b.requests <- noticeBackendRequest{id: id, title: title, body: body, done: done}
}
func (b *noticeBackendProbe) Remove(id string, done func(error)) {
	b.requests <- noticeBackendRequest{id: id, remove: true, done: done}
}
func (b *noticeBackendProbe) next(t *testing.T) noticeBackendRequest {
	t.Helper()
	select {
	case r := <-b.requests:
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("backend request missing")
		return noticeBackendRequest{}
	}
}

func TestNotifierSystemModesAndSerializedRemoval(t *testing.T) {
	b := newNoticeBackendProbe()
	n := Notifier().SystemBackend(b, nil)
	closed := 0
	n.Notify(Notice{Title: "Local", Timeout: -1})
	id := n.Notify(Notice{Title: "Only system", Delivery: NoticeSystemOnly, OnClose: func() { closed++ }})
	both := n.Notify(Notice{Title: "Both", Body: "Plain", Delivery: NoticeInAppAndSystem, Timeout: -1, OnClose: func() { closed++ }})
	h := renderNotifierContent(n, 600, 1)
	if !shown(h, "Local") || !shown(h, "Both") || shown(h, "Only system") {
		t.Fatal("delivery modes rendered incorrectly")
	}
	first := b.next(t)
	if first.title != "Only system" {
		t.Fatal("local notice reached backend")
	}
	n.Dismiss(id)
	// The next post and removal must wait for completion, not just Post returning.
	select {
	case <-b.requests:
		t.Fatal("overlapping backend requests")
	default:
	}
	first.done(nil)
	second := b.next(t)
	if second.title != "Both" || second.body != "Plain" {
		t.Fatal("post order")
	}
	second.done(nil)
	removal := b.next(t)
	if !removal.remove || removal.id != first.id {
		t.Fatal("wrong system removal")
	}
	removal.done(nil)
	n.Dismiss(both)
	last := b.next(t)
	last.done(nil)
	if !last.remove || last.id != second.id || closed != 1 {
		t.Fatal("system-only close callback or both removal", closed)
	}
}

func TestNotifierBothExpiresLocallyAndRetainsIdentity(t *testing.T) {
	b := newNoticeBackendProbe()
	n := Notifier().SystemBackend(b, nil)
	c := &clock{now: time.Unix(100, 0)}
	closed := 0
	id := n.NotifyKey("job", Notice{Title: "First", Delivery: NoticeInAppAndSystem, Timeout: time.Second, OnClose: func() { closed++ }})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	first := b.next(t)
	first.done(nil)
	c.advance(h, time.Second)
	h.Frame()
	if shown(h, "First") || closed != 1 || n.Len() != 1 {
		t.Fatal("in-app expiry removed system identity")
	}
	if next := n.NotifyKey("job", Notice{Title: "Updated", Delivery: NoticeInAppAndSystem, Timeout: -1}); next != id {
		t.Fatal("expiry lost key")
	}
	second := b.next(t)
	if second.remove || second.id != first.id || second.title != "Updated" {
		t.Fatal("expiry retracted system or changed ID")
	}
	second.done(nil)
	h.Frame()
	if !shown(h, "Updated") {
		t.Fatal("replacement did not reappear")
	}
	n.Clear()
	last := b.next(t)
	last.done(nil)
	if !last.remove || last.id != first.id || closed != 1 {
		t.Fatal("clear of updated notice")
	}
}

func TestNotifierSystemTransitionsAndResults(t *testing.T) {
	b := newNoticeBackendProbe()
	other := newNoticeBackendProbe()
	result := []NoticeSystemResult{}
	n := Notifier().SystemBackend(b, func(r NoticeSystemResult) { result = append(result, r) }).Delivery(NoticeSystemOnly)
	id := n.Notify(Notice{Title: "Default system"})
	h := renderNotifierContent(n, 500, 1)
	request := b.next(t)
	failure := errors.New("denied")
	request.done(failure)
	request.done(nil)
	h.Frame()
	if len(result) != 1 || result[0].ID != id || n.SystemError() != failure {
		t.Fatal("result dispatch or duplicate completion", result)
	}
	n.SystemBackend(other, nil)
	n.Update(id, Notice{Title: "Local now", Delivery: NoticeInApp, Timeout: -1})
	removal := b.next(t)
	if !removal.remove {
		t.Fatal("mode switch did not retract")
	}
	removal.done(nil)
	h.Frame()
	if !shown(h, "Local now") {
		t.Fatal("system to local transition")
	}
	newID := n.Notify(Notice{Title: "New backend"})
	newRequest := other.next(t)
	newRequest.done(nil)
	if newID == id || newRequest.id == request.id {
		t.Fatal("identity reused")
	}
	n.Clear()
	r := other.next(t)
	r.done(nil)
}

func TestNotifierSystemEmptyAndMissingBackend(t *testing.T) {
	b := newNoticeBackendProbe()
	n := Notifier().SystemBackend(b, nil)
	n.Notify(Notice{Delivery: NoticeSystemOnly, Content: Label("Rich only")})
	n.Notify(Notice{Title: "Sentinel", Delivery: NoticeSystemOnly})
	r := b.next(t)
	r.done(nil)
	if r.title != "Sentinel" {
		t.Fatal("empty system notification sent")
	}
	n.Clear()
	r = b.next(t)
	r.done(nil)
	missing := Notifier()
	missing.Notify(Notice{Title: "No backend", Delivery: NoticeInAppAndSystem, Timeout: -1})
	h := renderNotifierContent(missing, 500, 1)
	deadline := time.Now().Add(3 * time.Second)
	for missing.SystemError() == nil && time.Now().Before(deadline) {
		h.Frame()
		time.Sleep(time.Millisecond)
	}
	if !errors.Is(missing.SystemError(), ErrNoticeSystemUnavailable) || !shown(h, "No backend") {
		t.Fatal("missing backend error or in-app fallback")
	}
}
