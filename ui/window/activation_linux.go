//go:build linux && !android

package window

import (
	"sync"
	"unsafe"

	gioapp "gioui.org/app"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"

	"github.com/dyike/keel/ui/internal/loop"
)

func activationWindowEvent(w *Window, e any) {
	switch e := e.(type) {
	case gioapp.X11ViewEvent:
		loop.Lock()
		w.x11Window = uint32(e.Window)
		loop.Unlock()
		w.waylandDisplay.Store(nil)
		w.waylandSurface.Store(nil)
	case gioapp.WaylandViewEvent:
		loop.Lock()
		w.x11Window = 0
		loop.Unlock()
		w.waylandSurface.Store((*byte)(e.Surface))
		w.waylandDisplay.Store((*byte)(e.Display))
	}
}

// x11 is a connection of our own for activation requests, separate from
// Gio's Xlib display so it never contends with its event loop.
var x11 struct {
	sync.Mutex
	conn *xgb.Conn
}

// platformActivate hands the token to the compositor on Wayland, and on X11
// tags the window with the startup ID and asks the window manager to
// activate it, per the startup-notification and EWMH specs.
func platformActivate(w *Window, token string) bool {
	if display, surface := unsafe.Pointer(w.waylandDisplay.Load()), unsafe.Pointer(w.waylandSurface.Load()); display != nil && surface != nil {
		// Wayland: the token goes to xdg-activation with our surface.
		go func() {
			if w.isClosed() || waylandActivate(display, surface, token) != 0 {
				w.Raise()
			}
		}()
		return true
	}
	loop.Lock()
	win := w.x11Window
	loop.Unlock()
	if win == 0 {
		return false
	}
	go func() {
		if err := x11Activate(xproto.Window(win), token); err != nil {
			w.Raise()
		}
	}()
	return true
}

func x11Activate(win xproto.Window, token string) error {
	x11.Lock()
	defer x11.Unlock()
	if x11.conn == nil {
		c, err := xgb.NewConn()
		if err != nil {
			return err
		}
		x11.conn = c
	}
	c := x11.conn
	atom := func(name string) (xproto.Atom, error) {
		r, err := xproto.InternAtom(c, false, uint16(len(name)), name).Reply()
		if err != nil {
			return 0, err
		}
		return r.Atom, nil
	}
	startup, err := atom("_NET_STARTUP_ID")
	if err != nil {
		x11.conn.Close()
		x11.conn = nil
		return err
	}
	utf8, err := atom("UTF8_STRING")
	if err != nil {
		return err
	}
	active, err := atom("_NET_ACTIVE_WINDOW")
	if err != nil {
		return err
	}
	if err := xproto.ChangePropertyChecked(c, xproto.PropModeReplace, win, startup, utf8, 8, uint32(len(token)), []byte(token)).Check(); err != nil {
		return err
	}
	data := activeWindowData(token)
	ev := xproto.ClientMessageEvent{Format: 32, Window: win, Type: active, Data: xproto.ClientMessageDataUnionData32New(data[:])}
	root := xproto.Setup(c).DefaultScreen(c).Root
	mask := uint32(xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify)
	return xproto.SendEventChecked(c, false, root, mask, string(ev.Bytes())).Check()
}
