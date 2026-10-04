// Package notification posts local operating-system notifications independently
// of Keel's UI. Implemented for bundled macOS applications with cgo and Linux desktop D-Bus.
package notification

import (
	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
	"strings"
	"unicode/utf8"
)

// Activation carries optional platform data for opening a notification.
// Token is an opaque Linux X11 startup ID or Wayland xdg-activation token.
// It is empty when the server or platform does not supply one.
type Activation struct{ Token string }

// Message is a system notification. ID is an application-wide stable identifier;
// posting it again replaces the matching notification. Title or Body is required.
type Message struct {
	ID, Title, Body string
	// OnClick runs on a separate goroutine when the system notification is opened.
	// Supported on macOS and Linux servers advertising actions. It does not
	// raise a specific UI window.
	OnClick func()
	// OnActivate receives the platform's activation data on the callback goroutine.
	// If both callbacks are set, it runs before OnClick. No window is raised here.
	OnActivate func(Activation)
}

// Available reports whether this process can use the platform implementation.
// It does not report notification permission or guarantee banner presentation.
// On Linux it queries the session bus and can wait for a service response.
func Available() bool { return sys.NotificationAvailable() }

// RequestPermission asks for alert authorization. done runs once on a separate
// goroutine. It may be nil. macOS needs a bundled app and a running event loop.
// Linux only checks the notification service and does not show a permission prompt.
func RequestPermission(done func(error)) { sys.NotificationPermission(completion(done)) }

// Post submits a notification without requesting permission. done reports OS
// acceptance, not that a banner was shown or read. It runs once on a separate
// goroutine. RequestPermission must have completed successfully first.
func Post(m Message, done func(error)) {
	cb := completion(done)
	if !valid(m.ID) || m.ID == "" || !valid(m.Title) || !valid(m.Body) || m.Title == "" && m.Body == "" {
		cb(native.ErrInvalidArgument)
		return
	}
	if m.OnActivate != nil {
		sys.NotificationPostActivated(m.ID, m.Title, m.Body, func(token string) {
			m.OnActivate(Activation{Token: token})
			if m.OnClick != nil {
				m.OnClick()
			}
		}, cb)
	} else if m.OnClick != nil {
		sys.NotificationPostInteractive(m.ID, m.Title, m.Body, m.OnClick, cb)
	} else {
		sys.NotificationPost(m.ID, m.Title, m.Body, cb)
	}
}

// Remove requests removal of both pending and delivered notifications with id.
// Completion confirms the removal calls were issued; macOS provides no removal
// acknowledgment. Removing an unknown ID is harmless. done may be nil.
func Remove(id string, done func(error)) {
	cb := completion(done)
	if id == "" || !valid(id) {
		cb(native.ErrInvalidArgument)
		return
	}
	sys.NotificationRemove(id, cb)
}

func valid(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func completion(done func(error)) func(error) {
	return func(err error) {
		if done != nil {
			go done(err)
		}
	}
}
