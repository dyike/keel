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
	call    func(destination, method string, args ...any) *dbus.Call
	owner   string
	markup  bool
	actions bool
	clicks  map[string]func(string)
	tokens  map[string]string
	ids     map[string]uint32
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
		n.actions = false
		n.clicks = make(map[string]func(string))
		n.tokens = make(map[string]string)
		var caps []string
		if err := n.call(owner, notificationService+".GetCapabilities").Store(&caps); err != nil {
			n.owner = ""
			return notificationDBusError(err)
		}
		for _, cap := range caps {
			if cap == "actions" {
				n.actions = true
			}
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
	return n.postInteractive(key, title, body, nil)
}
func (n *notificationDBus) postInteractive(key, title, body string, onClick func()) error {
	var activate func(string)
	if onClick != nil {
		activate = func(string) { onClick() }
	}
	return n.postActivated(key, title, body, activate)
}
func (n *notificationDBus) postActivated(key, title, body string, onClick func(string)) error {
	if err := n.refresh(); err != nil {
		return err
	}
	if onClick != nil && !n.actions {
		return fmt.Errorf("%w: notification server has no actions", native.ErrUnsupported)
	}
	actions := []string{}
	if onClick != nil {
		actions = []string{"default", "Open"}
	}
	if n.markup {
		body = html.EscapeString(body)
	}
	var id uint32
	err := n.call(n.owner, notificationService+".Notify", "Keel", n.ids[key], "", title, body, actions, map[string]dbus.Variant{}, int32(-1)).Store(&id)
	if err != nil {
		return notificationDBusError(err)
	}
	if id == 0 {
		return notificationDBusError(fmt.Errorf("invalid zero notification ID"))
	}
	n.ids[key] = id
	delete(n.tokens, key)
	if onClick != nil {
		n.clicks[key] = onClick
	} else {
		delete(n.clicks, key)
	}
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
	delete(n.tokens, key)
	delete(n.clicks, key)
	return nil
}
func (n *notificationDBus) closed(sender string, id uint32) {
	if sender != n.owner {
		return
	}
	for key, current := range n.ids {
		if current == id {
			delete(n.ids, key)
			delete(n.tokens, key)
			delete(n.clicks, key)
		}
	}
}

// Return the callback to run outside the connection lock. Consume once while
// retaining the numeric ID until the daemon closes it or the app retracts it.
func (n *notificationDBus) activated(sender string, id uint32, action string) func() {
	if sender != n.owner {
		return nil
	}
	for key, current := range n.ids {
		if current != id {
			continue
		}
		token := n.tokens[key]
		delete(n.tokens, key)
		if action != "default" {
			return nil
		}
		fn := n.clicks[key]
		delete(n.clicks, key)
		if fn == nil {
			return nil
		}
		return func() { fn(token) }
	}
	return nil
}

func (n *notificationDBus) activationToken(sender string, id uint32, token string) {
	if sender != n.owner {
		return
	}
	for key, current := range n.ids {
		if current == id && n.clicks[key] != nil {
			if token == "" {
				delete(n.tokens, key)
			} else {
				n.tokens[key] = token
			}
			return
		}
	}
}

// signal validates the shared D-Bus payload before touching notification state.
// Call under the connection lock, then invoke the returned callback outside it.
func (n *notificationDBus) signal(signal *dbus.Signal) func() {
	if signal == nil || signal.Path != "/org/freedesktop/Notifications" || len(signal.Body) != 2 {
		return nil
	}
	id, ok := signal.Body[0].(uint32)
	if !ok {
		return nil
	}
	switch signal.Name {
	case notificationService + ".NotificationClosed":
		if _, ok := signal.Body[1].(uint32); ok {
			n.closed(signal.Sender, id)
		}
	case notificationService + ".ActivationToken":
		if token, ok := signal.Body[1].(string); ok {
			n.activationToken(signal.Sender, id, token)
		}
	case notificationService + ".ActionInvoked":
		if action, ok := signal.Body[1].(string); ok {
			return n.activated(signal.Sender, id, action)
		}
	}
	return nil
}
