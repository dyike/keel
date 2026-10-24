// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11

package app

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"

	syscall "golang.org/x/sys/unix"

	"github.com/dyike/keel/gio/app/internal/ibus"
	"github.com/dyike/keel/gio/app/internal/xkb"
)

// The Xlib calls live in os_x11_cgo.go and os_x11_purego.go. Keel patch:
// this file holds the logic both share, over Go definitions of the X
// structures it reads and writes; os_x11_cgo.go checks their layout and
// the values below against the headers.

const (
	_NET_WM_STATE_REMOVE = 0
	_NET_WM_STATE_ADD    = 1
)

// Xlib values.
const (
	xFalse            = 0
	xTrue             = 1
	xNone             = 0
	xCurrentTime      = 0
	xNoEventMask      = 0
	xKeyPress         = 2
	xKeyRelease       = 3
	xButtonPress      = 4
	xButtonRelease    = 5
	xMotionNotify     = 6
	xFocusIn          = 9
	xFocusOut         = 10
	xExpose           = 12
	xConfigureNotify  = 22
	xSelectionRequest = 30
	xSelectionNotify  = 31
	xClientMessage    = 33

	xKeyPressMask             = 1 << 0
	xKeyReleaseMask           = 1 << 1
	xButtonPressMask          = 1 << 2
	xButtonReleaseMask        = 1 << 3
	xPointerMotionMask        = 1 << 6
	xExposureMask             = 1 << 15
	xStructureNotifyMask      = 1 << 17
	xSubstructureNotifyMask   = 1 << 19
	xSubstructureRedirectMask = 1 << 20
	xFocusChangeMask          = 1 << 21

	xCWBackPixmap       = 1 << 0
	xCWOverrideRedirect = 1 << 9
	xCWEventMask        = 1 << 11
	xCopyFromParent     = 0
	xInputOutput        = 1
	xPMinSize           = 1 << 4
	xPMaxSize           = 1 << 5
	xInputHint          = 1 << 0
	xPropModeReplace    = 0

	xButton1 = 1
	xButton2 = 2
	xButton3 = 3
	xButton4 = 4
	xButton5 = 5

	xkbUseCoreKbd            = 0x0100
	xkbMajorVersion          = 1
	xkbMinorVersion          = 0
	xkbNewKeyboardNotify     = 0
	xkbMapNotify             = 1
	xkbStateNotify           = 2
	xkbNewKeyboardNotifyMask = 1 << 0
	xkbMapNotifyMask         = 1 << 1
	xkbStateNotifyMask       = 1 << 2
)

