// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"image"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
)

// UITouchPhase values from UIKit.
const (
	uiTouchPhaseBegan     = 0
	uiTouchPhaseMoved     = 1
	uiTouchPhaseEnded     = 3
	uiTouchPhaseCancelled = 4
)

type UIKitViewEvent struct {
	// ViewController is a CFTypeRef for the UIViewController backing a Window.
	ViewController uintptr
}

type window struct {
	view        uintptr
	w           *callbacks
	displayLink *displayLink
	loop        *eventLoop

	hidden bool
	cursor pointer.Cursor
	config Config

	pointerMap []uintptr
}

var mainWindow = newWindowRendezvous()

func init() {
	// Darwin requires UI operations happen on the main thread only.
	runtime.LockOSThread()
}

func onCreate(view, controller uintptr) {
	wopts := <-mainWindow.out
	w := &window{
		view: view,
		w:    wopts.window,
	}
	w.loop = newEventLoop(w.w, w.wakeup)
	w.w.SetDriver(w)
	mainWindow.windows <- struct{}{}
	dl, err := newDisplayLink(func() {
		w.draw(false)
	})
	if err != nil {
		w.w.ProcessEvent(DestroyEvent{Err: err})
		return
	}
	w.displayLink = dl
	setIOSView(view, w)
	w.Configure(wopts.options)
	w.ProcessEvent(UIKitViewEvent{ViewController: controller})
}

func (w *window) draw(sync bool) {
	if w.hidden {
		return
	}
	params := viewDrawParams(w.view)
	if params.width == 0 || params.height == 0 {
		return
	}
	const inchPrDp = 1.0 / 163
	m := unit.Metric{
		PxPerDp: float32(params.dpi) * inchPrDp,
		PxPerSp: float32(params.sdpi) * inchPrDp,
	}
	dppp := unit.Dp(1. / m.PxPerDp)
	w.ProcessEvent(frameEvent{
		FrameEvent: FrameEvent{
			Now: time.Now(),
			Size: image.Point{
				X: int(params.width + .5),
				Y: int(params.height + .5),
			},
			Insets: Insets{
				Top:    unit.Dp(params.top) * dppp,
				Bottom: unit.Dp(params.bottom) * dppp,
				Left:   unit.Dp(params.left) * dppp,
				Right:  unit.Dp(params.right) * dppp,
			},
			Metric: m,
		},
		Sync: sync,
	})
}

func (w *window) onStop() {
	w.hidden = true
}

func (w *window) onStart() {
	w.hidden = false
	w.draw(true)
}

func (w *window) onDestroy() {
	w.ProcessEvent(UIKitViewEvent{})
	w.ProcessEvent(DestroyEvent{})
	if w.displayLink != nil {
		w.displayLink.Close()
		w.displayLink = nil
	}
	deleteIOSView(w.view)
	w.view = 0
}

func (w *window) onFocus(focus bool) {
	w.config.Focused = focus
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func onLowMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

func (w *window) onText(text string) {
	w.w.EditorInsert(text)
}

func (w *window) onTouch(last bool, touchRef uintptr, phase int, x, y, ti float64) {
	var kind pointer.Kind
	switch phase {
	case uiTouchPhaseBegan:
		kind = pointer.Press
	case uiTouchPhaseMoved:
		kind = pointer.Move
	case uiTouchPhaseEnded:
		kind = pointer.Release
	case uiTouchPhaseCancelled:
		kind = pointer.Cancel
	default:
		return
	}
	t := time.Duration(ti * float64(time.Second))
	p := f32.Point{X: float32(x), Y: float32(y)}
	w.ProcessEvent(pointer.Event{
		Kind:      kind,
		Source:    pointer.Touch,
		PointerID: w.lookupTouch(last, touchRef),
		Position:  p,
		Time:      t,
	})
}

func (w *window) ReadClipboard() {
	content := readClipboard()
	w.ProcessEvent(transfer.DataEvent{
		Type: "application/text",
		Open: func() io.ReadCloser {
			return io.NopCloser(strings.NewReader(content))
		},
	})
}

func (w *window) WriteClipboard(mime string, s []byte) {
	writeClipboard(string(s))
}

func (w *window) Configure([]Option) {
	// Decorations are never disabled.
	w.config.Decorated = true
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func (w *window) EditorStateChanged(old, new editorState) {}

func (w *window) Perform(system.Action) {}

func (w *window) SetAnimating(anim bool) {
	if anim {
		w.displayLink.Start()
	} else {
		w.displayLink.Stop()
	}
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor = windowSetCursor(w.cursor, cursor)
}

func (w *window) onKeyCommand(name key.Name) {
	w.ProcessEvent(key.Event{
		Name: name,
	})
}

// lookupTouch maps an UITouch pointer value to an index. If
// last is set, the map is cleared.
func (w *window) lookupTouch(last bool, touch uintptr) pointer.ID {
	id := -1
	for i, ref := range w.pointerMap {
		if ref == touch {
			id = i
			break
		}
	}
	if id == -1 {
		id = len(w.pointerMap)
		w.pointerMap = append(w.pointerMap, touch)
	}
	if last {
		w.pointerMap = w.pointerMap[:0]
	}
	return pointer.ID(id)
}

func (w *window) contextView() uintptr {
	return w.view
}

func (w *window) ShowTextInput(show bool) {
	if show {
		showTextInput(w.view)
	} else {
		hideTextInput(w.view)
	}
}

func (w *window) SetInputHint(_ key.InputHint) {}

func (w *window) ProcessEvent(e event.Event) {
	w.w.ProcessEvent(e)
	w.loop.FlushEvents()
}

func (w *window) Event() event.Event {
	return w.loop.Event()
}

func (w *window) Invalidate() {
	w.loop.Invalidate()
}

func (w *window) Run(f func()) {
	w.loop.Run(f)
}

func (w *window) Frame(frame *op.Ops) {
	w.loop.Frame(frame)
}

func newWindow(win *callbacks, options []Option) {
	mainWindow.in <- windowAndConfig{win, options}
	<-mainWindow.windows
}

var mainMode = mainModeUndefined

const (
	mainModeUndefined = iota
	mainModeExe
	mainModeLibrary
)

func osMain() {
	switch mainMode {
	case mainModeUndefined:
		if !isMainThread() {
			panic("app.Main must be run on the main goroutine")
		}

		mainMode = mainModeExe
		iosApplicationMain(os.Args)
	case mainModeExe:
		panic("app.Main may be called only once")
	case mainModeLibrary:
		// Do nothing, we're embedded as a library.
	}
	select {}
}

func onOpenURI(uri string) {
	evt, err := newURLEvent(uri)
	if err != nil {
		return
	}
	processGlobalEvent(evt)
}

func gioRunMain() {
	if !isMainThread() {
		panic("app.Main must be run on the main goroutine")
	}
	switch mainMode {
	case mainModeUndefined:
		mainMode = mainModeLibrary
		runMain()
	case mainModeExe:
		// Do nothing, main has already been called.
	}
}

func (UIKitViewEvent) implementsViewEvent() {}
func (UIKitViewEvent) ImplementsEvent()     {}
func (u UIKitViewEvent) Valid() bool {
	return u != (UIKitViewEvent{})
}
