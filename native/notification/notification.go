// Package notification posts local operating-system notifications independently
// of Keel's UI. Currently implemented for bundled macOS applications with cgo.
package notification

import (
	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
	"strings"
	"unicode/utf8"
)

// Message is a system notification. ID is an application-wide stable identifier;
// posting it again replaces the matching notification. Title or Body is required.
type Message struct{ ID, Title, Body string }

// Available reports whether this process can use the platform implementation.
// It does not report notification permission or guarantee banner presentation.
func Available() bool { return sys.NotificationAvailable() }

// RequestPermission asks for alert authorization. done runs once on a separate
// goroutine. It may be nil. Call from a bundled app with its event loop running.
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
	sys.NotificationPost(m.ID, m.Title, m.Body, cb)
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
