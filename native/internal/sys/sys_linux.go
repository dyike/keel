//go:build linux && !android

package sys

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"strconv"
	"strings"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"

	"github.com/dyike/keel/native"
)

// Linux support speaks the X11 protocol, so it works on X11 sessions and,
// for X clients only, under XWayland. Wayland itself offers no global
// capture or input injection without portals. Points are pixels divided by
// Xft.dpi/96, the scale Gio uses on X11.

var x struct {
	sync.Mutex
	conn   *xgb.Conn
	root   xproto.Window
	scale  float64
	xtest  bool
	randr  bool
	keymap map[xproto.Keysym]xproto.Keycode
}

// display opens the shared connection on first use.
func display() (*xgb.Conn, error) {
	x.Lock()
	defer x.Unlock()
	if x.conn != nil {
		return x.conn, nil
	}
	c, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("%w: no X11 display: %v", native.ErrUnsupported, err)
	}
	x.conn, x.root = c, xproto.Setup(c).DefaultScreen(c).Root
	x.xtest = xtest.Init(c) == nil
	x.randr = randr.Init(c) == nil
	x.scale = xftScale(c, x.root)
	return c, nil
}

func xftScale(c *xgb.Conn, root xproto.Window) float64 {
	const name = "RESOURCE_MANAGER"
	a, err := xproto.InternAtom(c, true, uint16(len(name)), name).Reply()
	if err != nil || a.Atom == 0 {
		return 1
	}
	p, err := xproto.GetProperty(c, false, root, a.Atom, xproto.AtomString, 0, 1<<16).Reply()
	if err != nil {
		return 1
	}
	for line := range strings.SplitSeq(string(p.Value), "\n") {
		if v, ok := strings.CutPrefix(line, "Xft.dpi:"); ok {
			if dpi, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && dpi > 0 {
				return dpi / 96
			}
		}
	}
	return 1
}

func failed(op string, err error) error { return fmt.Errorf("%w: %s: %v", native.ErrFailed, op, err) }

func Permission(kind int, request bool) (bool, error) { return true, nil }

type monitor struct {
	x, y, w, h int
	primary    bool
}

func monitors(c *xgb.Conn) ([]monitor, error) {
	if x.randr {
		if r, err := randr.GetMonitors(c, x.root, true).Reply(); err == nil && len(r.Monitors) > 0 {
			out := make([]monitor, len(r.Monitors))
			for i, m := range r.Monitors {
				out[i] = monitor{int(m.X), int(m.Y), int(m.Width), int(m.Height), m.Primary}
			}
			if !anyPrimary(out) {
				out[0].primary = true
			}
			return out, nil
		}
	}
	s := xproto.Setup(c).DefaultScreen(c)
	return []monitor{{0, 0, int(s.WidthInPixels), int(s.HeightInPixels), true}}, nil
}

func anyPrimary(ms []monitor) bool {
	for _, m := range ms {
		if m.primary {
			return true
		}
	}
	return false
}

func Displays() ([]Display, error) {
	c, err := display()
	if err != nil {
		return nil, err
	}
	ms, err := monitors(c)
	if err != nil {
		return nil, err
	}
	// Report the origin relative to the primary display, as on macOS.
	var ox, oy int
	for _, m := range ms {
		if m.primary {
			ox, oy = m.x, m.y
		}
	}
	s := x.scale
	out := make([]Display, len(ms))
	for i, m := range ms {
		out[i] = Display{ID: uint32(i + 1), X: float64(m.x-ox) / s, Y: float64(m.y-oy) / s,
			Width: float64(m.w) / s, Height: float64(m.h) / s, PixelWidth: m.w, PixelHeight: m.h, Primary: m.primary}
	}
	return out, nil
}

// Capture reads the display's pixels from the root window and encodes them
// as PNG. It handles the 24- and 32-bit true color visuals X servers use.
func Capture(id uint32) ([]byte, error) {
	c, err := display()
	if err != nil {
		return nil, err
	}
	ms, err := monitors(c)
	if err != nil {
		return nil, err
	}
	if id == 0 || int(id) > len(ms) {
		return nil, native.ErrInvalidArgument
	}
	m := ms[id-1]
	r, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(x.root),
		int16(m.x), int16(m.y), uint16(m.w), uint16(m.h), 0xffffffff).Reply()
	if err != nil {
		return nil, failed("GetImage", err)
	}
	if len(r.Data) < m.w*m.h*4 {
		return nil, fmt.Errorf("%w: unsupported pixel format (depth %d)", native.ErrFailed, r.Depth)
	}
	msb := xproto.Setup(c).ImageByteOrder == xproto.ImageOrderMSBFirst
	img := image.NewNRGBA(image.Rect(0, 0, m.w, m.h))
	for i := 0; i < m.w*m.h*4; i += 4 {
		p := r.Data[i : i+4]
		if msb { // XRGB
			img.Pix[i], img.Pix[i+1], img.Pix[i+2] = p[1], p[2], p[3]
		} else { // BGRX
			img.Pix[i], img.Pix[i+1], img.Pix[i+2] = p[2], p[1], p[0]
		}
		img.Pix[i+3] = 0xff
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, failed("png", err)
	}
	return out.Bytes(), nil
}

