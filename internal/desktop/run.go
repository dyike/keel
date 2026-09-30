//go:build (darwin && cgo) || windows || linux

// Package desktop selects the window event loop at build time.
package desktop

import (
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"
)

func Run(app *gui.App, windows ...*gui.Window) error {
	restore, err := preparePlatform(windows)
	if err != nil {
		return err
	}
	defer restore()
	backend.RunApp(app, windows...)
	return nil
}
