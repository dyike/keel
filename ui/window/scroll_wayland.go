//go:build linux && !android && !nowayland

package window

/*
#cgo pkg-config: wayland-client
#include "scroll_wayland.h"
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/dyike/keel/ui/core"
)

// Wayland scroll gestures: each window's connection gets a pointer of our
// own (scroll_wayland.c) whose queued events are handled when components
// ask core.CurrentScrollGesture.

var wlScroll struct {
	sync.Mutex
	once    sync.Once
	watches map[*Window]*C.struct_keel_wlscroll
	gen     map[*Window]uint64 // the latest request per window
}

// watchWaylandScroll starts watching display for w, replacing an earlier
// watch; a nil display stops it. Gio announces a nil display before it
// destroys the connection, and the stop happens at once, so no watch
// outlives it. Binding takes roundtrips, so it runs off Gio's event path.
func watchWaylandScroll(w *Window, display unsafe.Pointer) {
	wlScroll.once.Do(func() {
		wlScroll.watches = map[*Window]*C.struct_keel_wlscroll{}
		wlScroll.gen = map[*Window]uint64{}
		core.SetScrollGesturePoll(pollWaylandScroll)
	})
	wlScroll.Lock()
	defer wlScroll.Unlock()
	wlScroll.gen[w]++
	gen := wlScroll.gen[w]
	if old := wlScroll.watches[w]; old != nil {
		C.keel_wlscroll_close(old)
		delete(wlScroll.watches, w)
	}
	if display == nil {
		delete(wlScroll.gen, w)
		return
	}
	go func() {
		wlScroll.Lock()
		defer wlScroll.Unlock()
		if wlScroll.gen[w] != gen {
			return // replaced or closed meanwhile
		}
		if s := C.keel_wlscroll_open(display, 0); s != nil {
			wlScroll.watches[w] = s
		}
	}()
}

func pollWaylandScroll() {
	wlScroll.Lock()
	defer wlScroll.Unlock()
	for _, s := range wlScroll.watches {
		C.keel_wlscroll_poll(s)
	}
}

//export keel_wl_scroll_frame
func keel_wl_scroll_frame(handle C.uintptr_t, source C.uint32_t, scrolled, stopped C.int) {
	reportWaylandScroll(uint32(source), scrolled != 0, stopped != 0)
}
