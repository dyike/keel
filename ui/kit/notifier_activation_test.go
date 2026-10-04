package kit

import (
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

type interactiveNoticeProbe struct {
	*noticeBackendProbe
	activations chan func()
}

func (b *interactiveNoticeProbe) PostInteractive(id, title, body string, activate func(), done func(error)) {
	b.activations <- activate
	b.Post(id, title, body, done)
}
func TestNotifierSystemActivationLifecycle(t *testing.T) {
	for _, mode := range []NoticeDelivery{NoticeSystemOnly, NoticeInAppAndSystem} {
		b := &interactiveNoticeProbe{newNoticeBackendProbe(), make(chan func(), 8)}
		order := []string{}
		n := Notifier().SystemBackend(b, nil)
		n.OnSystemActivate(func() {
			order = append(order, "raise")
			if n.Len() != 0 {
				t.Fatal("not detached before activation")
			}
		})
		id := n.Notify(Notice{Title: "Open", Delivery: mode, Timeout: -1, OnClose: func() { order = append(order, "close") }, OnClick: func() { order = append(order, "click") }})
		h := renderNotifierContent(n, 500, 1)
		r := b.next(t)
		activate := <-b.activations
		r.done(nil)
		activate()
		activate()
		h.Frame()
		want := []string{"raise", "click"}
		if mode == NoticeInAppAndSystem {
			want = []string{"raise", "close", "click"}
		}
		if !reflect.DeepEqual(order, want) || n.Len() != 0 || shown(h, "Open") {
			t.Fatal("activation order/deduplication", order)
		}
		remove := b.next(t)
		remove.done(nil)
		if !remove.remove || remove.id != r.id {
			t.Fatal("activation did not retract")
		}
		n.Dismiss(id)
		activate()
		h.Frame()
		if !reflect.DeepEqual(order, want) {
			t.Fatal("stale activation")
		}
	}
}
func TestNotifierSystemActivationAfterExpiryUsesLatestCallback(t *testing.T) {
	b := &interactiveNoticeProbe{newNoticeBackendProbe(), make(chan func(), 8)}
	n := Notifier().SystemBackend(b, nil)
	closed, old, latest := 0, 0, 0
	id := n.Notify(Notice{Title: "Old", Delivery: NoticeInAppAndSystem, Timeout: -1, OnClick: func() { old++ }})
	h := renderNotifierContent(n, 500, 1)
	r := b.next(t)
	activate := <-b.activations
	r.done(nil)
	n.Update(id, Notice{Title: "New", Delivery: NoticeInAppAndSystem, Timeout: -1, OnClose: func() { closed++ }, OnClick: func() { latest++; n.Notify(Notice{Title: "Opened", Timeout: -1}) }})
	r = b.next(t)
	<-b.activations
	r.done(nil)
	n.expire(id)
	h.Frame()
	activate()
	h.Frame()
	r = b.next(t)
	r.done(nil)
	if closed != 1 || old != 0 || latest != 1 || !shown(h, "Opened") {
		t.Fatal("expired activation/reentrant callback", closed, old, latest)
	}
}
func TestNotifierSystemActivationIgnoredAfterLocalTransition(t *testing.T) {
	b := &interactiveNoticeProbe{newNoticeBackendProbe(), make(chan func(), 8)}
	n := Notifier().SystemBackend(b, nil)
	calls := 0
	id := n.Notify(Notice{Title: "System", Delivery: NoticeSystemOnly, OnClick: func() { calls++ }})
	h := renderNotifierContent(el.ViewFunc(n.Render), 500, 1)
	r := b.next(t)
	activate := <-b.activations
	r.done(nil)
	n.Update(id, Notice{Title: "Local", Delivery: NoticeInApp, Timeout: -1})
	r = b.next(t)
	r.done(nil)
	activate()
	h.Frame()
	if calls != 0 || !shown(h, "Local") || n.Len() != 1 {
		t.Fatal("stale system event removed local replacement")
	}
}

type tokenNoticeProbe struct {
	*interactiveNoticeProbe
	tokens chan func(NoticeActivation)
}

func (b *tokenNoticeProbe) PostActivated(id, title, body string, activate func(NoticeActivation), done func(error)) {
	b.tokens <- activate
	b.Post(id, title, body, done)
}
func TestNotifierActivationDataAndOrder(t *testing.T) {
	b := &tokenNoticeProbe{&interactiveNoticeProbe{newNoticeBackendProbe(), make(chan func(), 2)}, make(chan func(NoticeActivation), 2)}
	var order []string
	n := Notifier().SystemBackend(b, nil)
	n.OnSystemActivation(func(a NoticeActivation) {
		order = append(order, a.Token)
		if n.Len() != 0 {
			t.Fatal("callback reentrancy before detach")
		}
	})
	n.OnSystemActivate(func() { order = append(order, "raise") })
	n.Notify(Notice{Title: "Token", Delivery: NoticeInAppAndSystem, Timeout: -1, OnClose: func() { order = append(order, "close") }, OnClick: func() { order = append(order, "click") }})
	h := renderNotifierContent(n, 500, 1)
	request := b.next(t)
	activate := <-b.tokens
	if len(b.activations) != 0 {
		t.Fatal("data-capable backend not preferred")
	}
	request.done(nil)
	activate(NoticeActivation{Token: "opaque-token"})
	activate(NoticeActivation{Token: "duplicate"})
	if len(order) != 0 {
		t.Fatal("backend callback ran outside UI update")
	}
	h.Frame()
	if !reflect.DeepEqual(order, []string{"opaque-token", "raise", "close", "click"}) {
		t.Fatal(order)
	}
	removal := b.next(t)
	removal.done(nil)
	if !removal.remove || removal.id != request.id {
		t.Fatal("activation did not remove notification")
	}
}

func TestNotifierActivationWithoutToken(t *testing.T) {
	b := &interactiveNoticeProbe{newNoticeBackendProbe(), make(chan func(), 2)}
	calls := 0
	n := Notifier().SystemBackend(b, nil).OnSystemActivation(func(a NoticeActivation) {
		calls++
		if a.Token != "" {
			t.Fatal("invented token", a)
		}
	})
	n.Notify(Notice{Title: "Plain", Delivery: NoticeSystemOnly})
	h := renderNotifierContent(n, 500, 1)
	request := b.next(t)
	activate := <-b.activations
	request.done(nil)
	activate()
	h.Frame()
	if calls != 1 {
		t.Fatal("missing empty-token activation", calls)
	}
	b.next(t).done(nil)
}
