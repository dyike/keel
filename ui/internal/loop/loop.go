// Package loop owns the state shared by every window: one lock that serializes
// rendering and callbacks, a queue of updates from other goroutines, and the
// windows to redraw. Only ui/core and ui/window use it.
package loop

import "sync"

var (
	frame sync.Mutex

	mu      sync.Mutex
	pending []func()
	windows = map[any]func(){}
)

// Lock is held while a window lays out a frame, and so during every UI callback.
func Lock()   { frame.Lock() }
func Unlock() { frame.Unlock() }

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

// InvalidateAll requests a redraw of every window. It never blocks on them.
func InvalidateAll() {
	mu.Lock()
	defer mu.Unlock()
	for _, fn := range windows {
		fn()
	}
}

// Post queues fn for the next frame and wakes every window.
func Post(fn func()) {
	mu.Lock()
	pending = append(pending, fn)
	mu.Unlock()
	InvalidateAll()
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
