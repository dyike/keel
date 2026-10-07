package window

import (
	"image"
	"os"
	"sync/atomic"
	"time"
	"unsafe"

	gioapp "gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

// TrafficLightLayout positions macOS's standard window buttons in dp.
// Buttons retain their system size, rendering and native behavior.
// Height is the custom titlebar height; buttons are centered vertically in it.
// Left is the first button's left inset. OffsetY moves the center down (or up
// when negative). Spacing is the distance between button centers; zero keeps
// AppKit's spacing. A nil layout in Options keeps AppKit's default placement.
type TrafficLightLayout struct {
	Height, Left, OffsetY, Spacing float32
}

// Options configures a window. Width and Height are in dp; zero uses 640×480.
type Options struct {
	Title         string
	Width, Height int
	Content       core.Widget
	// Overlay is drawn over the whole window, above Content, for hand-written
	// Gio content; el views declare overlays with cx.Overlay instead. It should
	// take no space while it has nothing to show.
	Overlay core.Widget
	// Shortcuts maps accelerators to callbacks while the window has focus, e.g.
	// "mod+," (Cmd on macOS, Ctrl elsewhere), "ctrl+shift+s", "esc".
	Shortcuts map[string]func()
	OnClose   func()
	// Frameless hides the system title bar so the content can draw its own,
	// e.g. a kit.TitleBar; the content then starts at the window's top edge.
	Frameless bool
	// NativeTrafficLights keeps AppKit's standard window buttons visible over
	// frameless content on macOS. Reserve the top-left titlebar area in Content.
	// Other platforms ignore this option.
	NativeTrafficLights bool
	TrafficLightLayout  *TrafficLightLayout
}

type Window struct {
	win                       *gioapp.Window // nil in automation mode
	virt                      *virtual       // non-nil in automation mode
	opts                      Options
	shortcuts                 []shortcut
	root                      root
	closed                    bool // guarded by the frame lock
	focused                   bool
	nativeTrafficLightsDirty  bool
	nativeView, lastTitleView uintptr
	x11Window                 uint32               // the X11 window ID on Linux X11, for Activate
	waylandDisplay            atomic.Pointer[byte] // the window's wl_display on Linux Wayland
	waylandSurface            atomic.Pointer[byte] // its wl_surface, for Activate
	titleArea, lastTitleArea  [4]float32
	maximized                 bool // guarded by the frame lock; from the platform's config
	deco                      widget.Decorations
	drawsTitle                bool          // Keel draws the title bar; see decorations.go
	shown                     chan struct{} // closed at the first frame or at destruction
}

// Open creates and shows a centered window where the platform supports it.
// Call it before Main or from any callback.
// It panics on an invalid shortcut, which is a programming error.
func Open(o Options) *Window {
	w := newWindow(o)
	registerDevelopmentWindow(w)
	loadRunIcon()
	if offScreen() {
		openVirtual(w, true)
		return w
	}
	w.win = new(gioapp.Window)
	w.win.Option(gioapp.Title(o.Title), gioapp.Size(unit.Dp(w.opts.Width), unit.Dp(w.opts.Height)), gioapp.Decorated(askDecorations(o.Frameless)))
	loop.Register(w, w.win.Invalidate)
	if automating() {
		openVirtual(w, false) // shadow of the real window, driven by agents
	}
	go w.run()
	return w
}

func newWindow(o Options) *Window {
	if o.TrafficLightLayout != nil {
		layout := *o.TrafficLightLayout
		o.TrafficLightLayout = &layout
	}
	if o.Width == 0 {
		o.Width = 640
	}
	if o.Height == 0 {
		o.Height = 480
	}
	return &Window{opts: o, focused: true, shortcuts: mustParseShortcuts(o.Shortcuts), shown: make(chan struct{})}
}

// SetTrafficLightLayout updates native button placement without recreating
// the window. It is safe from callbacks or background goroutines; the change
// is applied on the next frame. NativeTrafficLights must be enabled in Options.
func (w *Window) SetTrafficLightLayout(layout TrafficLightLayout) {
	core.Update(func() {
		if !w.closed {
			w.opts.TrafficLightLayout = &layout
			w.nativeTrafficLightsDirty = true
		}
	})
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
	watchSystemPreferences()
	gioapp.Main()
}

// Close closes the window as if the user clicked its close button.
func (w *Window) Close() { w.perform(system.ActionClose) }

// Raise brings the window to the front.
func (w *Window) Raise() { w.perform(system.ActionRaise) }

// WaylandDisplay is this window's wl_display on Linux Wayland, nil elsewhere
// or before the window is shown. Pass it to native/clipboard's
// UseWaylandDisplay to read the clipboard while this window has focus.
func (w *Window) WaylandDisplay() unsafe.Pointer { return unsafe.Pointer(w.waylandDisplay.Load()) }

// Activate brings the window to the front with an activation token that
// another program granted, such as notification.Activation.Token after a
// system notification was clicked. Window managers let a token through
// their focus-stealing prevention, where a plain Raise may only flash the
// taskbar. On Wayland the token goes to xdg-activation; on X11 it is a
// startup ID. Elsewhere, with an empty token, or if the platform refuses,
// Activate is Raise.
func (w *Window) Activate(token string) {
	if token == "" || w.win == nil || !platformActivate(w, token) {
		w.Raise()
	}
}

// Minimize hides the window in the Dock or taskbar.
func (w *Window) Minimize() { w.perform(system.ActionMinimize) }

// ToggleMaximize maximizes the window (zooms it on macOS), or restores it
// when it is maximized.
func (w *Window) ToggleMaximize() {
	if w.win == nil { // headless: track the state so tests and agents see it
		w.maximized = !w.maximized
		return
	}
	if w.maximized {
		w.perform(system.ActionUnmaximize)
	} else {
		w.perform(system.ActionMaximize)
	}
}

// Maximized reports whether the window is maximized. Call it from UI code.
func (w *Window) Maximized() bool { return w.maximized }

// Frameless reports whether the window draws its own title bar.
func (w *Window) Frameless() bool { return w.opts.Frameless }
func (w *Window) Focused() bool   { return w.focused }
func (w *Window) TitleBarArea(x, y, width, height float32) {
	if w.opts.Frameless {
		w.titleArea = [4]float32{x, y, width, height}
	}
}
func (w *Window) setFocused(focused bool) {
	if w.focused != focused {
		w.focused = focused
		if w.win != nil {
			w.win.Invalidate()
		}
	}
}

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
	positioned := false
	for {
		switch e := w.win.Event().(type) {
		case gioapp.DestroyEvent:
			w.markShown()
			w.destroy()
			return
		case gioapp.ConfigEvent:
			loop.Lock()
			w.nativeTrafficLightsDirty = true
			w.setFocused(e.Config.Focused)
			if m := e.Config.Mode == gioapp.Maximized; m != w.maximized {
				w.maximized = m
				w.win.Invalidate() // title bars show maximize or restore
			}
			if w.updateDecorations(e.Config.Decorated, e.Config.Mode == gioapp.Fullscreen) {
				w.win.Invalidate()
			}
			loop.Unlock()
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
			if !positioned {
				positioned = true
				loop.Lock()
				centerNewWindow(w)
				loop.Unlock()
			}
			w.markShown()
		default:
			platformWindowEvent(w, e)
			iconWindowEvent(w, e)
		}
	}
}

func (w *Window) layout(gtx core.C) {
	defer core.SetCurrentWindow(w)()
	w.titleArea = [4]float32{}
	defer func() { syncTitleBar(w) }()
	w.handleShortcuts(gtx)
	gtx, below := w.belowTitleBar(gtx)
	defer below()
	w.root.Layout(gtx, w.opts.Content)
	if w.opts.Overlay != nil {
		o := gtx
		o.Constraints = layout.Exact(gtx.Constraints.Max)
		w.opts.Overlay.Layout(o)
	}
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
	unregisterDevelopmentWindow(w)
	remaining := loop.Unregister(w)
	loop.Lock()
	w.closed = true
	clearPlatformWindow(w)
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