// X structures. C's int is int32, long is int and unsigned long (Window,
// Atom, Time, XID) is uintptr, on 32- and 64-bit systems alike.
type (
	xWindow = uintptr
	xAtom   = uintptr

	// xEvent is the XEvent union: long pad[24].
	xEvent [24]int

	xAnyEvent struct {
		Type      int32
		Serial    uintptr
		SendEvent int32
		Display   unsafe.Pointer
		Window    xWindow
	}

	xKeyEvent struct {
		Type                    int32
		Serial                  uintptr
		SendEvent               int32
		Display                 unsafe.Pointer
		Window, Root, Subwindow xWindow
		Time                    uintptr
		X, Y, XRoot, YRoot      int32
		State, Keycode          uint32
		SameScreen              int32
	}

	xButtonEvent struct {
		Type                    int32
		Serial                  uintptr
		SendEvent               int32
		Display                 unsafe.Pointer
		Window, Root, Subwindow xWindow
		Time                    uintptr
		X, Y, XRoot, YRoot      int32
		State, Button           uint32
		SameScreen              int32
	}

	xMotionEvent struct {
		Type                    int32
		Serial                  uintptr
		SendEvent               int32
		Display                 unsafe.Pointer
		Window, Root, Subwindow xWindow
		Time                    uintptr
		X, Y, XRoot, YRoot      int32
		State                   uint32
		IsHint                  int8
		SameScreen              int32
	}

	xExposeEvent struct {
		Type                       int32
		Serial                     uintptr
		SendEvent                  int32
		Display                    unsafe.Pointer
		Window                     xWindow
		X, Y, Width, Height, Count int32
	}

	xConfigureEvent struct {
		Type                             int32
		Serial                           uintptr
		SendEvent                        int32
		Display                          unsafe.Pointer
		Event, Window                    xWindow
		X, Y, Width, Height, BorderWidth int32
		Above                            xWindow
		OverrideRedirect                 int32
	}

	xSelectionEvent struct {
		Type                        int32
		Serial                      uintptr
		SendEvent                   int32
		Display                     unsafe.Pointer
		Requestor                   xWindow
		Selection, Target, Property xAtom
		Time                        uintptr
	}

	xSelectionRequestEvent struct {
		Type                        int32
		Serial                      uintptr
		SendEvent                   int32
		Display                     unsafe.Pointer
		Owner, Requestor            xWindow
		Selection, Target, Property xAtom
		Time                        uintptr
	}

	xClientMessageEvent struct {
		Type        int32
		Serial      uintptr
		SendEvent   int32
		Display     unsafe.Pointer
		Window      xWindow
		MessageType xAtom
		Format      int32
		Data        [5]int // the long member of the data union
	}

	xkbAnyEvent struct {
		Type      int32
		Serial    uintptr
		SendEvent int32
		Display   unsafe.Pointer
		Time      uintptr
		XkbType   int32
		Device    uint32
	}

	xkbStateNotifyEvent struct {
		Type                                                   int32
		Serial                                                 uintptr
		SendEvent                                              int32
		Display                                                unsafe.Pointer
		Time                                                   uintptr
		XkbType, Device                                        int32
		Changed                                                uint32
		Group, BaseGroup, LatchedGroup, LockedGroup            int32
		Mods, BaseMods, LatchedMods, LockedMods                uint32
		CompatState                                            int32
		GrabMods, CompatGrabMods, LookupMods, CompatLookupMods uint8
		PtrButtons                                             int32
		Keycode                                                uint8
		EventType, ReqMajor, ReqMinor                          int8
	}

	xSizeHints struct {
		Flags                                    int
		X, Y, Width, Height                      int32
		MinWidth, MinHeight, MaxWidth, MaxHeight int32
		WidthInc, HeightInc                      int32
		MinAspect, MaxAspect                     struct{ X, Y int32 }
		BaseWidth, BaseHeight, WinGravity        int32
	}

	xTextProperty struct {
		Value    unsafe.Pointer
		Encoding xAtom
		Format   int32
		Nitems   uintptr
	}

	xWindowAttributes struct {
		X, Y, Width, Height, BorderWidth, Depth          int32
		Visual                                           unsafe.Pointer
		Root                                             xWindow
		Class                                            int32
		BitGravity, WinGravity, BackingStore             int32
		BackingPlanes, BackingPixel                      uintptr
		SaveUnder                                        int32
		Colormap                                         uintptr
		MapInstalled                                     int32
		MapState                                         int32
		AllEventMasks, YourEventMask, DoNotPropagateMask int
		OverrideRedirect                                 int32
		Screen                                           unsafe.Pointer
	}

	xSetWindowAttributes struct {
		BackgroundPixmap, BackgroundPixel    uintptr
		BorderPixmap, BorderPixel            uintptr
		BitGravity, WinGravity, BackingStore int32
		BackingPlanes, BackingPixel          uintptr
		SaveUnder                            int32
		EventMask, DoNotPropagateMask        int
		OverrideRedirect                     int32
		Colormap, Cursor                     uintptr
	}

	xWMHints struct {
		Flags        int
		Input        int32
		InitialState int32
		IconPixmap   uintptr
		IconWindow   xWindow
		IconX, IconY int32
		IconMask     uintptr
		WindowGroup  uintptr
	}

	xClassHint struct {
		ResName, ResClass unsafe.Pointer
	}

	xrmValue struct {
		Size uint32
		Addr unsafe.Pointer
	}
)

// cString returns s NUL-terminated in Go memory.
func cString(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// goStringN copies n bytes at p, which points outside the Go heap.
func goStringN(p unsafe.Pointer, n int) string {
	if p == nil || n == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(p), n))
}

// goString copies the NUL-terminated string at p.
func goString(p unsafe.Pointer) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}
	return goStringN(p, n)
}

type x11Window struct {
	w            *callbacks
	x            unsafe.Pointer // Display *
	xkb          *xkb.Context
	xkbEventBase int32
	xw           xWindow
	ibus         *ibus.IBus

	atoms struct {
		// "UTF8_STRING".
		utf8string xAtom
		// "text/plain;charset=utf-8".
		plaintext xAtom
		// "TARGETS"
		targets xAtom
		// "CLIPBOARD".
		clipboard xAtom
		// "PRIMARY".
		primary xAtom
		// "CLIPBOARD_CONTENT", the clipboard destination property.
		clipboardContent xAtom
		// "WM_DELETE_WINDOW"
		evDelWindow xAtom
		// "ATOM"
		atom xAtom
		// "GTK_TEXT_BUFFER_CONTENTS"
		gtk_text_buffer_contents xAtom
		// "_NET_WM_NAME"
		wmName xAtom
		// "_NET_WM_STATE"
		wmState xAtom
		// "_NET_WM_STATE_FULLSCREEN"
		wmStateFullscreen xAtom
		// "_NET_ACTIVE_WINDOW"
		wmActiveWindow xAtom
		// _NET_WM_STATE_MAXIMIZED_HORZ
		wmStateMaximizedHorz xAtom
		// _NET_WM_STATE_MAXIMIZED_VERT
		wmStateMaximizedVert xAtom
	}
	metric unit.Metric
	notify struct {
		read, write int
	}

	animating bool

	pointerBtns pointer.Buttons

	clipboard struct {
		content []byte
	}
	cursor    pointer.Cursor
	config    Config
	closing   *ClosingEvent
	delivered *ClosingEvent

	wakeups     chan struct{}
	dispatchFns chan func()
	handler     x11EventHandler
	buf         [100]byte
}

var (
	newX11EGLContext    func(w *x11Window) (context, error)
	newX11VulkanContext func(w *x11Window) (context, error)
)

