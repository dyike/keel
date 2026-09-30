package window

import (
	"image"
	"os"
	"time"

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
	win       *gioapp.Window // nil in automation mode
	virt      *virtual       // non-nil in automation mode
	opts      Options
	shortcuts []shortcut
	root      root
	closed    bool          // guarded by the frame lock
	shown     chan struct{} // closed at the first frame or at destruction
}

// Open creates and shows a window. Call it before Main or from any callback.
// It panics on an invalid shortcut, which is a programming error.
func Open(o Options) *Window {
	w := newWindow(o)
	if offScreen() {
		openVirtual(w, true)
		return w
	}
	w.win = new(gioapp.Window)
	w.win.Option(gioapp.Title(o.Title), gioapp.Size(unit.Dp(w.opts.Width), unit.Dp(w.opts.Height)))
	loop.Register(w, w.win.Invalidate)
	if automating() {
		openVirtual(w, false) // shadow of the real window, driven by agents
	}
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
	return &Window{opts: o, shortcuts: mustParseShortcuts(o.Shortcuts), shown: make(chan struct{})}
}

// Main runs the platform event loop. The process exits after the last window
// closes. With KEEL_AUTOMATION set it also serves automation requests, or only
// those with KEEL_HEADLESS=1; see automation.go.
func Main() {
	if offScreen() {
		serveAutomation()
		return
	}
	if automating() {
		go serveAutomation()
	}
	gioapp.Main()
}

// Close closes the window as if the user clicked its close button.
func (w *Window) Close() { w.perform(system.ActionClose) }

// Raise brings the window to the front.
func (w *Window) Raise() { w.perform(system.ActionRaise) }

// perform must not block: callers hold the frame lock, while Gio's Perform
// waits for the main thread, which may be waiting for another window's frame,
// which waits for the frame lock. So it runs after the caller returns.
//
// It also waits for the window's first frame: Gio v0.10.3 on macOS crashes if
// a window is closed before its native window is fully created (it cascades
// the position of the already released NSWindow). An agent easily closes a
// window within milliseconds of opening it.
func (w *Window) perform(a system.Action) {
	if w.win == nil { // headless automation
		queueAction(w, a)
		return
	}
	go func() {
		<-w.shown
		if !w.isClosed() {
			w.win.Perform(a)
		}
	}()
}

func (w *Window) isClosed() bool {
	loop.Lock()
	defer loop.Unlock()
	return w.closed
}

// Closed reports whether the window has been destroyed. Call it from UI code.
func (w *Window) Closed() bool { return w.closed }

func (w *Window) run() {
	var ops op.Ops
	for {
		switch e := w.win.Event().(type) {
		case gioapp.DestroyEvent:
			w.markShown()
			w.destroy()
			return
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, e)
			if w.virt != nil { // keep the shadow the same size as the window
				w.virt.setSize(image.Pt(int(e.Metric.PxToDp(e.Size.X)), int(e.Metric.PxToDp(e.Size.Y))))
			}
			loop.Lock()
			loop.Drain()
			w.layout(gtx)
			loop.Unlock()
			e.Frame(gtx.Ops)
			w.markShown()
		}
	}
}

func (w *Window) layout(gtx core.C) {
	w.handleShortcuts(gtx)
	w.root.Layout(gtx, w.opts.Content)
}

func (w *Window) destroy() {
	remaining := w.finish()
	if w.virt != nil {
		forgetVirtual(w)
	}
	if remaining == 0 {
		if automating() {
			// Give the automation server time to answer the request that
			// closed the window, and AppKit time to finish tearing it down.
			time.AfterFunc(500*time.Millisecond, func() { os.Remove(auto.addr); os.Exit(0) })
			return
		}
		os.Exit(0)
	}
}

// finish marks the window closed, runs OnClose and reports how many windows remain.
func (w *Window) finish() int {
	remaining := loop.Unregister(w)
	loop.Lock()
	w.closed = true
	if w.opts.OnClose != nil {
		w.opts.OnClose()
	}
	loop.Unlock()
	return remaining
}

func (w *Window) markShown() {
	select {
	case <-w.shown:
	default:
		close(w.shown)
	}
}
