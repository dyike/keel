// Package keel combines Go-Gui desktop windows with native platform capabilities.
package keel

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/appinfo"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/input"
	"github.com/dyike/keel/internal/desktop"
	"github.com/dyike/keel/permissions"
	"github.com/dyike/keel/screen"
	"github.com/dyike/keel/shortcut"
	"github.com/dyike/keel/window"
)

// Keep the program's main goroutine on the OS main thread for the desktop event loop.
func init() { runtime.LockOSThread() }

type Options struct {
	Name, ID string
	// ExitOnMainClose exits when the first window closes. By default all windows
	// must close; a configured tray can keep the Go-Gui application alive.
	ExitOnMainClose bool
}
type WindowOptions = window.Options

type Kit struct {
	mu          sync.Mutex
	shutdownErr error
	App         *gui.App
	Window      *window.Manager
	Input       *input.Controller
	Permissions *permissions.Manager
	Screen      *screen.Manager
	Shortcut    *shortcut.Manager
}

func New() *Kit { return NewWithOptions(Options{Name: "Keel"}) }
func NewWithOptions(o Options) *Kit {
	app := gui.NewApp()
	if !o.ExitOnMainClose {
		app.ExitMode = gui.ExitOnTrayRemoved
	}
	kit, err := Attach(app)
	if err != nil {
		panic(err)
	}
	kit.Window.SetAppInfo(appinfo.Info{Name: o.Name, ID: o.ID})
	return kit
}

// Attach creates Keel managers for an existing Go-Gui App. Use Kit.Window to create
// its startup windows and Kit.Run to own the event loop and cleanup.
func Attach(app *gui.App) (*Kit, error) {
	if app == nil {
		return nil, capability.ErrInvalidArgument
	}
	windows, err := window.NewManager(app)
	if err != nil {
		return nil, err
	}
	return &Kit{App: app, Window: windows, Input: input.New(), Permissions: permissions.New(), Screen: screen.New(), Shortcut: shortcut.New()}, nil
}

// Run starts the native desktop event loop. Call once, from main, after creating
// at least one window. Runtime windows can be opened from Go callbacks or workers.
func (k *Kit) Run() (err error) {
	windows, err := k.Window.BeginRun()
	if err != nil {
		return err
	}
	defer func() {
		if value := recover(); value != nil {
			err = errors.Join(err, fmt.Errorf("keel: GUI backend failed: %v", value))
		}
		k.Window.EndRun()
		closeErr := k.Close()
		k.mu.Lock()
		err = errors.Join(err, closeErr, k.shutdownErr)
		k.mu.Unlock()
	}()
	return desktop.Run(k.App, windows...)
}
func (k *Kit) Close() error { k.Window.Shutdown(); return k.Shortcut.Close() }

// Quit schedules native cleanup and closes the application on the UI thread.
func (k *Kit) Quit() error {
	return k.Window.Quit(func() {
		err := k.Close()
		k.mu.Lock()
		k.shutdownErr = errors.Join(k.shutdownErr, err)
		k.mu.Unlock()
	})
}

// Dispatch runs a Go callback on the UI thread. Global shortcut callbacks run
// on workers: dispatch operations that read or change application UI state.
func (k *Kit) Dispatch(fn func()) error { return k.Window.Dispatch(fn) }