// X11 and Vulkan doesn't work reliably on NVIDIA systems.
// See https://gioui.org/issue/347.
const vulkanBuggy = true

func (w *x11Window) NewContext() (context, error) {
	var firstErr error
	if f := newX11VulkanContext; f != nil && !vulkanBuggy {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if f := newX11EGLContext; f != nil {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errors.New("x11: no available GPU backends")
}

func (w *x11Window) SetAnimating(anim bool) {
	w.animating = anim
}

func (w *x11Window) ReadClipboard() {
	xDeleteProperty(w.x, w.xw, w.atoms.clipboardContent)
	xConvertSelection(w.x, w.atoms.clipboard, w.atoms.utf8string, w.atoms.clipboardContent, w.xw, xCurrentTime)
}

func (w *x11Window) WriteClipboard(mime string, s []byte) {
	w.clipboard.content = s
	xSetSelectionOwner(w.x, w.atoms.clipboard, w.xw, xCurrentTime)
	xSetSelectionOwner(w.x, w.atoms.primary, w.xw, xCurrentTime)
}

func (w *x11Window) Configure(options []Option) {
	var shints xSizeHints
	prev := w.config
	cnf := w.config
	cnf.apply(w.metric, options)
	// Decorations are never disabled.
	cnf.Decorated = true

	switch cnf.Mode {
	case Fullscreen:
		switch prev.Mode {
		case Fullscreen:
		case Minimized:
			w.raise()
			fallthrough
		default:
			w.config.Mode = Fullscreen
			w.sendWMStateEvent(_NET_WM_STATE_ADD, w.atoms.wmStateFullscreen, 0)
		}
	case Minimized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Minimized
			screen := xDefaultScreen(w.x)
			xIconifyWindow(w.x, w.xw, screen)
		}
	case Maximized:
		switch prev.Mode {
		case Fullscreen:
		case Minimized:
			w.raise()
			fallthrough
		default:
			w.config.Mode = Maximized
			w.sendWMStateEvent(_NET_WM_STATE_ADD, w.atoms.wmStateMaximizedHorz, w.atoms.wmStateMaximizedVert)
			w.setTitle(prev, cnf)
		}
	case Windowed:
		switch prev.Mode {
		case Fullscreen:
			w.config.Mode = Windowed
			w.sendWMStateEvent(_NET_WM_STATE_REMOVE, w.atoms.wmStateFullscreen, 0)
			xResizeWindow(w.x, w.xw, uint32(cnf.Size.X), uint32(cnf.Size.Y))
		case Minimized:
			w.config.Mode = Windowed
			w.raise()
		case Maximized:
			w.config.Mode = Windowed
			w.sendWMStateEvent(_NET_WM_STATE_REMOVE, w.atoms.wmStateMaximizedHorz, w.atoms.wmStateMaximizedVert)
		}
		w.setTitle(prev, cnf)
		if prev.Size != cnf.Size {
			w.config.Size = cnf.Size
			xResizeWindow(w.x, w.xw, uint32(cnf.Size.X), uint32(cnf.Size.Y))
		}
		if prev.MinSize != cnf.MinSize {
			w.config.MinSize = cnf.MinSize
			shints.MinWidth = int32(cnf.MinSize.X)
			shints.MinHeight = int32(cnf.MinSize.Y)
			shints.Flags = xPMinSize
		}
		if prev.MaxSize != cnf.MaxSize {
			w.config.MaxSize = cnf.MaxSize
			shints.MaxWidth = int32(cnf.MaxSize.X)
			shints.MaxHeight = int32(cnf.MaxSize.Y)
			shints.Flags = shints.Flags | xPMaxSize
		}
		if shints.Flags != 0 {
			xSetWMNormalHints(w.x, w.xw, &shints)
		}
	}
	if cnf.Decorated != prev.Decorated {
		w.config.Decorated = cnf.Decorated
	}
	// Keel patch: see Perform.
	xFlush(w.x)
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func (w *x11Window) setTitle(prev, cnf Config) {
	if prev.Title != cnf.Title {
		title := cnf.Title
		ctitle := cString(title)
		xStoreName(w.x, w.xw, ctitle)
		// set _NET_WM_NAME as well for UTF-8 support in window title.
		xSetTextProperty(w.x, w.xw, ctitle[:len(title)], w.atoms.utf8string, 8, w.atoms.wmName)
	}
}

func (w *x11Window) Perform(acts system.Action) {
	walkActions(acts, func(a system.Action) {
		switch a {
		case system.ActionCenter:
			w.center()
		case system.ActionRaise:
			w.raise()
		}
	})
	if acts&system.ActionClose != 0 {
		w.close()
	}
	// Keel patch: Perform runs outside the event loop, which waits in poll
	// for the X connection; Xlib only buffers these requests, so flush
	// them or a close never reaches the server.
	xFlush(w.x)
}

func (w *x11Window) center() {
	screen := xDefaultScreen(w.x)
	width := xDisplayWidth(w.x, screen)
	height := xDisplayHeight(w.x, screen)

	var attrs xWindowAttributes
	xGetWindowAttributes(w.x, w.xw, &attrs)
	width -= attrs.BorderWidth
	height -= attrs.BorderWidth

	sz := w.config.Size
	x := (int(width) - sz.X) / 2
	y := (int(height) - sz.Y) / 2

	xMoveResizeWindow(w.x, w.xw, int32(x), int32(y), uint32(sz.X), uint32(sz.Y))
}

func (w *x11Window) raise() {
	xev := new(xEvent)
	ev := (*xClientMessageEvent)(unsafe.Pointer(xev))
	*ev = xClientMessageEvent{
		Type:        xClientMessage,
		Display:     w.x,
		Window:      w.xw,
		MessageType: w.atoms.wmActiveWindow,
		Format:      32,
	}
	xSendEvent(
		w.x,
		xDefaultRootWindow(w.x), // MUST be the root window
		xFalse,
		xSubstructureNotifyMask|xSubstructureRedirectMask,
		xev,
	)
	xMapRaised(w.display(), w.xw)
}

func (w *x11Window) SetCursor(cursor pointer.Cursor) {
	if cursor == pointer.CursorNone {
		w.cursor = cursor
		xFixesHideCursor(w.x, w.xw)
		return
	}

	xcursor := xCursor[cursor]
	c := xcursorLibraryLoadCursor(w.x, cString(xcursor))
	if c == 0 {
		cursor = pointer.CursorDefault
	}
	w.cursor = cursor
	// If c if null (i.e. cursor was not found),
	// XDefineCursor will use the default cursor.
	xDefineCursor(w.x, w.xw, c)
}

func (w *x11Window) ShowTextInput(show bool) {}

func (w *x11Window) SetInputHint(_ key.InputHint) {}

func (w *x11Window) EditorStateChanged(old, new editorState) {
	if shouldCancelComposition(old, new) {
		w.ibus.Reset()
		w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
	}
}

// close asks the event loop to process a close request.
func (w *x11Window) close() {
	xev := new(xEvent)
	ev := (*xClientMessageEvent)(unsafe.Pointer(xev))
	*ev = xClientMessageEvent{
		Type:        xClientMessage,
		Display:     w.x,
		Window:      w.xw,
		MessageType: w.atom("WM_PROTOCOLS", true),
		Format:      32,
	}
	ev.Data[0] = int(w.atoms.evDelWindow)
	ev.Data[1] = xCurrentTime
	w.ibus.Stop()
	xSendEvent(w.x, w.xw, xFalse, xNoEventMask, xev)
}

// action is one of _NET_WM_STATE_REMOVE, _NET_WM_STATE_ADD.
func (w *x11Window) sendWMStateEvent(action int, atom1, atom2 xAtom) {
	xev := new(xEvent)
	ev := (*xClientMessageEvent)(unsafe.Pointer(xev))
	*ev = xClientMessageEvent{
		Type:        xClientMessage,
		Display:     w.x,
		Window:      w.xw,
		MessageType: w.atoms.wmState,
		Format:      32,
	}
	ev.Data[0] = action
	ev.Data[1] = int(atom1)
	ev.Data[2] = int(atom2)
	ev.Data[3] = 1 // application

	xSendEvent(
		w.x,
		xDefaultRootWindow(w.x), // MUST be the root window
		xFalse,
		xSubstructureNotifyMask|xSubstructureRedirectMask,
		xev,
	)
}

var x11OneByte = make([]byte, 1)

func (w *x11Window) ProcessEvent(e event.Event) {
	w.w.ProcessEvent(e)
}

func (w *x11Window) shutdown(err error) {
	w.ProcessEvent(X11ViewEvent{})
	w.ProcessEvent(DestroyEvent{Err: err})
	w.destroy()
}

func (w *x11Window) Event() event.Event {
	for {
		if w.delivered != nil {
			e := w.delivered
			w.delivered = nil
			if !e.abort {
				w.shutdown(nil)
			}
		}
		evt, ok := w.w.nextEvent()
		if !ok {
			w.dispatch()
			continue
		}
		if e, ok := evt.(*ClosingEvent); ok && e == w.closing {
			w.delivered = w.closing
			w.closing = nil
		}
		return evt
	}
}

func (w *x11Window) Run(f func()) {
	f()
}

func (w *x11Window) Frame(frame *op.Ops) {
	w.w.ProcessFrame(frame, nil)
}

func (w *x11Window) Invalidate() {
	select {
	case w.wakeups <- struct{}{}:
	default:
	}
	w.invalidate()
}

func (w *x11Window) invalidate() {
	if _, err := syscall.Write(w.notify.write, x11OneByte); err != nil && err != syscall.EAGAIN {
		panic(fmt.Errorf("failed to write to pipe: %v", err))
	}
}

func (w *x11Window) display() unsafe.Pointer {
	return w.x
}

func (w *x11Window) window() (xWindow, int, int) {
	return w.xw, w.config.Size.X, w.config.Size.Y
}

func (w *x11Window) dispatch() {
	if w.x == nil {
		// Only Invalidate can wake us up.
		<-w.wakeups
		w.w.Invalidate()
		return
	}

	select {
	case <-w.wakeups:
		w.w.Invalidate()
	default:
	}

dispatchFnsLoop:
	for {
		select {
		case fn := <-w.dispatchFns:
			fn()
		default:
			break dispatchFnsLoop
		}
	}

	xfd := xConnectionNumber(w.x)

	// Poll for events and notifications.
	pollfds := []syscall.PollFd{
		{Fd: int32(xfd), Events: syscall.POLLIN | syscall.POLLERR},
		{Fd: int32(w.notify.read), Events: syscall.POLLIN | syscall.POLLERR},
	}
	xEvents := &pollfds[0].Revents
	// Plenty of room for a backlog of notifications.

	var syn, anim bool
	// Check for pending draw events before checking animation or blocking.
	// This fixes an issue on Xephyr where on startup XPending() > 0 but
	// poll will still block. This also prevents no-op calls to poll.
	syn = w.handler.handleEvents()
	if w.x == nil {
		// handleEvents received a close request and destroyed the window.
		return
	}
	if !syn {
		anim = w.animating
		if !anim {
			// Clear poll events.
			*xEvents = 0
			// Wait for X event or gio notification.
			if _, err := syscall.Poll(pollfds, -1); err != nil && err != syscall.EINTR {
				panic(fmt.Errorf("x11 loop: poll failed: %w", err))
			}
			switch {
			case *xEvents&syscall.POLLIN != 0:
				syn = w.handler.handleEvents()
				if w.x == nil {
					return
				}
			case *xEvents&(syscall.POLLERR|syscall.POLLHUP) != 0:
			}
		}
	}
	// Clear notifications.
	for {
		_, err := syscall.Read(w.notify.read, w.buf[:])
		if err == syscall.EAGAIN {
			break
		}
		if err != nil {
			panic(fmt.Errorf("x11 loop: read from notify pipe failed: %w", err))
		}
	}
	if (anim || syn) && w.config.Size.X != 0 && w.config.Size.Y != 0 {
		w.ProcessEvent(frameEvent{
			FrameEvent: FrameEvent{
				Now:    time.Now(),
				Size:   w.config.Size,
				Metric: w.metric,
			},
			Sync: syn,
		})
	}
}

func (w *x11Window) destroy() {
	if w.notify.write != 0 {
		syscall.Close(w.notify.write)
		w.notify.write = 0
	}
	if w.notify.read != 0 {
		syscall.Close(w.notify.read)
		w.notify.read = 0
	}
	if w.xkb != nil {
		w.xkb.Destroy()
		w.xkb = nil
	}
	w.ibus.Stop()
	xDestroyWindow(w.x, w.xw)
	xCloseDisplay(w.x)
	w.x = nil
}

// atom is a wrapper around XInternAtom. Callers should cache the result
// in order to limit round-trips to the X server.
func (w *x11Window) atom(name string, onlyIfExists bool) xAtom {
	flag := int32(xFalse)
	if onlyIfExists {
		flag = xTrue
	}
	return xInternAtom(w.x, cString(name), flag)
}

// x11EventHandler wraps static variables for the main event loop.
// Its sole purpose is to prevent heap allocation and reduce clutter
// in x11window.loop.
type x11EventHandler struct {
	w    *x11Window
	text []byte
	xev  *xEvent
}

// handleEvents returns true if the window needs to be redrawn.
func (h *x11EventHandler) handleEvents() bool {
	w := h.w
	xev := h.xev
	redraw := false
	for xPending(w.x) != 0 {
		xNextEvent(w.x, xev)
		if xFilterEvent(xev, xNone) == xTrue {
			continue
		}
		switch _type := (*xAnyEvent)(unsafe.Pointer(xev)).Type; _type {
		case h.w.xkbEventBase:
			xkbEvent := (*xkbAnyEvent)(unsafe.Pointer(xev))
			switch xkbEvent.XkbType {
			case xkbNewKeyboardNotify, xkbMapNotify:
				if err := h.w.updateXkbKeymap(); err != nil {
					panic(err)
				}
			case xkbStateNotify:
				state := (*xkbStateNotifyEvent)(unsafe.Pointer(xev))
				h.w.xkb.UpdateMask(state.BaseMods, state.LatchedMods, state.LockedMods,
					uint32(state.BaseGroup), uint32(state.LatchedGroup), uint32(state.LockedGroup))
			}
		case xKeyPress, xKeyRelease:
			ks := key.Press
			if _type == xKeyRelease {
				ks = key.Release
			}
			kevt := (*xKeyEvent)(unsafe.Pointer(xev))
			sym, events := h.w.xkb.DispatchKey(kevt.Keycode, ks)
			handledByIME := false
			if _type == xKeyPress {
				handledByIME = w.ibus.ProcessKey(sym, kevt.Keycode, kevt.State)
				if handledByIME {
					// Update IME cursor position
					sel := w.w.EditorState().Selection
					top := sel.Transform.Transform(sel.Caret.Pos.Add(f32.Pt(0, -sel.Caret.Ascent)))
					bottom := sel.Transform.Transform(sel.Caret.Pos.Add(f32.Pt(0, sel.Caret.Descent)))
					var pos image.Rectangle
					pos.Min = image.Pt(int(top.X+.5), int(top.Y+.5))
					pos.Max = image.Pt(int(bottom.X+.5), int(bottom.Y+.5))
					if pos.Max.Y <= pos.Min.Y {
						pos.Max.Y = pos.Min.Y + 1
					}
					if pos.Max.X <= pos.Min.X {
						pos.Max.X = pos.Min.X + 1
					}
					root := xDefaultRootWindow(w.x)
					dstX, dstY := xTranslateCoordinates(w.x, w.xw, root, 0, 0)
					pos.Min.X += int(dstX)
					pos.Min.Y += int(dstY)
					pos.Max.X += int(dstX)
					pos.Max.Y += int(dstY)
					w.ibus.SetCursorLocation(pos)
				}
			}
			if !handledByIME {
				for _, e := range events {
					if ee, ok := e.(key.EditEvent); ok {
						w.w.EditorInsert(ee.Text)
					} else {
						w.ProcessEvent(e)
					}
				}
			}
		case xButtonPress, xButtonRelease:
			bevt := (*xButtonEvent)(unsafe.Pointer(xev))
			ev := pointer.Event{
				Kind:   pointer.Press,
				Source: pointer.Mouse,
				Position: f32.Point{
					X: float32(bevt.X),
					Y: float32(bevt.Y),
				},
				Time:      time.Duration(bevt.Time) * time.Millisecond,
				Modifiers: w.xkb.Modifiers(),
			}
			if bevt.Type == xButtonRelease {
				ev.Kind = pointer.Release
			}
			var btn pointer.Buttons
			const scrollScale = 10
			switch bevt.Button {
			case xButton1:
				btn = pointer.ButtonPrimary
			case xButton2:
				btn = pointer.ButtonTertiary
			case xButton3:
				btn = pointer.ButtonSecondary
			case xButton4:
				ev.Kind = pointer.Scroll
				// scroll up or left (if shift is pressed).
				if ev.Modifiers == key.ModShift {
					ev.Scroll.X = -scrollScale
				} else {
					ev.Scroll.Y = -scrollScale
				}
			case xButton5:
				// scroll down or right (if shift is pressed).
				ev.Kind = pointer.Scroll
				if ev.Modifiers == key.ModShift {
					ev.Scroll.X = +scrollScale
				} else {
					ev.Scroll.Y = +scrollScale
				}
			case 6:
				// http://xahlee.info/linux/linux_x11_mouse_button_number.html
				// scroll left.
				ev.Kind = pointer.Scroll
				ev.Scroll.X = -scrollScale * 2
			case 7:
				// scroll right
				ev.Kind = pointer.Scroll
				ev.Scroll.X = +scrollScale * 2
			default:
				continue
			}
			switch _type {
			case xButtonPress:
				w.pointerBtns |= btn
			case xButtonRelease:
				w.pointerBtns &^= btn
			}
			ev.Buttons = w.pointerBtns
			w.ProcessEvent(ev)
		case xMotionNotify:
			mevt := (*xMotionEvent)(unsafe.Pointer(xev))
			w.ProcessEvent(pointer.Event{
				Kind:    pointer.Move,
				Source:  pointer.Mouse,
				Buttons: w.pointerBtns,
				Position: f32.Point{
					X: float32(mevt.X),
					Y: float32(mevt.Y),
				},
				Time:      time.Duration(mevt.Time) * time.Millisecond,
				Modifiers: w.xkb.Modifiers(),
			})
		case xExpose: // update
			// redraw only on the last expose event
			redraw = (*xExposeEvent)(unsafe.Pointer(xev)).Count == 0
		case xFocusIn:
			w.config.Focused = true
			w.ProcessEvent(ConfigEvent{Config: w.config})
			w.ibus.Focused(true)
		case xFocusOut:
			w.config.Focused = false
			w.ProcessEvent(ConfigEvent{Config: w.config})
			w.ibus.Focused(false)
		case xConfigureNotify: // window configuration change
			cevt := (*xConfigureEvent)(unsafe.Pointer(xev))
			if sz := image.Pt(int(cevt.Width), int(cevt.Height)); sz != w.config.Size {
				w.config.Size = sz
				w.ProcessEvent(ConfigEvent{Config: w.config})
			}
			// redraw will be done by a later expose event
		case xSelectionNotify:
			cevt := (*xSelectionEvent)(unsafe.Pointer(xev))
			prop := w.atoms.clipboardContent
			if cevt.Property != prop {
				break
			}
			if cevt.Selection != w.atoms.clipboard {
				break
			}
			var text xTextProperty
			if st := xGetTextProperty(w.x, w.xw, &text, prop); st == 0 {
				// Failed; ignore.
				break
			}
			if text.Format != 8 || text.Encoding != w.atoms.utf8string {
				// Ignore non-utf-8 encoded strings.
				break
			}
			str := goStringN(text.Value, int(text.Nitems))
			w.ProcessEvent(transfer.DataEvent{
				Type: "application/text",
				Open: func() io.ReadCloser {
					return io.NopCloser(strings.NewReader(str))
				},
			})
		case xSelectionRequest:
			cevt := (*xSelectionRequestEvent)(unsafe.Pointer(xev))
			if (cevt.Selection != w.atoms.clipboard && cevt.Selection != w.atoms.primary) || cevt.Property == xNone {
				// Unsupported clipboard or obsolete requestor.
				break
			}
			notify := func() {
				xev := new(xEvent)
				ev := (*xSelectionEvent)(unsafe.Pointer(xev))
				*ev = xSelectionEvent{
					Type:      xSelectionNotify,
					Display:   cevt.Display,
					Requestor: cevt.Requestor,
					Selection: cevt.Selection,
					Target:    cevt.Target,
					Property:  cevt.Property,
					Time:      cevt.Time,
				}
				xSendEvent(w.x, cevt.Requestor, 0, 0, xev)
			}
			switch cevt.Target {
			case w.atoms.targets:
				// The requestor wants the supported clipboard
				// formats. First write the targets...
				// Format 32 data is an array of C longs.
				formats := []int{
					int(w.atoms.targets),
					int(w.atoms.utf8string),
					int(w.atoms.plaintext),
					// GTK clients need this.
					int(w.atoms.gtk_text_buffer_contents),
				}
				xChangeProperty(w.x, cevt.Requestor, cevt.Property, w.atoms.atom,
					32 /* bitwidth of formats */, xPropModeReplace,
					unsafe.Pointer(&formats[0]), int32(len(formats)),
				)
				// ...then notify the requestor.
				notify()
			case w.atoms.plaintext, w.atoms.utf8string, w.atoms.gtk_text_buffer_contents:
				content := w.clipboard.content
				xChangeProperty(w.x, cevt.Requestor, cevt.Property, cevt.Target,
					8 /* bitwidth */, xPropModeReplace,
					unsafe.Pointer(unsafe.SliceData(content)), int32(len(content)),
				)
				notify()
			}
		case xClientMessage: // extensions
			cevt := (*xClientMessageEvent)(unsafe.Pointer(xev))
			switch cevt.Data[0] {
			case int(w.atoms.evDelWindow):
				if w.closing != nil {
					return true
				}
				e := &ClosingEvent{}
				w.closing = e
				w.ProcessEvent(e)
				return true
			}
		}
	}
	return redraw
}

var x11Threads sync.Once

func init() {
	x11Driver = newX11Window
}

func newX11Window(gioWin *callbacks, options []Option) error {
	var err error

	pipe := make([]int, 2)
	if err := syscall.Pipe2(pipe, syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		return fmt.Errorf("NewX11Window: failed to create pipe: %w", err)
	}

	x11Threads.Do(func() {
		if err = loadX11(); err != nil {
			return
		}
		if xInitThreads() == 0 {
			err = errors.New("x11: threads init failed")
		}
		xrmInitialize()
	})
	if err != nil {
		return err
	}
	dpy := xOpenDisplay()
	if dpy == nil {
		return errors.New("x11: cannot connect to the X server")
	}
	xkbEventBase, ok := xkbQueryExtension(dpy, xkbMajorVersion, xkbMinorVersion)
	if !ok {
		xCloseDisplay(dpy)
		return errors.New("x11: XkbQueryExtension failed")
	}
	const bits = uint32(xkbNewKeyboardNotifyMask | xkbMapNotifyMask | xkbStateNotifyMask)
	if xkbSelectEvents(dpy, xkbUseCoreKbd, bits, bits) != xTrue {
		xCloseDisplay(dpy)
		return errors.New("x11: XkbSelectEvents failed")
	}
	xkb, err := xkb.New()
	if err != nil {
		xCloseDisplay(dpy)
		return fmt.Errorf("x11: %v", err)
	}

	ppsp := x11DetectUIScale(dpy)
	cfg := unit.Metric{PxPerDp: ppsp, PxPerSp: ppsp}
	// Only use cnf for getting the window size.
	var cnf Config
	cnf.apply(cfg, options)

	swa := xSetWindowAttributes{
		EventMask: xExposureMask | xFocusChangeMask | // update
			xKeyPressMask | xKeyReleaseMask | // keyboard
			xButtonPressMask | xButtonReleaseMask | // mouse clicks
			xPointerMotionMask | // mouse movement
			xStructureNotifyMask, // resize
		BackgroundPixmap: xNone,
		OverrideRedirect: xFalse,
	}
	win := xCreateWindow(dpy, xDefaultRootWindow(dpy),
		0, 0, uint32(cnf.Size.X), uint32(cnf.Size.Y),
		0, xCopyFromParent, xInputOutput,
		xCWEventMask|xCWBackPixmap|xCWOverrideRedirect, &swa)

	w := &x11Window{
		w: gioWin, x: dpy, xw: win,
		metric:       cfg,
		xkb:          xkb,
		xkbEventBase: xkbEventBase,
		wakeups:      make(chan struct{}, 1),
		dispatchFns:  make(chan func(), 20),
		config:       Config{Size: cnf.Size},
	}
	w.handler = x11EventHandler{w: w, xev: new(xEvent), text: make([]byte, 4)}
	w.notify.read = pipe[0]
	w.notify.write = pipe[1]
	w.w.SetDriver(w)

	if err := w.updateXkbKeymap(); err != nil {
		w.destroy()
		return err
	}

	var hints xWMHints
	hints.Input = xTrue
	hints.Flags = xInputHint
	xSetWMHints(dpy, win, &hints)

	name := cString(ID)
	xSetClassHint(dpy, win, name, name)

	w.atoms.utf8string = w.atom("UTF8_STRING", false)
	w.atoms.plaintext = w.atom("text/plain;charset=utf-8", false)
	w.atoms.gtk_text_buffer_contents = w.atom("GTK_TEXT_BUFFER_CONTENTS", false)
	w.atoms.evDelWindow = w.atom("WM_DELETE_WINDOW", false)
	w.atoms.clipboard = w.atom("CLIPBOARD", false)
	w.atoms.primary = w.atom("PRIMARY", false)
	w.atoms.clipboardContent = w.atom("CLIPBOARD_CONTENT", false)
	w.atoms.atom = w.atom("ATOM", false)
	w.atoms.targets = w.atom("TARGETS", false)
	w.atoms.wmName = w.atom("_NET_WM_NAME", false)
	w.atoms.wmState = w.atom("_NET_WM_STATE", false)
	w.atoms.wmStateFullscreen = w.atom("_NET_WM_STATE_FULLSCREEN", false)
	w.atoms.wmActiveWindow = w.atom("_NET_ACTIVE_WINDOW", false)
	w.atoms.wmStateMaximizedHorz = w.atom("_NET_WM_STATE_MAXIMIZED_HORZ", false)
	w.atoms.wmStateMaximizedVert = w.atom("_NET_WM_STATE_MAXIMIZED_VERT", false)

	// extensions
	xSetWMProtocols(dpy, win, []xAtom{w.atoms.evDelWindow})

	// IBus IME
	if strings.Contains(os.Getenv("XMODIFIERS"), "@im=ibus") {
		w.ibus, _ = ibus.Start(xDisplayString(dpy), (*ibusEventHandler)(w))
	}

	// make the window visible on the screen
	xMapWindow(dpy, win)
	w.Configure(options)
	w.ProcessEvent(X11ViewEvent{Display: dpy, Window: uintptr(win)})
	return nil
}

// detectUIScale reports the system UI scale, or 1.0 if it fails.
func x11DetectUIScale(dpy unsafe.Pointer) float32 {
	// default fixed DPI value used in most desktop UI toolkits
	const defaultDesktopDPI = 96
	var scale float32 = 1.0

	// Get actual DPI from X resource Xft.dpi (set by GTK and Qt).
	// This value is entirely based on user preferences and conflates both
	// screen (UI) scaling and font scale.
	rms := xResourceManagerString(dpy)
	if rms != nil {
		db := xrmGetStringDatabase(rms)
		if db != nil {
			if typ, v, ok := xrmGetResource(db, "Xft.dpi", "Xft.Dpi"); ok {
				if typ == "String" {
					f, err := strconv.ParseFloat(goString(v.Addr), 32)
					if err == nil {
						scale = float32(f) / defaultDesktopDPI
					}
				}
			}
			xrmDestroyDatabase(db)
		}
	}

	return scale
}

func (w *x11Window) updateXkbKeymap() error {
	w.xkb.DestroyKeymapState()
	ctx := w.xkb.Ctx
	xcb := xGetXCBConnection(w.x)
	if xcb == nil {
		return errors.New("x11: XGetXCBConnection failed")
	}
	xkbDevID := xkbX11GetCoreKeyboardDeviceID(xcb)
	if xkbDevID == -1 {
		return errors.New("x11: xkb_x11_get_core_keyboard_device_id failed")
	}
	keymap := xkbX11KeymapNewFromDevice(ctx, xcb, xkbDevID)
	if keymap == nil {
		return errors.New("x11: xkb_x11_keymap_new_from_device failed")
	}
	state := xkbX11StateNewFromDevice(keymap, xcb, xkbDevID)
	if state == nil {
		xkbKeymapUnref(keymap)
		return errors.New("x11: xkb_x11_keymap_new_from_device failed")
	}
	w.xkb.SetKeymap(keymap, state)
	return nil
}

type ibusEventHandler x11Window

func (h *ibusEventHandler) Event(new string, sel key.Range, visible, preedit bool) {
	w := (*x11Window)(h)
	handlefn := func() {
		state := w.w.EditorState()
		rng := imeRange(state)
		if rng.Start != rng.End || new != "" {
			w.w.EditorReplace(rng, new)
		}
		end := rng.Start + utf8.RuneCountInString(new)
		if preedit && visible {
			rng.End = end
			w.w.SetComposingRegion(rng)
			if sel.Start < 0 || sel.End < 0 {
				w.w.SetEditorSelection(key.Range{Start: rng.End, End: rng.End})
			} else {
				w.w.SetEditorSelection(key.Range{Start: rng.Start + sel.Start, End: rng.Start + sel.End})
			}
		} else {
			w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
			w.w.SetEditorSelection(key.Range{Start: end, End: end})
		}
	}
	// This function is executed in the dbus event handling goroutine.
	// If we block in this function dbus will not be able to use its connection
	// which means that as soon as we receive a keypress from the X server the
	// dbus call to process it will also block, leading to a deadlock.
	select {
	case w.dispatchFns <- handlefn:
		w.invalidate()
	default:
	}
}
