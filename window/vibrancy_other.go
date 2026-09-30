//go:build !darwin || !cgo

package window

import "github.com/dyike/keel/capability"

func setGUIVibrancy(*guiDriver, Vibrancy) error { return capability.ErrUnsupported }
