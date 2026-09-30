// Package permission checks and requests macOS privacy permissions.
package permission

import (
	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

type Kind int

const (
	Accessibility   Kind = iota // needed by native/input
	ScreenRecording             // needed by screen.Capture
	InputMonitoring
)

// Granted reports whether k is granted, without prompting.
func Granted(k Kind) (bool, error) { return check(k, false) }

// Request may show a system prompt; the result is the grant at return time,
// not the user's eventual answer.
func Request(k Kind) (bool, error) { return check(k, true) }

func check(k Kind, request bool) (bool, error) {
	if k < Accessibility || k > InputMonitoring {
		return false, native.ErrInvalidArgument
	}
	return sys.Permission(int(k), request)
}
