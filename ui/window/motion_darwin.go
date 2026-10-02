//go:build darwin && !ios && cgo

package window

/*
#cgo LDFLAGS: -framework AppKit
void keel_watch_motion(void);
*/
import "C"

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"sync"
)

var motionOnce sync.Once
var motionUpdates = make(chan bool, 1)

func watchSystemPreferences() {
	motionOnce.Do(func() {
		go func() {
			for reduce := range motionUpdates {
				core.Update(func() { theme.SetSystemReducedMotion(reduce) })
			}
		}()
		C.keel_watch_motion()
	})
}

//export keel_motion_changed
func keel_motion_changed(reduce C.int) {
	// AppKit callbacks must not block on the UI frame lock. Preserve the newest
	// value while a frame is busy, with one consumer applying updates in order.
	select {
	case motionUpdates <- reduce != 0:
	default:
		select {
		case <-motionUpdates:
		default:
		}
		motionUpdates <- reduce != 0
	}
}
