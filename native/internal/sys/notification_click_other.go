//go:build (!darwin && !linux && !windows) || ios || android || (darwin && !cgo)

package sys

import "github.com/dyike/keel/native"

func NotificationPostInteractive(id, title, body string, onClick func(), done func(error)) {
	done(native.ErrUnsupported)
}
