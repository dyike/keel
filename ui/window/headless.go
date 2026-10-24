package window

import (
	"fmt"
	"image"
	"sync/atomic"

	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
)

// inMemory is set by OpenHeadless: windows then get no native window, as
// with KEEL_HEADLESS=1, and no automation socket.
var inMemory atomic.Bool

// Headless is a window drawn in memory only, without a native window or
// Main: what an agent driving the app through keel-mcp sees, available to Go
// tests. Package windowtest wraps it for testing.T. Use a Headless from one
// goroutine; Keel runs its callbacks and core.Update functions during the
// calls that render.
type Headless struct{ w *Window }

// OpenHeadless opens o in memory. Width and Height default to 800×600, in
// dp with 1dp = 1px. Reduced motion is turned on, as for agents, so nothing
// is caught mid-animation. Every window the process opens afterwards, for
// example from a callback, is in memory too.
func OpenHeadless(o Options) *Headless {
	inMemory.Store(true)
	loop.Lock()
	theme.SetReducedMotion(true)
	loop.Unlock()
	if o.Width <= 0 {
		o.Width = 800
	}
	if o.Height <= 0 {
		o.Height = 600
	}
	return &Headless{w: Open(o)}
}

// Window is the window being drawn, for its own methods.
func (h *Headless) Window() *Window { return h.w }

// Render lays out frames until nothing asks for an immediate redraw, running
// queued core.Update functions first.
func (h *Headless) Render() {
	if !h.w.closed {
		h.w.render()
	}
}

// Elements renders and lists the window's elements, as keel-mcp reports them:
// a control absorbs the labels inside it, and elements scrolled out of view
// are left out.
func (h *Headless) Elements() []Element {
	if h.w.closed {
		return nil
	}
	return h.w.snapshot()
}

// Find renders and returns the element whose name or value best matches
// text: an exact match wins over a partial one, a control over plain text,
// and of equals the one drawn last (a dialog over the page).
func (h *Headless) Find(text string) (Element, error) {
	if h.w.closed {
		return Element{}, fmt.Errorf("window %q is closed", h.w.opts.Title)
	}
	return h.w.find("", text)
}

// Click clicks the center of the element Find returns for text.
func (h *Headless) Click(text string) error {
	e, err := h.Find(text)
	if err != nil {
		return err
	}
	h.ClickElement(e)
	return nil
}

// ClickElement clicks the center of e.
func (h *Headless) ClickElement(e Element) { h.w.click(e.center()) }

// Type inserts s at the caret of the focused text box, replacing any
// selection. Click a text box first.
func (h *Headless) Type(s string) error { return h.w.typeText(s) }

// Press sends one key chord: "enter", "tab", "shift+tab", "mod+a" (Cmd on
// macOS, Ctrl elsewhere). An unhandled Tab moves focus, as in a window.
func (h *Headless) Press(chord string) error { return h.w.press(chord) }

// Scroll turns the mouse wheel by dy dp (positive: down) over the element
// Find returns for text.
func (h *Headless) Scroll(text string, dy float32) error {
	e, err := h.Find(text)
	if err != nil {
		return err
	}
	h.w.scroll(e.center(), dy)
	return nil
}

// Resize changes the window's size in dp and renders.
func (h *Headless) Resize(width, height int) {
	h.w.virt.setSize(image.Pt(width, height))
	h.Render()
}

// Screenshot renders the window with the GPU and encodes it as PNG.
func (h *Headless) Screenshot() ([]byte, error) { return h.w.screenshot() }

// Close closes the window, running its OnClose callback. The process keeps
// running.
func (h *Headless) Close() {
	if h.w.closed {
		return
	}
	if v := h.w.virt; v.gpu != nil {
		v.gpu.Release()
		v.gpu = nil
	}
	closeVirtual(h.w)
}
