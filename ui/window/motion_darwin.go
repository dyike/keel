//go:build darwin && !ios && cgo

package window

/*
#cgo LDFLAGS: -framework AppKit
void keel_watch_motion(void);
void keel_watch_scroll(void);
void keel_watch_scrollers(void);
int keel_overlay_scrollers(void);
*/
import "C"

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
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
		C.keel_watch_scroll()
		C.keel_watch_scrollers()
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

//export keel_scroll_event
func keel_scroll_event(precise, active, momentum, ended C.int) {
	device := core.ScrollDeviceWheel
	if precise != 0 {
		device = core.ScrollDeviceTrackpad
	}
	core.ReportScrollGesture(device, active != 0, momentum != 0, ended != 0)
}

//export keel_scrollers_changed
func keel_scrollers_changed(overlay C.int) {
	mode := el.ScrollbarAlways
	if overlay != 0 {
		mode = el.ScrollbarScrolling
	}
	el.SetSystemScrollbars(mode)
	go loop.InvalidateAll()
}

func overlayScrollers() bool { return C.keel_overlay_scrollers() != 0 }