// primaryOrigin is the primary display's top left in root pixels.
func primaryOrigin(c *xgb.Conn) (int, int) {
	ms, _ := monitors(c)
	for _, m := range ms {
		if m.primary {
			return m.x, m.y
		}
	}
	return 0, 0
}

func MousePosition() (float64, float64, error) {
	c, err := display()
	if err != nil {
		return 0, 0, err
	}
	r, err := xproto.QueryPointer(c, x.root).Reply()
	if err != nil {
		return 0, 0, failed("QueryPointer", err)
	}
	ox, oy := primaryOrigin(c)
	return float64(int(r.RootX)-ox) / x.scale, float64(int(r.RootY)-oy) / x.scale, nil
}

// withXTest returns the connection if the server has the XTEST extension,
// which synthetic input needs.
func withXTest() (*xgb.Conn, error) {
	c, err := display()
	if err != nil {
		return nil, err
	}
	if !x.xtest {
		return nil, fmt.Errorf("%w: the X server lacks the XTEST extension", native.ErrUnsupported)
	}
	return c, nil
}

func fake(c *xgb.Conn, typ, detail byte, rx, ry int16) error {
	if err := xtest.FakeInputChecked(c, typ, detail, 0, x.root, rx, ry, 0).Check(); err != nil {
		return failed("FakeInput", err)
	}
	return nil
}

func MouseMove(px, py float64) error {
	c, err := withXTest()
	if err != nil {
		return err
	}
	ox, oy := primaryOrigin(c)
	return fake(c, xproto.MotionNotify, 0, int16(px*x.scale)+int16(ox), int16(py*x.scale)+int16(oy))
}

func Click(button int) error {
	detail := [...]byte{1, 3, 2} // left, right, middle
	if button < 0 || button >= len(detail) {
		return native.ErrInvalidArgument
	}
	c, err := withXTest()
	if err != nil {
		return err
	}
	if err := fake(c, xproto.ButtonPress, detail[button], 0, 0); err != nil {
		return err
	}
	return fake(c, xproto.ButtonRelease, detail[button], 0, 0)
}

// keysyms maps Keel key names to X keysyms, the same names as on macOS.
var keysyms = func() map[string]xproto.Keysym {
	m := map[string]xproto.Keysym{
		"enter": 0xff0d, "tab": 0xff09, "space": 0x20, "backspace": 0xff08, "escape": 0xff1b,
		"shift": 0xffe1, "ctrl": 0xffe3, "alt": 0xffe9, "cmd": 0xffeb,
		"home": 0xff50, "left": 0xff51, "up": 0xff52, "right": 0xff53, "down": 0xff54,
		"pageup": 0xff55, "pagedown": 0xff56, "end": 0xff57, "delete": 0xffff,
	}
	for _, r := range "abcdefghijklmnopqrstuvwxyz0123456789;=,-./`[\\]'" {
		m[string(r)] = xproto.Keysym(r)
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("f%d", i)] = xproto.Keysym(0xffbe + i - 1)
	}
	return m
}()

// keycode finds the key that types a keysym in the current keyboard layout.
func keycode(c *xgb.Conn, name string) (xproto.Keycode, error) {
	sym, ok := keysyms[name]
	if !ok {
		return 0, fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, name)
	}
	x.Lock()
	defer x.Unlock()
	if x.keymap == nil {
		setup := xproto.Setup(c)
		lo, hi := setup.MinKeycode, setup.MaxKeycode
		r, err := xproto.GetKeyboardMapping(c, lo, byte(hi-lo+1)).Reply()
		if err != nil {
			return 0, failed("GetKeyboardMapping", err)
		}
		x.keymap = map[xproto.Keysym]xproto.Keycode{}
		per := int(r.KeysymsPerKeycode)
		for i := len(r.Keysyms)/per - 1; i >= 0; i-- { // lowest keycode wins
			for j := range min(per, 2) { // unshifted and shifted
				if s := r.Keysyms[i*per+j]; s != 0 {
					x.keymap[s] = lo + xproto.Keycode(i)
				}
			}
		}
	}
	code, ok := x.keymap[sym]
	if !ok {
		return 0, fmt.Errorf("%w: key %q is not on this keyboard layout", native.ErrFailed, name)
	}
	return code, nil
}

