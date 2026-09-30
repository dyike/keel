//go:build (!darwin || !cgo) && !windows && !linux

package desktop

import (
	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
)

func Run(*gui.App, ...*gui.Window) error { return capability.ErrUnsupported }
