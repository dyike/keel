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
	"sync"
	"sync/atomic"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

var motionOnce sync.Once
var motionUpdates = make(chan bool, 1)
var scrollersAutoHide atomic.Bool

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
	// Gio's Invalidate can synchronously re-enter its event loop on AppKit's
	// main thread while holding invMu. Post from another goroutine instead,
	// and read the latest preference when applying possibly queued updates.
	scrollersAutoHide.Store(overlay != 0)
	go core.Update(func() { theme.SetSystemScrollbarsAutoHide(scrollersAutoHide.Load()) })
}

func overlayScrollers() bool { return C.keel_overlay_scrollers() != 0 }
