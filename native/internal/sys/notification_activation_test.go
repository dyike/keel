//go:build (linux && !android) || (darwin && !ios)

package sys

import (
	"github.com/godbus/dbus/v5"
	"reflect"
	"testing"
)

func noticeSignal(owner, name string, body ...any) *dbus.Signal {
	return &dbus.Signal{Sender: owner, Path: "/org/freedesktop/Notifications", Name: notificationService + "." + name, Body: body}
}
func TestNotificationActivationTokensRouteAndConsume(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3", actions: true}
	n := notificationDBus{call: f.call}
	var got []string
	for _, id := range []string{"a", "b"} {
		if err := n.postActivated(id, "Title", "", func(token string) { got = append(got, id+":"+token) }); err != nil {
			t.Fatal(err)
		}
	}
	n.signal(noticeSignal(f.owner, "ActivationToken", uint32(2), "wayland-b"))
	n.signal(noticeSignal(f.owner, "ActivationToken", uint32(1), "startup-a"))
	first := n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "default"))
	if first == nil || len(got) != 0 {
		t.Fatal("callback did not detach for lock-free execution")
	}
	n.signal(noticeSignal(f.owner, "ActivationToken", uint32(1), "late"))
	second := n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(2), "default"))
	if second == nil {
		t.Fatal("missing second activation")
	}
	second()
	first()
	if !reflect.DeepEqual(got, []string{"b:wayland-b", "a:startup-a"}) || len(n.tokens) != 0 {
		t.Fatal(got, n.tokens)
	}
	if n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "default")) != nil {
		t.Fatal("duplicate callback")
	}
	if err := n.postActivated("a", "Title", "", func(token string) { got = append(got, "absent:"+token) }); err != nil {
		t.Fatal(err)
	}
	n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "default"))()
	if got[2] != "absent:" {
		t.Fatal("absent token reused previous value", got)
	}
}

func TestNotificationActivationRejectsMalformedAndForeignSignals(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3", actions: true}
	n := notificationDBus{call: f.call}
	value := ""
	if err := n.postActivated("a", "Title", "", func(token string) { value = token }); err != nil {
		t.Fatal(err)
	}
	badPath := noticeSignal(f.owner, "ActivationToken", uint32(1), "bad")
	badPath.Path = "/unrelated"
	for _, signal := range []*dbus.Signal{
		nil, badPath, noticeSignal(":1.4", "ActivationToken", uint32(1), "bad"),
		noticeSignal(f.owner, "ActivationToken", uint32(99), "bad"),
		noticeSignal(f.owner, "ActivationToken", int32(1), "bad"),
		noticeSignal(f.owner, "ActivationToken", uint32(1), uint32(9)),
		noticeSignal(f.owner, "ActivationToken", uint32(1)),
		noticeSignal(f.owner, "ActivationToken", uint32(1), "bad", "extra"),
		noticeSignal(f.owner, "NotificationClosed", uint32(1), "bad"),
		noticeSignal(f.owner, "ActionInvoked", uint32(1), uint32(1)),
	} {
		if fn := n.signal(signal); fn != nil {
			t.Fatal("invalid signal invoked callback")
		}
	}
	if len(n.tokens) != 0 || n.ids["a"] != 1 {
		t.Fatal("invalid signal mutated state")
	}
	n.signal(noticeSignal(f.owner, "ActivationToken", uint32(1), "valid"))
	if n.signal(noticeSignal(":1.4", "ActionInvoked", uint32(1), "default")) != nil {
		t.Fatal("foreign activation")
	}
	fn := n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "default"))
	if fn == nil {
		t.Fatal("valid handler was lost")
	}
	fn()
	if value != "valid" {
		t.Fatal(value)
	}
}

func TestNotificationActivationTokenLifecycle(t *testing.T) {
	for _, mode := range []string{"replace", "plain", "remove", "closed", "restart", "failed-replace", "failed-remove", "unknown-action", "empty"} {
		t.Run(mode, func(t *testing.T) {
			f := &notificationFakeBus{owner: ":1.3", actions: true}
			n := notificationDBus{call: f.call}
			got := "not-called"
			fn := func(token string) { got = token }
			if err := n.postActivated("a", "Title", "", fn); err != nil {
				t.Fatal(err)
			}
			n.signal(noticeSignal(f.owner, "ActivationToken", uint32(1), "token"))
			switch mode {
			case "replace":
				n.postActivated("a", "Replacement", "", fn)
			case "plain":
				n.post("a", "Plain", "")
			case "remove":
				n.remove("a")
			case "closed":
				n.signal(noticeSignal(f.owner, "NotificationClosed", uint32(1), uint32(2)))
			case "restart":
				f.owner = ":1.9"
				n.refresh()
			case "failed-replace":
				f.fail = notificationService + ".Notify"
				if n.postActivated("a", "Replacement", "", fn) == nil {
					t.Fatal("expected failure")
				}
			case "failed-remove":
				f.fail = notificationService + ".CloseNotification"
				if n.remove("a") == nil {
					t.Fatal("expected failure")
				}
			case "unknown-action":
				if n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "unknown")) != nil {
					t.Fatal("unknown action activated")
				}
			case "empty":
				n.signal(noticeSignal(f.owner, "ActivationToken", uint32(1), ""))
			}
			if mode != "failed-replace" && mode != "failed-remove" && len(n.tokens) != 0 {
				t.Fatal("stale token", n.tokens)
			}
			activation := n.signal(noticeSignal(f.owner, "ActionInvoked", uint32(1), "default"))
			if activation != nil {
				activation()
			}
			want := "not-called"
			switch mode {
			case "replace", "unknown-action", "empty":
				want = ""
			case "failed-replace", "failed-remove":
				want = "token"
			}
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	}
}
