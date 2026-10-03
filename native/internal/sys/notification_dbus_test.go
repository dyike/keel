//go:build (linux && !android) || (darwin && !ios)

package sys

import (
	"errors"
	"github.com/dyike/keel/native"
	"github.com/godbus/dbus/v5"
	"strings"
	"testing"
)

type notificationFakeBus struct {
	actions     bool
	sentActions []string
	owner       string
	markup      bool
	fail        string
	replace     []uint32
	closed      []uint32
	body        string
	next        uint32
}

func (f *notificationFakeBus) call(dest, method string, args ...any) *dbus.Call {
	if f.fail == method {
		return &dbus.Call{Err: errors.New("bus failure")}
	}
	switch method {
	case "org.freedesktop.DBus.GetNameOwner":
		return &dbus.Call{Body: []any{f.owner}}
	case notificationService + ".GetCapabilities":
		caps := []string{"body"}
		if f.actions {
			caps = append(caps, "actions")
		}
		if f.markup {
			caps = append(caps, "body-markup")
		}
		return &dbus.Call{Body: []any{caps}}
	case notificationService + ".Notify":
		if dest != f.owner || len(args) != 8 {
			panic("invalid destination/arguments")
		}
		id := args[1].(uint32)
		f.replace = append(f.replace, id)
		if id == 0 {
			f.next++
			id = f.next
		}
		f.body = args[4].(string)
		f.sentActions = append([]string(nil), args[5].([]string)...)
		_ = args[6].(map[string]dbus.Variant)
		if args[7].(int32) != -1 {
			panic("wrong default timeout")
		}
		return &dbus.Call{Body: []any{id}}
	case notificationService + ".CloseNotification":
		if dest != f.owner {
			panic("wrong owner")
		}
		f.closed = append(f.closed, args[0].(uint32))
		return &dbus.Call{}
	}
	panic(method)
}
func TestNotificationDBusReplacementAndRemoval(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3", markup: true}
	n := notificationDBus{call: f.call}
	for i := 0; i < 2; i++ {
		if err := n.post("a", "Title", "<b>A & B</b>"); err != nil {
			t.Fatal(err)
		}
	}
	if f.replace[0] != 0 || f.replace[1] != 1 || !strings.Contains(f.body, "&lt;b&gt;") {
		t.Fatal("replacement/plain text", f.replace, f.body)
	}
	if err := n.remove("a"); err != nil {
		t.Fatal(err)
	}
	if len(f.closed) != 1 || f.closed[0] != 1 {
		t.Fatal("wrong removal")
	}
	if err := n.remove("a"); err != nil || len(f.closed) != 1 {
		t.Fatal("repeated remove")
	}
}
func TestNotificationDBusClosedAndRestart(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3"}
	n := notificationDBus{call: f.call}
	if err := n.post("a", "Title", "<literal>"); err != nil {
		t.Fatal(err)
	}
	if f.body != "<literal>" {
		t.Fatal("escaped body without markup capability")
	}
	n.closed(":1.2", 1)
	if n.ids["a"] != 1 {
		t.Fatal("foreign signal removed notification")
	}
	n.closed(f.owner, 1)
	if err := n.post("a", "Title", ""); err != nil || f.replace[1] != 0 {
		t.Fatal("closed ID reused")
	}
	f.owner = ":1.4"
	if err := n.remove("a"); err != nil || len(f.closed) != 0 {
		t.Fatal("old daemon ID sent to new daemon")
	}
	if err := n.post("a", "Title", ""); err != nil || f.replace[2] != 0 {
		t.Fatal("restart replacement ID")
	}
}
func TestNotificationDBusErrorsPreserveState(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3"}
	n := notificationDBus{call: f.call}
	if err := n.post("a", "Title", ""); err != nil {
		t.Fatal(err)
	}
	f.fail = notificationService + ".Notify"
	if err := n.post("a", "New", ""); !errors.Is(err, native.ErrFailed) || n.ids["a"] != 1 {
		t.Fatal("failed replacement lost ID", err)
	}
	f.fail = notificationService + ".CloseNotification"
	if err := n.remove("a"); !errors.Is(err, native.ErrFailed) || n.ids["a"] != 1 {
		t.Fatal("failed remove lost ID", err)
	}
	f.fail = "org.freedesktop.DBus.GetNameOwner"
	if err := n.post("b", "Title", ""); !errors.Is(err, native.ErrUnsupported) {
		t.Fatal("missing service", err)
	}
	f.fail = notificationService + ".GetCapabilities"
	f.owner = ":1.4"
	if err := n.post("b", "Title", ""); !errors.Is(err, native.ErrFailed) || n.owner != "" {
		t.Fatal("failed capability lookup cached", err)
	}
}

func TestNotificationDBusActions(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3", actions: true}
	n := notificationDBus{call: f.call}
	calls := 0
	if err := n.postInteractive("a", "Title", "", func() { calls++ }); err != nil {
		t.Fatal(err)
	}
	if len(f.sentActions) != 2 || f.sentActions[0] != "default" {
		t.Fatal("missing default action", f.sentActions)
	}
	if n.activated(":1.2", 1, "default") != nil || n.activated(f.owner, 1, "unknown") != nil || n.activated(f.owner, 99, "default") != nil {
		t.Fatal("foreign or invalid action accepted")
	}
	fn := n.activated(f.owner, 1, "default")
	if fn == nil {
		t.Fatal("missing action")
	}
	fn()
	n.closed(f.owner, 1)
	if calls != 1 || n.activated(f.owner, 1, "default") != nil || len(n.clicks) != 0 {
		t.Fatal("closed/repeated action")
	}
	if err := n.postInteractive("a", "Title", "", func() { calls++ }); err != nil {
		t.Fatal(err)
	}
	n.closed(f.owner, n.ids["a"])
	if n.activated(f.owner, 2, "default") != nil {
		t.Fatal("late action after close")
	}
}
func TestNotificationDBusActionReplacementAndCapabilities(t *testing.T) {
	f := &notificationFakeBus{owner: ":1.3", actions: true}
	n := notificationDBus{call: f.call}
	calls := 0
	n.postInteractive("a", "Old", "", func() { calls++ })
	f.fail = notificationService + ".Notify"
	if err := n.postInteractive("a", "New", "", func() { calls += 10 }); err == nil {
		t.Fatal("expected failure")
	}
	fn := n.activated(f.owner, 1, "default")
	if fn == nil {
		t.Fatal("failure lost callback")
	}
	fn()
	if calls != 1 {
		t.Fatal("failed replacement installed callback")
	}
	f.fail = ""
	n.postInteractive("a", "New", "", func() { calls += 10 })
	n.post("a", "Plain", "")
	if n.activated(f.owner, 1, "default") != nil {
		t.Fatal("plain replacement retained callback")
	}
	n.postInteractive("a", "New", "", func() { calls += 10 })
	if err := n.remove("a"); err != nil || len(n.clicks) != 0 {
		t.Fatal("remove retained callback", err)
	}
	f.owner = ":1.4"
	f.actions = false
	if err := n.postInteractive("b", "Unsupported", "", func() {}); !errors.Is(err, native.ErrUnsupported) {
		t.Fatal("missing capability not reported", err)
	}
	if len(n.clicks) != 0 {
		t.Fatal("restart retained callbacks")
	}
	if err := n.post("b", "Plain", ""); err != nil {
		t.Fatal("plain notification requires actions", err)
	}
}
