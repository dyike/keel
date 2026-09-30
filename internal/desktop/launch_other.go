//go:build windows || linux

package desktop

import "github.com/go-gui-org/go-gui/gui"

func preparePlatform([]*gui.Window) (func(), error) { return func() {}, nil }
