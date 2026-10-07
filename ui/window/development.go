package window

import (
	"os"
	"os/signal"
	"sync"

	"github.com/dyike/keel/ui/internal/loop"
)

var development struct {
	sync.Mutex
	once    sync.Once
	windows map[*Window]bool
}

// Only keel run's watched window processes participate. Background helpers
// that inherit the environment but never open a window keep their own lifetime.
func registerDevelopmentWindow(w *Window) {
	if os.Getenv("KEEL_RUN_WATCH") != "1" {
		return
	}
	development.Lock()
	if development.windows == nil {
		development.windows = make(map[*Window]bool)
	}
	development.windows[w] = true
	development.Unlock()
	development.once.Do(func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		go func() {
			<-signals
			development.Lock()
			windows := make([]*Window, 0, len(development.windows))
			for w := range development.windows {
				windows = append(windows, w)
			}
			development.Unlock()
			closeDevelopmentWindows(windows)
			os.Exit(0)
		}()
	})
}

func unregisterDevelopmentWindow(w *Window) {
	development.Lock()
	delete(development.windows, w)
	development.Unlock()
}

// Run the same persistence/disposal callbacks as window closure, while UI state
// is serialized. Exit follows immediately; native teardown need not render a
// frame, which also permits reloading while the app is inactive or minimized.
func closeDevelopmentWindows(windows []*Window) {
	loop.Lock()
	defer loop.Unlock()
	for _, w := range windows {
		if w.closed {
			continue
		}
		w.closed = true
		if callback := w.opts.OnClose; callback != nil {
			w.opts.OnClose = nil
			callback()
		}
	}
}
