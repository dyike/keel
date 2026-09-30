package window

import (
	"os"

	gioapp "gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

// Options configures a window. Width and Height are in dp; zero uses 640×480.
type Options struct {
	Title         string
	Width, Height int
	Content       core.Widget
	// Shortcuts maps accelerators to callbacks while the window has focus, e.g.
	// "mod+," (Cmd on macOS, Ctrl elsewhere), "ctrl+shift+s", "esc".
	Shortcuts map[string]func()
	OnClose   func()
}

type Window struct {
	win       *gioapp.Window
	opts      Options
	shortcuts []shortcut
	root      root
	closed    bool // guarded by the frame lock
}

// Open creates and shows a window. Call it before Main or from any callback.
// It panics on an invalid shortcut, which is a programming error.
func Open(o Options) *Window {
	w := newWindow(o)
	w.win = new(gioapp.Window)
	w.win.Option(gioapp.Title(o.Title), gioapp.Size(unit.Dp(w.opts.Width), unit.Dp(w.opts.Height)))
	loop.Register(w, w.win.Invalidate)
	go w.run()
	return w
}

func newWindow(o Options) *Window {
	if o.Width == 0 {
		o.Width = 640
	}
	if o.Height == 0 {
		o.Height = 480
	}
	return &Window{opts: o, shortcuts: mustParseShortcuts(o.Shortcuts)}
}

// Main runs the platform event loop. The process exits after the last window closes.
func Main() { gioapp.Main() }

// Close closes the window as if the user clicked its close button.
func (w *Window) Close() { w.perform(system.ActionClose) }

// Raise brings the window to the front.
func (w *Window) Raise() { w.perform(system.ActionRaise) }

// perform must not block: callers hold the frame lock, while Gio's Perform
// waits for the main thread, which may be waiting for another window's frame,
// which waits for the frame lock. So it runs after the caller returns.
func (w *Window) perform(a system.Action) { go w.win.Perform(a) }

// Closed reports whether the window has been destroyed. Call it from UI code.
func (w *Window) Closed() bool { return w.closed }

func (w *Window) run() {
	var ops op.Ops
	for {
		switch e := w.win.Event().(type) {
		case gioapp.DestroyEvent:
			w.destroy()
			return
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, e)
			loop.Lock()
			loop.Drain()
			w.layout(gtx)
			loop.Unlock()
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) layout(gtx core.C) {
	w.handleShortcuts(gtx)
	w.root.Layout(gtx, w.opts.Content)
}

func (w *Window) destroy() {
	remaining := loop.Unregister(w)
	loop.Lock()
	w.closed = true
	if w.opts.OnClose != nil {
		w.opts.OnClose()
	}
	loop.Unlock()
	if remaining == 0 {
		os.Exit(0)
	}
}
