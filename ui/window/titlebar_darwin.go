//go:build darwin && !ios && cgo

package window

/*
#include <stdint.h>
void keel_native_traffic_lights(uintptr_t view, int custom, double height, double left, double offsetY, double spacing);
void keel_titlebar_area(uintptr_t view, double x, double y, double width, double height);
*/
import "C"

import (
	gioapp "gioui.org/app"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"sync"
)

var nativeWindows sync.Map // uintptr NSView -> *Window; never pass Go pointers to C
func platformWindowEvent(w *Window, event any) {
	if e, ok := event.(gioapp.AppKitViewEvent); ok {
		loop.Lock()
		defer loop.Unlock()
		clearPlatformWindow(w)
		w.nativeView = e.View
		w.nativeTrafficLightsDirty = true
		if e.View != 0 {
			nativeWindows.Store(e.View, w)
		}
	}
}
func clearPlatformWindow(w *Window) {
	if w.nativeView != 0 {
		nativeWindows.Delete(w.nativeView)
		C.keel_titlebar_area(C.uintptr_t(w.nativeView), 0, 0, 0, 0)
		w.nativeView, w.lastTitleView = 0, 0
	}
}
func syncTitleBar(w *Window) {
	if w.nativeView != 0 && w.opts.Frameless && w.opts.NativeTrafficLights && w.nativeTrafficLightsDirty {
		w.nativeTrafficLightsDirty = false
		var custom C.int
		var layout TrafficLightLayout
		if w.opts.TrafficLightLayout != nil {
			custom = 1
			layout = *w.opts.TrafficLightLayout
		}
		C.keel_native_traffic_lights(C.uintptr_t(w.nativeView), custom, C.double(layout.Height), C.double(layout.Left), C.double(layout.OffsetY), C.double(layout.Spacing))
	}
	if w.nativeView == 0 || w.nativeView == w.lastTitleView && w.titleArea == w.lastTitleArea {
		return
	}
	w.lastTitleView, w.lastTitleArea = w.nativeView, w.titleArea
	a := w.titleArea
	C.keel_titlebar_area(C.uintptr_t(w.nativeView), C.double(a[0]), C.double(a[1]), C.double(a[2]), C.double(a[3]))
}

//export keel_titlebar_double_click
func keel_titlebar_double_click(view C.uintptr_t, action C.int) {
	postTitleBarAction(uintptr(view), int(action))
}

func postTitleBarAction(view uintptr, action int) {
	if value, ok := nativeWindows.Load(view); ok {
		w := value.(*Window)
		// AppKit invokes this on the main thread. Gio's Invalidate wakes and
		// flushes events inline there, re-entering its own invalidation mutex.
		// Post from another goroutine so the native callback returns first.
		go core.Update(func() {
			if w.closed {
				return
			}
			if action == 1 {
				w.Minimize()
			} else if action == 2 {
				w.ToggleMaximize()
			}
		})
	}
}
