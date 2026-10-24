//go:build windows

package window

// Platform accessibility on Windows: UI Automation (ui/internal/uia), for
// Narrator and other assistive technology. A window answers WM_GETOBJECT in
// the procedure Keel already installs on Gio's window (see
// application_menu_windows.go). Nothing runs until the first such request;
// from then on Gio hands each frame's semantic tree to an observer on the
// window's thread, which builds the elements (elementsOf) for the provider.
// UI Automation calls arrive on its own threads; presses and scrolls become
// input queued for the window (Window.QueueInput).

import (
	"image"
	"strconv"
	"sync"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uia"
)

var uiaWindows struct {
	sync.Mutex
	byHWND map[uintptr]*uiaWindow
}

type uiaWindow struct {
	w   *Window
	uia *uia.Window

	buf       []uia.Node // window thread only
	observing bool       // window thread only

	mu  sync.Mutex
	els map[string]Element // the latest frame's elements, by key
}

func accessibilityOpen(*Window)               {}
func accessibilityFrame(*Window, image.Point) {}
func accessibilityFocus(*Window, bool)        {}
func accessibilityClose(*Window)              {}

// accessibilityGetObject answers WM_GETOBJECT on the window's thread.
func accessibilityGetObject(w *Window, hwnd, wparam, lparam uintptr) (uintptr, bool) {
	uiaWindows.Lock()
	if uiaWindows.byHWND == nil {
		uiaWindows.byHWND = map[uintptr]*uiaWindow{}
	}
	a := uiaWindows.byHWND[hwnd]
	if a == nil {
		a = &uiaWindow{w: w, els: map[string]Element{}}
		a.uia = uia.NewWindow(hwnd, a)
		uiaWindows.byHWND[hwnd] = a
	}
	uiaWindows.Unlock()
	result, ok := a.uia.GetObject(wparam, lparam)
	if ok && !a.observing {
		a.observing = true
		w.win.ObserveSemantics(a.observe)
		// Describe what is scrolled out of view and draw a frame. Not from
		// here: invalidating inside the window procedure re-enters Gio.
		go func() {
			loop.SetAccessibility(true)
			w.win.Invalidate()
		}()
	}
	return result, ok
}

// accessibilityDestroyed disconnects a window's providers while Windows
// destroys it.
func accessibilityDestroyed(hwnd uintptr) {
	uiaWindows.Lock()
	a := uiaWindows.byHWND[hwnd]
	delete(uiaWindows.byHWND, hwnd)
	uiaWindows.Unlock()
	if a != nil {
		a.w.win.ObserveSemantics(nil)
		a.uia.Disconnect()
	}
}

// observe runs on the window's thread after each frame.
func (a *uiaWindow) observe(nodes []input.SemanticNode, size image.Point) {
	els := elementsOf(nodes, image.Rectangle{Max: size}, true)
	byKey := make(map[string]Element, len(els))
	a.buf = a.buf[:0]
	for _, e := range els {
		key := strconv.FormatUint(uint64(e.id), 10)
		byKey[key] = e
		a.buf = append(a.buf, uia.Node{Key: key, Role: e.Role, Name: e.Name, Value: e.Value,
			Checked: e.Checked, Selected: e.Selected, Disabled: e.Disabled, Focused: e.focused,
			Bounds: e.box(), Viewport: e.viewport, Parent: e.parent})
	}
	a.mu.Lock()
	a.els = byKey
	a.mu.Unlock()
	a.uia.Update(a.buf)
}

func (a *uiaWindow) queue(key string, events func(Element) []event.Event) bool {
	a.mu.Lock()
	e, ok := a.els[key]
	a.mu.Unlock()
	if !ok {
		return false
	}
	evs := events(e)
	if len(evs) == 0 {
		return false
	}
	a.w.win.QueueInput(evs...)
	return true
}

func (a *uiaWindow) Press(key string) bool    { return a.queue(key, clickEvents) }
func (a *uiaWindow) ScrollTo(key string) bool { return a.queue(key, scrollToEvents) }
