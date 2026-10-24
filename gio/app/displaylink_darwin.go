// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/gio/io/pointer"
)

// The display link and cursor logic shared by the darwin bindings
// (os_darwin.go with cgo, os_darwin_purego.go without). Keel patch: moved
// from os_darwin.go, over plain handles.

// displayLink is the state for a display link (CVDisplayLinkRef on macOS,
// CADisplayLink on iOS). It runs a state-machine goroutine that keeps the
// display link running for a while after being stopped to avoid the thread
// start/stop overhead and because the CVDisplayLink sometimes fails to
// start, stop and start again within a short duration.
type displayLink struct {
	callback func()
	// states is for starting or stopping the display link.
	states chan bool
	// done is closed when the display link is destroyed.
	done chan struct{}
	// dids receives the display id when the callback owner is moved
	// to a different screen.
	dids chan uint64
	// running tracks the desired state of the link. running is accessed
	// with atomic.
	running uint32
	// requested is the state last asked for with Start or Stop, accessed
	// with atomic (Keel patch).
	requested uint32
}

// displayLinks maps display link handles (uintptr) to *displayLinks.
var displayLinks sync.Map

func newDisplayLink(callback func()) (*displayLink, error) {
	d := &displayLink{
		callback: callback,
		done:     make(chan struct{}),
		states:   make(chan bool),
		dids:     make(chan uint64),
	}
	dl := createDisplayLink()
	if dl == 0 {
		return nil, errors.New("app: failed to create display link")
	}
	go d.run(dl)
	return d, nil
}

func (d *displayLink) run(dl uintptr) {
	defer releaseDisplayLink(dl)
	displayLinks.Store(dl, d)
	defer displayLinks.Delete(dl)
	var stopTimer *time.Timer
	var tchan <-chan time.Time
	started := false
	for {
		select {
		case <-tchan:
			tchan = nil
			started = false
			stopDisplayLink(dl)
		case start := <-d.states:
			switch {
			case !start && tchan == nil:
				// stopTimeout is the delay before stopping the display link to
				// avoid the overhead of frequently starting and stopping the
				// link thread.
				const stopTimeout = 500 * time.Millisecond
				if stopTimer == nil {
					stopTimer = time.NewTimer(stopTimeout)
				} else {
					// stopTimer is always drained when tchan == nil.
					stopTimer.Reset(stopTimeout)
				}
				tchan = stopTimer.C
				atomic.StoreUint32(&d.running, 0)
			case start:
				if tchan != nil && !stopTimer.Stop() {
					<-tchan
				}
				tchan = nil
				atomic.StoreUint32(&d.running, 1)
				if !started {
					started = true
					startDisplayLink(dl)
				}
			}
		case did := <-d.dids:
			setDisplayLinkDisplay(dl, did)
		case <-d.done:
			return
		}
	}
}

// Start runs the display link. Keel patch: only a change of the requested
// state is sent. macOS windows call Start on every animated frame, and the
// send blocked the main thread on a goroutine handoff each time.
func (d *displayLink) Start() {
	if atomic.SwapUint32(&d.requested, 1) == 1 {
		return
	}
	d.states <- true
}

// Stop stops the display link after a delay; see Start.
func (d *displayLink) Stop() {
	if atomic.SwapUint32(&d.requested, 0) == 0 {
		return
	}
	d.states <- false
}

func (d *displayLink) Close() {
	close(d.done)
}

func (d *displayLink) SetDisplayID(did uint64) {
	d.dids <- did
}

// onFrameCallback runs on the display link's thread.
func onFrameCallback(ref uintptr) {
	d, exists := displayLinks.Load(ref)
	if !exists {
		return
	}
	dl := d.(*displayLink)
	if atomic.LoadUint32(&dl.running) != 0 {
		dl.callback()
	}
}

var macosCursorID = [...]byte{
	pointer.CursorDefault:                  0,
	pointer.CursorNone:                     1,
	pointer.CursorText:                     2,
	pointer.CursorVerticalText:             3,
	pointer.CursorPointer:                  4,
	pointer.CursorCrosshair:                5,
	pointer.CursorAllScroll:                6,
	pointer.CursorColResize:                7,
	pointer.CursorRowResize:                8,
	pointer.CursorGrab:                     9,
	pointer.CursorGrabbing:                 10,
	pointer.CursorNotAllowed:               11,
	pointer.CursorWait:                     12,
	pointer.CursorProgress:                 13,
	pointer.CursorNorthWestResize:          14,
	pointer.CursorNorthEastResize:          15,
	pointer.CursorSouthWestResize:          16,
	pointer.CursorSouthEastResize:          17,
	pointer.CursorNorthSouthResize:         18,
	pointer.CursorEastWestResize:           19,
	pointer.CursorWestResize:               20,
	pointer.CursorEastResize:               21,
	pointer.CursorNorthResize:              22,
	pointer.CursorSouthResize:              23,
	pointer.CursorNorthEastSouthWestResize: 24,
	pointer.CursorNorthWestSouthEastResize: 25,
}

// windowSetCursor updates the cursor from the current one to a new one
// and returns the new one.
func windowSetCursor(from, to pointer.Cursor) pointer.Cursor {
	if from == to {
		return to
	}
	if to == pointer.CursorNone {
		hideCursor()
		return to
	}
	if from == pointer.CursorNone {
		showCursor()
	}
	setCursor(uint(macosCursorID[to]))
	return to
}

func (w *window) wakeup() {
	runOnMain(func() {
		w.loop.Wakeup()
		w.loop.FlushEvents()
	})
}
