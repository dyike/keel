//go:build darwin && cgo

package desktop

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit
int keelApplicationPrepare(void);
void keelApplicationFinishLaunch(void);
void keelApplicationCleanup(void);
*/
import "C"

import (
	"fmt"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
)

func preparePlatform(windows []*gui.Window) (func(), error) {
	if C.keelApplicationPrepare() == 0 {
		return nil, fmt.Errorf("keel: another library initialized NSApplication before Kit.Run: %w", capability.ErrConflict)
	}
	if len(windows) == 0 {
		return func() { C.keelApplicationCleanup() }, nil
	}
	w := windows[0]
	init := w.Config.OnInit
	w.Config.OnInit = func(host *gui.Window) {
		C.keelApplicationFinishLaunch()
		if init != nil {
			init(host)
		}
	}
	return func() { w.Config.OnInit = init; C.keelApplicationCleanup() }, nil
}
