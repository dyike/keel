// Package loop owns the state shared by every window: one lock that serializes
// rendering and callbacks, a queue of updates from other goroutines, and the
// windows to redraw. UI infrastructure uses it; theme only requests redraws.
package loop

import "sync"

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

// InvalidateAll requests a redraw of every window. It never blocks on them.
func InvalidateAll() {
	mu.Lock()
	defer mu.Unlock()
	for _, fn := range windows {
		fn()
	}
}

// Post queues fn, wakes every window and signals the optional resident worker.
func Post(fn func()) {
	mu.Lock()
	pending = append(pending, fn)
	mu.Unlock()
	InvalidateAll()
	wakeBackground()
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
		Lock()
		mu.Lock()
		enabled := background
		mu.Unlock()
		if enabled {
			Drain()
		}
		Unlock()
	}
}
