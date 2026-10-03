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
	owner   string
	markup  bool
	fail    string
	replace []uint32
	closed  []uint32
	body    string
	next    uint32
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
		_ = args[5].([]string)
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