func HasKey(key string) bool { _, ok := keysyms[key]; return ok }

func Key(key string, down bool) error {
	c, err := withXTest()
	if err != nil {
		if !HasKey(key) {
			return fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
		}
		return err
	}
	code, err := keycode(c, key)
	if err != nil {
		return err
	}
	typ := byte(xproto.KeyRelease)
	if down {
		typ = xproto.KeyPress
	}
	return fake(c, typ, byte(code), 0, 0)
}

// Hotkeys grab keys on the root window over a second connection, whose
// only reader is the event loop below.
var hk struct {
	sync.Mutex
	conn  *xgb.Conn
	root  xproto.Window
	fired map[hotkeyID]chan struct{}
}

type hotkeyID struct {
	code xproto.Keycode
	mods uint16
}

// Lock and NumLock (Mod2) must not stop a hotkey, so each is grabbed with
// every combination of them.
const ignoredMods = xproto.ModMaskLock | xproto.ModMask2

var lockVariants = [...]uint16{0, xproto.ModMaskLock, xproto.ModMask2, xproto.ModMaskLock | xproto.ModMask2}

func hotkeyConn() (*xgb.Conn, error) {
	hk.Lock()
	defer hk.Unlock()
	if hk.conn != nil {
		return hk.conn, nil
	}
	c, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("%w: no X11 display: %v", native.ErrUnsupported, err)
	}
	hk.conn, hk.root, hk.fired = c, xproto.Setup(c).DefaultScreen(c).Root, map[hotkeyID]chan struct{}{}
	go func() {
		for {
			ev, err := c.WaitForEvent()
			if ev == nil && err == nil {
				return // connection closed
			}
			if k, ok := ev.(xproto.KeyPressEvent); ok {
				hk.Lock()
				if ch := hk.fired[hotkeyID{k.Detail, k.State &^ ignoredMods}]; ch != nil {
					select {
					case ch <- struct{}{}:
					default:
					}
				}
				hk.Unlock()
			}
		}
	}()
	return c, nil
}

// Hotkey grabs key+mods; fn runs on its own goroutine and presses that
// arrive while it runs are merged.
func Hotkey(key string, mods uint, fn func()) (func() error, error) {
	if !HasKey(key) {
		return nil, fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	c, err := hotkeyConn()
	if err != nil {
		return nil, err
	}
	code, err := keycode(c, key)
	if err != nil {
		return nil, err
	}
	var m uint16
	if mods&ModShift != 0 {
		m |= xproto.ModMaskShift
	}
	if mods&ModCtrl != 0 {
		m |= xproto.ModMaskControl
	}
	if mods&ModAlt != 0 {
		m |= xproto.ModMask1
	}
	if mods&ModCmd != 0 {
		m |= xproto.ModMask4 // Super
	}
	id := hotkeyID{code, m}
	hk.Lock()
	if hk.fired[id] != nil {
		hk.Unlock()
		return nil, native.ErrConflict
	}
	fired := make(chan struct{}, 1)
	hk.fired[id] = fired
	hk.Unlock()

	ungrab := func() {
		for _, v := range lockVariants {
			xproto.UngrabKey(c, code, hk.root, m|v)
		}
	}
	for _, v := range lockVariants {
		err := xproto.GrabKeyChecked(c, true, hk.root, m|v, code, xproto.GrabModeAsync, xproto.GrabModeAsync).Check()
		if err != nil {
			ungrab()
			hk.Lock()
			delete(hk.fired, id)
			hk.Unlock()
			var access xproto.AccessError
			if errors.As(err, &access) {
				return nil, native.ErrConflict // another client grabbed it
			}
			return nil, failed("GrabKey", err)
		}
	}
	go func() {
		for range fired {
			fn()
		}
	}()
	var once sync.Once
	return func() error {
		once.Do(func() {
			ungrab()
			hk.Lock()
			delete(hk.fired, id)
			hk.Unlock()
			close(fired)
		})
		return nil
	}, nil
}
