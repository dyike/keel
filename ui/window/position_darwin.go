//go:build darwin && !ios && cgo

package window

/*
#include <stdint.h>
void keel_center_window(uintptr_t view);
*/
import "C"

// Called once after the first frame: Gio has finished configuring and
// cascading the native window by then. AppKit work is queued, never awaited
// under the frame lock.
func centerNewWindow(w *Window) {
	if w.nativeView != 0 {
		C.keel_center_window(C.uintptr_t(w.nativeView))
	}
}
