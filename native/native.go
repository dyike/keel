// Package native holds the errors shared by Keel's platform capabilities. The
// capabilities themselves live in subpackages:
//
//	native/permission  check and request system permissions
//	native/screen      list displays, capture screenshots
//	native/input       synthesize mouse and keyboard input
//	native/hotkey      system-wide shortcuts
//
//	native/notification local notifications (currently macOS app bundles only)
//
// Implemented on macOS 14+ (cgo), Windows and Linux under X11; elsewhere, and
// in macOS builds without cgo, every call returns ErrUnsupported. To add a
// capability, add it to each platform file in internal/sys (plus a stub in
// sys_other.go) and a small public package that validates arguments.
package native

import "errors"

var (
	ErrUnsupported      = errors.New("native: unsupported on this platform")
	ErrPermissionDenied = errors.New("native: permission not granted")
	ErrInvalidArgument  = errors.New("native: invalid argument")
	ErrTimeout          = errors.New("native: operation timed out")
	ErrConflict         = errors.New("native: already registered")
	ErrFailed           = errors.New("native: operation failed")
)
