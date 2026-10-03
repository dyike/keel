//go:build (linux && !android) || (darwin && !ios)

package sys

import (
	"fmt"
	"github.com/dyike/keel/native"
	"github.com/godbus/dbus/v5"
	"html"
)

const notificationService = "org.freedesktop.Notifications"

// Access is serialized by the Linux connection lock, including signal handling.
// The transport seam keeps protocol/state tests independent of a desktop daemon.
type notificationDBus struct {
	call   func(destination, method string, args ...any) *dbus.Call
	owner  string
	markup bool
	ids    map[string]uint32
}

func (n *notificationDBus) refresh() error {
	var owner string
	if err := n.call("org.freedesktop.DBus", "org.freedesktop.DBus.GetNameOwner", notificationService).Store(&owner); err != nil {
		return fmt.Errorf("%w: notification service: %v", native.ErrUnsupported, err)
	}
	if owner != n.owner {
		n.owner = owner
		n.ids = make(map[string]uint32)
		n.markup = false
		var caps []string
		if err := n.call(owner, notificationService+".GetCapabilities").Store(&caps); err != nil {
			n.owner = ""
			return notificationDBusError(err)
		}
		for _, cap := range caps {
			if cap == "body-markup" {
				n.markup = true
			}
		}
	}
	return nil
}
func notificationDBusError(err error) error {
	return fmt.Errorf("%w: notification service: %v", native.ErrFailed, err)
}
func (n *notificationDBus) post(key, title, body string) error {
	if err := n.refresh(); err != nil {
		return err
	}
	if n.markup {
		body = html.EscapeString(body)
	}
	var id uint32
	err := n.call(n.owner, notificationService+".Notify", "Keel", n.ids[key], "", title, body, []string{}, map[string]dbus.Variant{}, int32(-1)).Store(&id)
	if err != nil {
		return notificationDBusError(err)
	}
	if id == 0 {
		return notificationDBusError(fmt.Errorf("invalid zero notification ID"))
	}
	n.ids[key] = id
	return nil
}
func (n *notificationDBus) remove(key string) error {
	if err := n.refresh(); err != nil {
		return err
	}
	id, ok := n.ids[key]
	if !ok {
		return nil
	}
	if err := n.call(n.owner, notificationService+".CloseNotification", id).Err; err != nil {
		return notificationDBusError(err)
	}
	delete(n.ids, key)
	return nil
}
func (n *notificationDBus) closed(sender string, id uint32) {
	if sender != n.owner {
		return
	}
	for key, current := range n.ids {
		if current == id {
			delete(n.ids, key)
		}
	}
}
