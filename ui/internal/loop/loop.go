// Package loop owns the state shared by every window: one lock that serializes
// rendering and callbacks, a queue of updates from other goroutines, and the
// windows to redraw. UI infrastructure uses it; theme only requests redraws.
package loop

import (
	"sync"
	"sync/atomic"
)

var (
	frame sync.Mutex

	mu             sync.Mutex
	pending        []func()
	windows        = map[any]func(){}
	background     bool
	backgroundOnce sync.Once
	wake           = make(chan struct{}, 1)
)

// Lock is held while a window lays out a frame, and so during every UI callback.
func Lock()   { memoryActivity(); frame.Lock() }
func Unlock() { frame.Unlock(); memoryActivity() }

// Register adds a window; invalidate requests a redraw of it.
func Register(key any, invalidate func()) {
	mu.Lock()
	windows[key] = invalidate
	mu.Unlock()
}

// Unregister removes a window and reports how many remain.
func Unregister(key any) int {
	mu.Lock()
	defer mu.Unlock()
	delete(windows, key)
	return len(windows)
}

// generation counts changes made outside a frame: every InvalidateAll (so
// every Post) and every Changed.
var generation atomic.Uint64

// Generation reports the count of changes made outside frames. A root that
// sees the same count, and no input, may repaint its retained content.
func Generation() uint64 { return generation.Load() }

// Changed records a change made outside a frame that a window draws, such
// as its focus or maximized state; the caller invalidates the window.
func Changed() { generation.Add(1) }

// InvalidateAll requests a redraw of every window. It never blocks on them.
func InvalidateAll() {
	generation.Add(1)
	mu.Lock()
	defer mu.Unlock()
	for _, fn := range windows {
		fn()
	}
}

// Post queues fn, wakes every window and signals the optional resident worker.
func Post(fn func()) {
	memoryActivity()
	mu.Lock()
	pending = append(pending, fn)
	mu.Unlock()
	InvalidateAll()
	wakeBackground()
}

// PostIf runs a background check under the frame lock. In resident mode it
// redraws only when fn reports a visible change. Without resident mode, a
// frame is still needed to run the check, just like Post.
func PostIf(fn func() bool) {
	mu.Lock()
	pending = append(pending, func() {
		if fn() {
			memoryActivity()
			InvalidateAll()
		}
	})
	resident := background
	mu.Unlock()
	if !resident {
		InvalidateAll()
	}
	wakeBackground()
}

// Pending reports whether updates are queued. It does not take the frame lock.
func Pending() bool {
	mu.Lock()
	defer mu.Unlock()
	return len(pending) > 0
}

// Drain runs queued functions. Call it with Lock held.
func Drain() {
	mu.Lock()
	fns := pending
	pending = nil
	mu.Unlock()
	for _, fn := range fns {
		fn()
	}
}

// SetBackgroundUpdates lets updates run without a frame, under the same lock.
// Only resident applications enable it. The worker sleeps until an explicit post.
func SetBackgroundUpdates(enabled bool) {
	mu.Lock()
	background = enabled
	mu.Unlock()
	if enabled {
		backgroundOnce.Do(func() { go runBackgroundUpdates() })
		wakeBackground()
	}
}

func wakeBackground() {
	select {
	case wake <- struct{}{}:
	default:
	}
}

func runBackgroundUpdates() {
	for range wake {
		frame.Lock()
		mu.Lock()
		enabled := background
		mu.Unlock()
		if enabled {
			Drain()
		}
		frame.Unlock()
	}
}
