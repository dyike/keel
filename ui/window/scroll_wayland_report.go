package window

import (
	"sync/atomic"

	"github.com/dyike/keel/ui/core"
)

// Wayland axis sources (wl_pointer.axis_source).
const (
	wlAxisWheel      = 0
	wlAxisFinger     = 1
	wlAxisContinuous = 2
	wlAxisWheelTilt  = 3
	wlAxisNone       = 0xffffffff // the frame named no source
)

// fingersDown tracks a finger scroll between frames. It is only touched
// while polling, under the watch lock, but reports run there too, so it
// must not call back into core.CurrentScrollGesture.
var fingersDown atomic.Bool

// reportWaylandScroll turns one wl_pointer frame into core's gesture state.
// Fingers report phases (axis_stop is the lift, possibly in a frame of its
// own without a source); a wheel steps; continuous sources say nothing.
func reportWaylandScroll(source uint32, scrolled, stopped bool) {
	switch {
	case source == wlAxisFinger || source == wlAxisNone && stopped && fingersDown.Load():
		fingersDown.Store(scrolled && !stopped)
		core.ReportScrollGesture(core.ScrollDeviceTrackpad, scrolled && !stopped, false, stopped)
	case source == wlAxisWheel || source == wlAxisWheelTilt:
		fingersDown.Store(false)
		core.ReportScrollGesture(core.ScrollDeviceWheel, false, false, false)
	}
}
