//go:build linux && !android

package window

import (
	"image"
	"sync"

	gioapp "gioui.org/app"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// X11 windows carry their icon in _NET_WM_ICON. Wayland has no common
// protocol for it yet: the compositor shows the icon of the .desktop entry
// whose name matches the app ID (keel build installs one).

var x11Icons struct {
	sync.Mutex
	windows map[*Window]uint32
}

func platformSetIcon(art image.Image) {
	x11Icons.Lock()
	ids := make([]uint32, 0, len(x11Icons.windows))
	for _, id := range x11Icons.windows {
		ids = append(ids, id)
	}
	x11Icons.Unlock()
	go func() {
		for _, id := range ids {
			setX11Icon(id, art)
		}
	}()
}

func iconWindowEvent(w *Window, e any) {
	var id uint32
	switch e := e.(type) {
	case gioapp.X11ViewEvent:
		id = uint32(e.Window)
	case gioapp.WaylandViewEvent:
	default:
		return
	}
	x11Icons.Lock()
	if x11Icons.windows == nil {
		x11Icons.windows = map[*Window]uint32{}
	}
	if id == 0 {
		delete(x11Icons.windows, w)
		x11Icons.Unlock()
		return
	}
	x11Icons.windows[w] = id
	x11Icons.Unlock()
	if art := currentIcon(); art != nil {
		go setX11Icon(id, art)
	}
}

func setX11Icon(id uint32, art image.Image) {
	data := netWMIcon(art)
	withX11(func(c *xgb.Conn, atom func(string) (xproto.Atom, error)) error {
		prop, err := atom("_NET_WM_ICON")
		if err != nil {
			return err
		}
		return xproto.ChangePropertyChecked(c, xproto.PropModeReplace, xproto.Window(id), prop, xproto.AtomCardinal, 32, uint32(len(data)/4), data).Check()
	})
}
