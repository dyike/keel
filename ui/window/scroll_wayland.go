//go:build linux && !android && !nowayland

package window

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/wayland"
)

// Wayland scroll gestures: each window's connection gets a wl_pointer of our
// own, on a private queue, to learn what Gio drops: whether scrolling comes
// from a wheel or fingers (axis_source) and when fingers lift (axis_stop).
// Needs wl_seat version 5. Its queued events are handled when components ask
// core.CurrentScrollGesture.

var wlScroll struct {
	sync.Mutex
	once    sync.Once
	watches map[*Window]*wlScrollWatch
	gen     map[*Window]uint64 // the latest request per window
}

// Request and event opcodes from wayland.xml.
const (
	seatGetPointer       = 0
	seatRelease          = 3
	seatCapabilities     = 0 // event
	seatCapabilityPtr    = 1
	pointerRelease       = 1
	pointerAxis          = 4 // events
	pointerFrame         = 5
	pointerAxisSource    = 6
	pointerAxisStop      = 7
	wlScrollKindRegistry = 1
	wlScrollKindSeat     = 2
	wlScrollKindPointer  = 3
)

type wlScrollWatch struct {
	display, queue, wrapper, registry, seat, pointer, token uintptr
	seatName                                                []byte
	source                                                  uint32 // axis_source of the current frame
	scrolled, stopped                                       bool
}

// event runs while the watch's queue is dispatched, under wlScroll's lock.
func (s *wlScrollWatch) event(kind, proxy uintptr, opcode uint32, args uintptr) {
	switch {
	case kind == wlScrollKindRegistry && opcode == wayland.RegistryGlobal:
		if s.seat == 0 && wayland.GoString(wayland.Arg(args, 1)) == "wl_seat" && uint32(wayland.Arg(args, 2)) >= 5 {
			s.seat = wayland.Marshal(proxy, wayland.RegistryBind, wayland.Interface("wl_seat_interface"), 5, 0, wayland.Arg(args, 0), wayland.Ptr(s.seatName), 5, 0)
			if s.seat != 0 {
				wayland.Listen(s.seat, s.token, wlScrollKindSeat)
			}
		}
	case kind == wlScrollKindSeat && opcode == seatCapabilities:
		if uint32(wayland.Arg(args, 0))&seatCapabilityPtr != 0 && s.pointer == 0 {
			s.pointer = wayland.Marshal(proxy, seatGetPointer, wayland.Interface("wl_pointer_interface"), wayland.Version(proxy), 0, 0)
			if s.pointer != 0 {
				wayland.Listen(s.pointer, s.token, wlScrollKindPointer)
			}
		}
	case kind == wlScrollKindPointer:
		switch opcode {
		case pointerAxis:
			s.scrolled = true
		case pointerAxisSource:
			s.source = uint32(wayland.Arg(args, 0))
		case pointerAxisStop:
			s.stopped = true
		case pointerFrame:
			if s.scrolled || s.stopped {
				reportWaylandScroll(s.source, s.scrolled, s.stopped)
			}
			s.source, s.scrolled, s.stopped = wlAxisNone, false, false
		}
	}
}

// openWaylandScroll binds a seat and its pointer on display. Events queue up
// as Gio reads the socket; poll handles them.
func openWaylandScroll(display uintptr) *wlScrollWatch {
	if !wayland.Load() {
		return nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	s := &wlScrollWatch{display: display, source: wlAxisNone, seatName: wayland.CString("wl_seat")}
	s.token = wayland.Register(s.event)
	if s.queue = wayland.CreateQueue(display); s.queue == 0 {
		s.close()
		return nil
	}
	if s.wrapper = wayland.Wrapper(display, s.queue); s.wrapper == 0 {
		s.close()
		return nil
	}
	if s.registry = wayland.Marshal(s.wrapper, wayland.DisplayGetRegistry, wayland.Interface("wl_registry_interface"), wayland.Version(s.wrapper), 0, 0); s.registry == 0 {
		s.close()
		return nil
	}
	wayland.Listen(s.registry, s.token, wlScrollKindRegistry)
	// One roundtrip finds the seat, the next its capabilities.
	if !wayland.Roundtrip(display, s.queue) || s.seat == 0 || !wayland.Roundtrip(display, s.queue) || s.pointer == 0 {
		s.close()
		return nil
	}
	return s
}

func (s *wlScrollWatch) poll() { wayland.DispatchPending(s.display, s.queue) }

func (s *wlScrollWatch) close() {
	if s.pointer != 0 {
		wayland.Marshal(s.pointer, pointerRelease, 0, wayland.Version(s.pointer), wayland.MarshalFlagDestroy)
	}
	if s.seat != 0 {
		wayland.Marshal(s.seat, seatRelease, 0, wayland.Version(s.seat), wayland.MarshalFlagDestroy)
	}
	if s.registry != 0 {
		wayland.Destroy(s.registry)
	}
	if s.wrapper != 0 {
		wayland.DestroyWrapper(s.wrapper)
	}
	wayland.Flush(s.display)
	if s.queue != 0 {
		wayland.DestroyQueue(s.queue)
	}
	wayland.Unregister(s.token)
	s.pointer, s.seat, s.registry, s.wrapper, s.queue = 0, 0, 0, 0, 0
}

// watchWaylandScroll starts watching display for w, replacing an earlier
// watch; a nil display stops it. Gio announces a nil display before it
// destroys the connection, and the stop happens at once, so no watch
// outlives it. Binding takes roundtrips, so it runs off Gio's event path.
func watchWaylandScroll(w *Window, display unsafe.Pointer) {
	wlScroll.once.Do(func() {
		wlScroll.watches = map[*Window]*wlScrollWatch{}
		wlScroll.gen = map[*Window]uint64{}
		core.SetScrollGesturePoll(pollWaylandScroll)
	})
	wlScroll.Lock()
	defer wlScroll.Unlock()
	wlScroll.gen[w]++
	gen := wlScroll.gen[w]
	if old := wlScroll.watches[w]; old != nil {
		old.close()
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
		if s := openWaylandScroll(uintptr(display)); s != nil {
			wlScroll.watches[w] = s
		}
	}()
}

func pollWaylandScroll() {
	wlScroll.Lock()
	defer wlScroll.Unlock()
	for _, s := range wlScroll.watches {
		s.poll()
	}
}
