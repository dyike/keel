package core

import (
	"sync"

	"github.com/dyike/keel/ui/internal/loop"
)

// ScrollDevice is the kind of device behind scroll input.
type ScrollDevice uint8

const (
	// ScrollDeviceUnknown: the platform does not say.
	ScrollDeviceUnknown ScrollDevice = iota
	// ScrollDeviceWheel scrolls in notches.
	ScrollDeviceWheel
	// ScrollDeviceTrackpad scrolls continuously, with gesture phases.
	ScrollDeviceTrackpad
)

// ScrollGesture is what the platform reports about scroll input beyond the
// deltas Gio delivers. macOS and Wayland report it; elsewhere Phases is
// false and components fall back to timing.
type ScrollGesture struct {
	Device ScrollDevice
	// Phases is true once the platform has reported gesture phases, so
	// Active and Ended can be trusted.
	Phases bool
	// Active is true while fingers are on the trackpad.
	Active bool
	// Momentum is true while inertial scrolling continues after a lift.
	Momentum bool
	// Ended counts finished gestures; a change means the fingers lifted.
	Ended uint64
}

var scrollGesture struct {
	sync.Mutex
	g    ScrollGesture
	poll func()
}

// SetScrollGesturePoll installs a function that brings the state up to date
// from queued platform events; window backends without push notification
// (Wayland) use it. CurrentScrollGesture calls it first.
func SetScrollGesturePoll(fn func()) {
	scrollGesture.Lock()
	scrollGesture.poll = fn
	scrollGesture.Unlock()
}

// CurrentScrollGesture returns the latest platform scroll state. Components
// waiting on a gesture's end should re-check it every frame or so: on some
// platforms the end is only seen when polled.
func CurrentScrollGesture() ScrollGesture {
	scrollGesture.Lock()
	poll := scrollGesture.poll
	scrollGesture.Unlock()
	if poll != nil {
		poll()
	}
	scrollGesture.Lock()
	defer scrollGesture.Unlock()
	return scrollGesture.g
}

// ReportScrollGesture records platform scroll state; window backends call it
// for each native scroll event; ScrollDeviceUnknown clears what earlier
// reports established. A finished gesture redraws every window so
// components waiting on it can settle. Safe from any goroutine.
func ReportScrollGesture(device ScrollDevice, active, momentum, ended bool) {
	scrollGesture.Lock()
	g := &scrollGesture.g
	g.Device, g.Phases, g.Active, g.Momentum = device, device == ScrollDeviceTrackpad || g.Phases && device != ScrollDeviceUnknown, active, momentum
	if ended {
		g.Ended++
	}
	scrollGesture.Unlock()
	if ended {
		go loop.InvalidateAll()
	}
}
