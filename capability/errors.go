// Package capability defines errors shared by all Keel APIs.
package capability

import "errors"

var (
	ErrUnsupported      = errors.New("keel: capability unsupported")
	ErrPermissionDenied = errors.New("keel: permission not granted")
	ErrInvalidArgument  = errors.New("keel: invalid argument")
	ErrNotReady         = errors.New("keel: native resource not ready")
	ErrClosed           = errors.New("keel: resource closed")
	ErrConflict         = errors.New("keel: shortcut already registered")
	ErrNative           = errors.New("keel: native operation failed")
	ErrTimeout          = errors.New("keel: operation timed out")
)
