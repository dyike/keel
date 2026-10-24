//go:build linux && !android

package window

// Platform accessibility on Linux: AT-SPI over D-Bus (ui/internal/atspi),
// for Orca and other assistive technology. The first window connects to the
// accessibility bus in the background and registers the application; where
// there is no such bus, nothing else happens. Elements are built only once a
// client has made a request: from then on, after each frame, on the
// window's goroutine, from the semantic tree Gio's driver keeps there.
// Requests that act (press, scroll) arrive on D-Bus goroutines; they are
// queued and run on the window's goroutine after its next event, the one
// place Gio's router is free.

import (
	"image"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/ui/internal/atspi"
	"github.com/dyike/keel/ui/internal/loop"
)

var atspiApp struct {
	once    sync.Once
	mu      sync.Mutex
	bridge  *atspi.Bridge
	next    int
	windows map[int]*axWindow
}

// axWindow is a window's accessibility state.
type axWindow struct {
	w  *Window
	id int

	nodes []input.SemanticNode // window goroutine only
	buf   []atspi.Node

	mu      sync.Mutex
	els     map[string]Element // the latest frame's elements, by key
	pending []func()
}

func axWindowOf(w *Window) *axWindow {
	atspiApp.mu.Lock()
	defer atspiApp.mu.Unlock()
	for _, a := range atspiApp.windows {
		if a.w == w {
			return a
		}
	}
	return nil
}

func accessibilityOpen(w *Window) {
	atspiApp.once.Do(func() { go startATSPI() })
	atspiApp.mu.Lock()
	if atspiApp.windows == nil {
		atspiApp.windows = map[int]*axWindow{}
	}
	atspiApp.next++
	a := &axWindow{w: w, id: atspiApp.next, els: map[string]Element{}}
	atspiApp.windows[a.id] = a
	b := atspiApp.bridge
	atspiApp.mu.Unlock()
	if b != nil {
		b.AddWindow(a.id, w.opts.Title)
		if b.Asked() {
			w.axActive.Store(true)
		}
	}
}

func startATSPI() {
	conn, err := atspi.Connect()
	if err != nil {
		return // no accessibility bus: nobody to serve
	}
	b, err := atspi.New(conn, filepath.Base(os.Args[0]), func(id int) atspi.Actions {
		atspiApp.mu.Lock()
		defer atspiApp.mu.Unlock()
		if a := atspiApp.windows[id]; a != nil {
			return a
		}
		return nil
	}, func() {
		atspiApp.mu.Lock()
		for _, a := range atspiApp.windows {
			a.w.axActive.Store(true)
		}
		atspiApp.mu.Unlock()
		loop.SetAccessibility(true) // redraws every window
	})
	if err != nil {
		conn.Close()
		return
	}
	atspiApp.mu.Lock()
	atspiApp.bridge = b
	var open []*axWindow
	for _, a := range atspiApp.windows {
		open = append(open, a)
	}
	atspiApp.mu.Unlock()
	for _, a := range open {
		b.AddWindow(a.id, a.w.opts.Title)
	}
	_ = b.Register() // without a registry, clients that know the bus name still can ask
}

func accessibilityClose(w *Window) {
	a := axWindowOf(w)
	if a == nil {
		return
	}
	atspiApp.mu.Lock()
	delete(atspiApp.windows, a.id)
	b := atspiApp.bridge
	atspiApp.mu.Unlock()
	if b != nil {
		go b.RemoveWindow(a.id)
	}
}

func accessibilityFocus(w *Window, focused bool) {
	atspiApp.mu.Lock()
	b := atspiApp.bridge
	atspiApp.mu.Unlock()
	if a := axWindowOf(w); a != nil && b != nil {
		go b.SetFocused(a.id, focused)
	}
}

// accessibilityFrame hands the frame's elements to the bridge.
func accessibilityFrame(w *Window, size image.Point) {
	if !w.axActive.Load() {
		return
	}
	a := axWindowOf(w)
	atspiApp.mu.Lock()
	b := atspiApp.bridge
	atspiApp.mu.Unlock()
	if a == nil || b == nil {
		return
	}
	a.nodes = w.win.AppendSemantics(a.nodes[:0])
	els := elementsOf(a.nodes, image.Rectangle{Max: size}, true)
	a.buf = a.buf[:0]
	byKey := make(map[string]Element, len(els))
	for _, e := range els {
		key := strconv.FormatUint(uint64(e.id), 10)
		byKey[key] = e
		a.buf = append(a.buf, atspi.Node{Key: key, Role: e.Role, Name: e.Name, Value: e.Value,
			Checked: e.Checked, Selected: e.Selected, Disabled: e.Disabled,
			Bounds: e.box(), Viewport: e.viewport, Parent: e.parent})
	}
	a.mu.Lock()
	a.els = byKey
	a.mu.Unlock()
	b.Update(a.id, w.opts.Title, a.buf)
}

// accessibilityInput runs the requests queued since the last event.
func accessibilityInput(w *Window) {
	a := axWindowOf(w)
	if a == nil {
		return
	}
	a.mu.Lock()
	todo := a.pending
	a.pending = nil
	a.mu.Unlock()
	for _, fn := range todo {
		fn()
	}
}

// queue runs fn with the element named key on the window's goroutine.
func (a *axWindow) queue(key string, fn func(Element)) bool {
	a.mu.Lock()
	if _, ok := a.els[key]; !ok {
		a.mu.Unlock()
		return false
	}
	a.pending = append(a.pending, func() {
		a.mu.Lock()
		e, ok := a.els[key]
		a.mu.Unlock()
		if ok {
			fn(e)
		}
	})
	a.mu.Unlock()
	a.w.win.Invalidate() // wakes Event
	return true
}

func (a *axWindow) Press(key string) bool {
	return a.queue(key, func(e Element) { pressElement(a.w, e) })
}

func (a *axWindow) ScrollTo(key string) bool {
	return a.queue(key, func(e Element) { scrollToElement(a.w, e) })
}

func (a *axWindow) ScrollPage(key string, dx, dy int) bool {
	return a.queue(key, func(e Element) { scrollPage(a.w, e, dx, dy) })
}
