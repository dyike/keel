package window

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/dyike/keel/ui/internal/loop"
)

// MainOptions configures application lifetime; the zero value preserves Main.
type MainOptions struct {
	// KeepAlive keeps running when the last window closes, and executes core.Update
	// even with no visible windows. Use Quit for an explicit exit.
	KeepAlive bool
	// OnQuit runs once under the frame lock, after open windows' OnClose callbacks.
	// Use it for application cleanup; Main does not return and defers do not run.
	OnQuit func()
}

var application struct {
	sync.Mutex
	keepAlive atomic.Bool
	quitting  atomic.Bool
	windows   map[*Window]bool
	onQuit    func()
}

func configureApplication(options MainOptions) {
	application.Lock()
	application.onQuit = options.OnQuit
	application.keepAlive.Store(options.KeepAlive)
	application.Unlock()
	loop.SetBackgroundUpdates(options.KeepAlive)
}

func registerApplicationWindow(w *Window) {
	application.Lock()
	if application.windows == nil {
		application.windows = make(map[*Window]bool)
	}
	application.windows[w] = true
	application.Unlock()
}

func unregisterApplicationWindow(w *Window) {
	application.Lock()
	delete(application.windows, w)
	application.Unlock()
}

// Quit requests process exit. It is safe in UI callbacks and without windows.
// Open windows' OnClose and MainOptions.OnQuit run once before exit. It never
// waits for the main thread while holding the frame lock.
func Quit() {
	if !application.quitting.CompareAndSwap(false, true) {
		return
	}
	go func() {
		runApplicationCleanup()
		if automating() {
			os.Remove(auto.addr)
		}
		os.Exit(0)
	}()
}

// Snapshot and callbacks share the frame lock with Open and destruction, so Quit
// cannot skip a window that was still finishing its OnClose callback.
func runApplicationCleanup() {
	loop.Lock()
	defer loop.Unlock()
	application.Lock()
	windows := make([]*Window, 0, len(application.windows))
	for w := range application.windows {
		windows = append(windows, w)
	}
	onQuit := application.onQuit
	application.onQuit = nil
	application.Unlock()
	closeWindowsLocked(windows)
	if onQuit != nil {
		onQuit()
	}
}
