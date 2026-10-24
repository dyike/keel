// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unsafe"

	syscall "golang.org/x/sys/unix"

	"github.com/dyike/keel/gio/app/internal/xkb"
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/internal/fling"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
)

// The libwayland calls live in os_wayland_cgo.go (with os_wayland.c and the
// scanner-generated protocol code) and os_wayland_purego.go. Keel patch:
// this file holds the logic both share. Requests are marshalled in
// wayland_requests.go; each binding delivers events to the wlOn* handlers.

// Use wayland-scanner to generate glue code for the xdg-shell and xdg-decoration extensions.
//go:generate wayland-scanner client-header /usr/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wayland_xdg_shell.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wayland_xdg_shell.c

//go:generate wayland-scanner client-header /usr/share/wayland-protocols/unstable/text-input/text-input-unstable-v3.xml wayland_text_input.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/unstable/text-input/text-input-unstable-v3.xml wayland_text_input.c

//go:generate wayland-scanner client-header /usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml wayland_xdg_decoration.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml wayland_xdg_decoration.c

//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n\\n;" wayland_xdg_shell.c
//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n\\n;" wayland_xdg_decoration.c
//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n\\n;" wayland_text_input.c

type wlDisplay struct {
	disp              unsafe.Pointer
	reg               unsafe.Pointer
	compositor        unsafe.Pointer
	wm                unsafe.Pointer
	imm               unsafe.Pointer
	shm               unsafe.Pointer
	dataDeviceManager unsafe.Pointer
	decor             unsafe.Pointer
	seat              *wlSeat
	xkb               *xkb.Context
	outputMap         map[uint32]unsafe.Pointer
	outputConfig      map[unsafe.Pointer]*wlOutput

	// Notification pipe fds.
	notify struct {
		read, write int
	}

	repeat        repeatState
	poller        poller
	readClipClose chan struct{}
}

type wlSeat struct {
	disp     *wlDisplay
	seat     unsafe.Pointer
	name     uint32
	pointer  unsafe.Pointer
	touch    unsafe.Pointer
	keyboard unsafe.Pointer
	im       unsafe.Pointer
	imFocus  *window
	imActive bool
	imHint   key.InputHint
	imState  textInputState

	// The most recent input serial.
	serial uint32
	// The most recent pointer enter serial.
	pointerSerial uint32

	pointerFocus  *window
	keyboardFocus *window
	touchFoci     map[int32]*window

	// Clipboard support.
	dataDev unsafe.Pointer
	// offers is a map from active wl_data_offers to
	// the list of mime types they support.
	offers map[unsafe.Pointer][]string
	// clipboard is the wl_data_offer for the clipboard.
	clipboard unsafe.Pointer
	// mimeType is the chosen mime type of clipboard.
	mimeType string
	// source represents the clipboard content of the most recent
	// clipboard write, if any.
	source unsafe.Pointer
	// content is the data belonging to source.
	content []byte
}

type repeatState struct {
	rate  int
	delay time.Duration

	key   uint32
	win   *window
	stopC chan struct{}

	start time.Duration
	last  time.Duration
	mu    sync.Mutex
	now   time.Duration
}

type window struct {
	w          *callbacks
	disp       *wlDisplay
	surf       unsafe.Pointer
	wmSurf     unsafe.Pointer
	topLvl     unsafe.Pointer
	decor      unsafe.Pointer
	ppdp, ppsp float32
	scroll     struct {
		time  time.Duration
		steps image.Point
		dist  f32.Point
	}
	pointerBtns pointer.Buttons
	lastPos     f32.Point
	lastTouch   f32.Point
	textInput   bool
	inputHint   key.InputHint

	cursor struct {
		theme  unsafe.Pointer
		cursor *wlCursor
		// system is the active cursor for system gestures
		// such as border resizes and window moves. It
		// is nil if the pointer is not in a system gesture
		// area.
		system  *wlCursor
		surf    unsafe.Pointer
		cursors struct {
			pointer         *wlCursor
			resizeNorth     *wlCursor
			resizeSouth     *wlCursor
			resizeWest      *wlCursor
			resizeEast      *wlCursor
			resizeNorthWest *wlCursor
			resizeNorthEast *wlCursor
			resizeSouthWest *wlCursor
			resizeSouthEast *wlCursor
		}
	}

	fling struct {
		yExtrapolation fling.Extrapolation
		xExtrapolation fling.Extrapolation
		anim           fling.Animation
		start          bool
		dir            f32.Point
	}

	configured        bool
	lastFrameCallback unsafe.Pointer

	animating bool
	// The most recent configure serial waiting to be ack'ed.
	serial uint32
	scale  int
	// size is the unscaled window size (unlike config.Size which is scaled).
	size         image.Point
	config       Config
	wsize        image.Point // window config size before going fullscreen or maximized
	inCompositor bool        // window is moving or being resized

	clipReads chan transfer.DataEvent

	wakeups   chan struct{}
	runFuncs  eventFuncs
	runWakeup func()
	done      chan struct{}

	closing    bool
	teardown   deferredTeardown
	closeEvent *ClosingEvent
	delivered  *ClosingEvent
}

type poller struct {
	pollfds [2]syscall.PollFd
	// buf is scratch space for draining the notification pipe.
	buf [100]byte
}

type wlOutput struct {
	width      int
	height     int
	physWidth  int
	physHeight int
	transform  int32
	scale      int
	windows    []*window
}

// callbackMap maps Wayland native handles to corresponding Go
// references. It is necessary because the Wayland client API
// forces the use of callbacks and storing pointers to Go values
// in C is forbidden.
var callbackMap sync.Map

// clipboardMimeTypes is a list of supported clipboard mime types, in
// order of preference.
var clipboardMimeTypes = []string{"text/plain;charset=utf8", "UTF8_STRING", "text/plain", "TEXT", "STRING"}

var (
	newWaylandEGLContext    func(w *window) (context, error)
	newWaylandVulkanContext func(w *window) (context, error)
)

func init() {
	wlDriver = newWLWindow
}

func newWLWindow(callbacks *callbacks, options []Option) error {
	d, err := newWLDisplay()
	if err != nil {
		return err
	}
	w, err := d.createNativeWindow(options)
	if err != nil {
		d.destroy()
		return err
	}
	w.runWakeup = d.wakeup
	w.w = callbacks
	w.w.SetDriver(w)

	// Finish and commit setup from createNativeWindow.
	w.Configure(options)
	w.draw(true)
	wlSurfaceCommit(w.surf)

	w.ProcessEvent(WaylandViewEvent{
		Display:  unsafe.Pointer(w.display()),
		Surface:  unsafe.Pointer(w.surf),
		Toplevel: unsafe.Pointer(w.topLvl),
	})
	return nil
}

func (d *wlDisplay) writeClipboard(content []byte) error {
	s := d.seat
	if s == nil {
		return nil
	}
	// Clear old offer.
	if s.source != nil {
		wlDataSourceDestroy(s.source)
		s.source = nil
		s.content = nil
	}
	if d.dataDeviceManager == nil || s.dataDev == nil {
		return nil
	}
	s.content = content
	s.source = wlDataDeviceManagerCreateDataSource(d.dataDeviceManager)
	wlAddListener(s.source, lk_data_source, unsafe.Pointer(s.seat))
	for _, mime := range clipboardMimeTypes {
		wlDataSourceOffer(s.source, mime)
	}
	wlDataDeviceSetSelection(s.dataDev, s.source, s.serial)
	return nil
}

func (d *wlDisplay) readClipboard() (io.ReadCloser, error) {
	s := d.seat
	if s == nil {
		return nil, nil
	}
	if s.clipboard == nil {
		return nil, nil
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	// wl_data_offer_receive performs and implicit dup(2) of the write end
	// of the pipe. Close our version.
	defer w.Close()
	wlDataOfferReceive(s.clipboard, s.mimeType, int32(w.Fd()))
	return r, nil
}

func (d *wlDisplay) createNativeWindow(options []Option) (*window, error) {
	if d.compositor == nil {
		return nil, errors.New("wayland: no compositor available")
	}
	if d.wm == nil {
		return nil, errors.New("wayland: no xdg_wm_base available")
	}
	if d.shm == nil {
		return nil, errors.New("wayland: no wl_shm available")
	}
	if len(d.outputMap) == 0 {
		return nil, errors.New("wayland: no outputs available")
	}
	var scale int
	for _, conf := range d.outputConfig {
		if s := conf.scale; s > scale {
			scale = s
		}
	}
	ppdp := detectUIScale()

	w := &window{
		disp:      d,
		scale:     scale,
		ppdp:      ppdp,
		ppsp:      ppdp,
		wakeups:   make(chan struct{}, 1),
		done:      make(chan struct{}),
		clipReads: make(chan transfer.DataEvent, 1),
	}
	w.surf = wlCompositorCreateSurface(d.compositor)
	if w.surf == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_compositor_create_surface failed")
	}
	wlSurfaceSetBufferScale(w.surf, int32(w.scale))
	callbackStore(unsafe.Pointer(w.surf), w)
	w.wmSurf = xdgWmBaseGetXdgSurface(d.wm, w.surf)
	if w.wmSurf == nil {
		w.destroy()
		return nil, errors.New("wayland: xdg_wm_base_get_xdg_surface failed")
	}
	w.topLvl = xdgSurfaceGetToplevel(w.wmSurf)
	if w.topLvl == nil {
		w.destroy()
		return nil, errors.New("wayland: xdg_surface_get_toplevel failed")
	}

	xdgToplevelSetAppId(w.topLvl, ID)

	cursorTheme := os.Getenv("XCURSOR_THEME")
	cursorSize := 32
	if envSize, ok := os.LookupEnv("XCURSOR_SIZE"); ok && envSize != "" {
		size, err := strconv.Atoi(envSize)
		if err == nil {
			cursorSize = size
		}
	}

	w.cursor.theme = wlCursorThemeLoad(cursorTheme, cursorSize*w.scale, d.shm)
	if w.cursor.theme == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_cursor_theme_load failed")
	}
	w.loadCursors()
	w.cursor.cursor = w.cursor.cursors.pointer
	if w.cursor.cursor == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_cursor_theme_get_cursor failed")
	}
	w.cursor.surf = wlCompositorCreateSurface(d.compositor)
	if w.cursor.surf == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_compositor_create_surface failed")
	}
	wlSurfaceSetBufferScale(w.cursor.surf, int32(w.scale))
	wlAddListener(d.wm, lk_xdg_wm_base, unsafe.Pointer(w.surf))
	wlAddListener(w.surf, lk_surface, unsafe.Pointer(w.surf))
	wlAddListener(w.wmSurf, lk_xdg_surface, unsafe.Pointer(w.surf))
	wlAddListener(w.topLvl, lk_xdg_toplevel, unsafe.Pointer(w.surf))

	if d.decor != nil {
		w.decor = zxdgDecorationManagerV1GetToplevelDecoration(d.decor, w.topLvl)
		wlAddListener(w.decor, lk_zxdg_toplevel_decoration_v1, unsafe.Pointer(w.surf))
	}
	w.updateOpaqueRegion()
	return w, nil
}

func (w *window) loadCursors() {
	w.cursor.cursors.pointer = w.loadCursor(pointer.CursorDefault)
	w.cursor.cursors.resizeNorth = w.loadCursor(pointer.CursorNorthResize)
	w.cursor.cursors.resizeSouth = w.loadCursor(pointer.CursorSouthResize)
	w.cursor.cursors.resizeWest = w.loadCursor(pointer.CursorWestResize)
	w.cursor.cursors.resizeEast = w.loadCursor(pointer.CursorEastResize)
	w.cursor.cursors.resizeSouthWest = w.loadCursor(pointer.CursorSouthWestResize)
	w.cursor.cursors.resizeSouthEast = w.loadCursor(pointer.CursorSouthEastResize)
	w.cursor.cursors.resizeNorthWest = w.loadCursor(pointer.CursorNorthWestResize)
	w.cursor.cursors.resizeNorthEast = w.loadCursor(pointer.CursorNorthEastResize)
}

func (w *window) loadCursor(name pointer.Cursor) *wlCursor {
	if name == pointer.CursorNone {
		return nil
	}
	xcursor := xCursor[name]
	c := wlCursorThemeGetCursor(w.cursor.theme, xcursor)
	if c == nil {
		// Fall back to default cursor.
		c = w.cursor.cursors.pointer
	}
	return c
}

func callbackDelete(k unsafe.Pointer) {
	callbackMap.Delete(k)
}

func callbackStore(k unsafe.Pointer, v interface{}) {
	callbackMap.Store(k, v)
}

func callbackLoad(k unsafe.Pointer) interface{} {
	v, exists := callbackMap.Load(k)
	if !exists {
		panic("missing callback entry")
	}
	return v
}

func wlOnSeatCapabilities(data unsafe.Pointer, seat unsafe.Pointer, caps uint32) {
	s := callbackLoad(data).(*wlSeat)
	s.updateCaps(caps)
}

// flushOffers remove all wl_data_offers that isn't the clipboard
// content.
func (s *wlSeat) flushOffers() {
	for o := range s.offers {
		if o == s.clipboard {
			continue
		}
		// We're only interested in clipboard offers.
		delete(s.offers, o)
		callbackDelete(unsafe.Pointer(o))
		wlDataOfferDestroy(o)
	}
}

func (s *wlSeat) destroy() {
	if s.source != nil {
		wlDataSourceDestroy(s.source)
		s.source = nil
	}
	if s.im != nil {
		zwpTextInputV3Destroy(s.im)
		s.im = nil
	}
	s.imFocus = nil
	s.resetTextInputState()
	if s.pointer != nil {
		wlPointerRelease(s.pointer)
	}
	if s.touch != nil {
		wlTouchRelease(s.touch)
	}
	if s.keyboard != nil {
		wlKeyboardRelease(s.keyboard)
	}
	s.clipboard = nil
	s.flushOffers()
	if s.dataDev != nil {
		wlDataDeviceRelease(s.dataDev)
	}
	if s.seat != nil {
		callbackDelete(unsafe.Pointer(s.seat))
		wlSeatRelease(s.seat)
	}
}

func (s *wlSeat) initTextInput() {
	if s.im == nil && s.disp.imm != nil {
		s.im = zwpTextInputManagerV3GetTextInput(s.disp.imm, s.seat)
		wlAddListener(s.im, lk_zwp_text_input_v3, unsafe.Pointer(s.seat))
	}
}

func (s *wlSeat) commitTextInput() {
	zwpTextInputV3Commit(s.im)
	s.imState.serial++
}

func (s *wlSeat) resetTextInputState() {
	s.imActive = false
	s.imState.reset()
}

func (s *wlSeat) updateCaps(caps uint32) {
	s.initTextInput()
	switch {
	case s.pointer == nil && caps&_WL_SEAT_CAPABILITY_POINTER != 0:
		s.pointer = wlSeatGetPointer(s.seat)
		wlAddListener(s.pointer, lk_pointer, unsafe.Pointer(s.seat))
	case s.pointer != nil && caps&_WL_SEAT_CAPABILITY_POINTER == 0:
		wlPointerRelease(s.pointer)
		s.pointer = nil
	}
	switch {
	case s.touch == nil && caps&_WL_SEAT_CAPABILITY_TOUCH != 0:
		s.touch = wlSeatGetTouch(s.seat)
		wlAddListener(s.touch, lk_touch, unsafe.Pointer(s.seat))
	case s.touch != nil && caps&_WL_SEAT_CAPABILITY_TOUCH == 0:
		wlTouchRelease(s.touch)
		s.touch = nil
	}
	switch {
	case s.keyboard == nil && caps&_WL_SEAT_CAPABILITY_KEYBOARD != 0:
		s.keyboard = wlSeatGetKeyboard(s.seat)
		wlAddListener(s.keyboard, lk_keyboard, unsafe.Pointer(s.seat))
	case s.keyboard != nil && caps&_WL_SEAT_CAPABILITY_KEYBOARD == 0:
		wlKeyboardRelease(s.keyboard)
		s.keyboard = nil
	}
}

func wlOnSeatName(data unsafe.Pointer, seat unsafe.Pointer, name string) {
}

func wlOnXdgSurfaceConfigure(data unsafe.Pointer, wmSurf unsafe.Pointer, serial uint32) {
	w := callbackLoad(data).(*window)
	w.serial = serial
	xdgSurfaceAckConfigure(wmSurf, serial)
	w.configured = true
	w.draw(true)
}

func wlOnToplevelClose(data unsafe.Pointer, topLvl unsafe.Pointer) {
	w := callbackLoad(data).(*window)
	w.requestClose()
}

func wlOnToplevelConfigure(data unsafe.Pointer, topLvl unsafe.Pointer, width, height int32, states unsafe.Pointer) {
	w := callbackLoad(data).(*window)
	if width != 0 && height != 0 {
		w.size = image.Pt(int(width), int(height))
		w.updateOpaqueRegion()
	}
}

func wlOnToplevelDecorationConfigure(data unsafe.Pointer, deco unsafe.Pointer, mode uint32) {
	w := callbackLoad(data).(*window)
	decorated := w.config.Decorated
	switch mode {
	case _ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE:
		w.config.Decorated = false
	case _ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE:
		w.config.Decorated = true
	}
	if decorated != w.config.Decorated {
		w.setWindowConstraints()
		if w.config.Decorated {
			w.size.Y -= int(w.config.decoHeight)
		} else {
			w.size.Y += int(w.config.decoHeight)
		}
		w.ProcessEvent(ConfigEvent{Config: w.config})
		w.draw(true)
	}
}

func wlOnOutputMode(data unsafe.Pointer, output unsafe.Pointer, flags uint32, width, height, refresh int32) {
	if flags&_WL_OUTPUT_MODE_CURRENT == 0 {
		return
	}
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.width = int(width)
	c.height = int(height)
}

func wlOnOutputGeometry(data unsafe.Pointer, output unsafe.Pointer, x, y, physWidth, physHeight, subpixel int32, make, model string, transform int32) {
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.transform = transform
	c.physWidth = int(physWidth)
	c.physHeight = int(physHeight)
}

func wlOnOutputScale(data unsafe.Pointer, output unsafe.Pointer, scale int32) {
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.scale = int(scale)
}

func wlOnOutputDone(data unsafe.Pointer, output unsafe.Pointer) {
	d := callbackLoad(data).(*wlDisplay)
	conf := d.outputConfig[output]
	for _, w := range conf.windows {
		w.updateOutputs()
	}
}

func wlOnSurfaceEnter(data unsafe.Pointer, surf unsafe.Pointer, output unsafe.Pointer) {
	w := callbackLoad(data).(*window)
	conf := w.disp.outputConfig[output]
	var found bool
	for _, w2 := range conf.windows {
		if w2 == w {
			found = true
			break
		}
	}
	if !found {
		conf.windows = append(conf.windows, w)
	}
	w.updateOutputs()
	if w.config.Mode == Minimized {
		// Minimized window got brought back up: it is no longer so.
		w.config.Mode = Windowed
		w.ProcessEvent(ConfigEvent{Config: w.config})
	}
}

func wlOnSurfaceLeave(data unsafe.Pointer, surf unsafe.Pointer, output unsafe.Pointer) {
	w := callbackLoad(data).(*window)
	conf := w.disp.outputConfig[output]
	for i, w2 := range conf.windows {
		if w2 == w {
			conf.windows = append(conf.windows[:i], conf.windows[i+1:]...)
			break
		}
	}
	w.updateOutputs()
}

func wlOnRegistryGlobal(data unsafe.Pointer, reg unsafe.Pointer, name uint32, intf string, version uint32) {
	d := callbackLoad(data).(*wlDisplay)
	switch intf {
	case "wl_compositor":
		d.compositor = wlRegistryBind(reg, name, wlIface.wl_compositor, "wl_compositor", 3)
	case "wl_output":
		output := wlRegistryBind(reg, name, wlIface.wl_output, "wl_output", 2)
		wlAddListener(output, lk_output, unsafe.Pointer(d.disp))
		d.outputMap[name] = output
		d.outputConfig[output] = new(wlOutput)
	case "wl_seat":
		if d.seat != nil {
			break
		}
		s := wlRegistryBind(reg, name, wlIface.wl_seat, "wl_seat", 5)
		if s == nil {
			// No support for v5 protocol.
			break
		}
		d.seat = &wlSeat{
			disp:      d,
			name:      name,
			seat:      s,
			offers:    make(map[unsafe.Pointer][]string),
			touchFoci: make(map[int32]*window),
		}
		callbackStore(unsafe.Pointer(s), d.seat)
		wlAddListener(s, lk_seat, unsafe.Pointer(s))
		d.seat.initTextInput()
		d.bindDataDevice()
	case "wl_shm":
		d.shm = wlRegistryBind(reg, name, wlIface.wl_shm, "wl_shm", 1)
	case "xdg_wm_base":
		d.wm = wlRegistryBind(reg, name, wlIface.xdg_wm_base, "xdg_wm_base", 1)
	case "zxdg_decoration_manager_v1":
		d.decor = wlRegistryBind(reg, name, wlIface.zxdg_decoration_manager_v1, "zxdg_decoration_manager_v1", 1)
	case "zwp_text_input_manager_v3":
		d.imm = wlRegistryBind(reg, name, wlIface.zwp_text_input_manager_v3, "zwp_text_input_manager_v3", 1)
		if d.seat != nil {
			d.seat.initTextInput()
		}
	case "wl_data_device_manager":
		d.dataDeviceManager = wlRegistryBind(reg, name, wlIface.wl_data_device_manager, "wl_data_device_manager", 3)
		d.bindDataDevice()
	}
}

func wlOnDataOfferOffer(data unsafe.Pointer, offer unsafe.Pointer, mime string) {
	s := callbackLoad(data).(*wlSeat)
	s.offers[offer] = append(s.offers[offer], mime)
}

func wlOnDataOfferSourceActions(data unsafe.Pointer, offer unsafe.Pointer, acts uint32) {
}

func wlOnDataOfferAction(data unsafe.Pointer, offer unsafe.Pointer, act uint32) {
}

func wlOnDataDeviceOffer(data unsafe.Pointer, dataDev unsafe.Pointer, id unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	callbackStore(unsafe.Pointer(id), s)
	wlAddListener(id, lk_data_offer, unsafe.Pointer(id))
	s.offers[id] = nil
}

func wlOnDataDeviceEnter(data unsafe.Pointer, dataDev unsafe.Pointer, serial uint32, surf unsafe.Pointer, x, y int32, id unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.flushOffers()
}

func wlOnDataDeviceLeave(data unsafe.Pointer, dataDev unsafe.Pointer) {
}

func wlOnDataDeviceMotion(data unsafe.Pointer, dataDev unsafe.Pointer, t uint32, x, y int32) {
}

func wlOnDataDeviceDrop(data unsafe.Pointer, dataDev unsafe.Pointer) {
}

func wlOnDataDeviceSelection(data unsafe.Pointer, dataDev unsafe.Pointer, id unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	defer s.flushOffers()
	s.clipboard = nil
loop:
	for _, want := range clipboardMimeTypes {
		for _, got := range s.offers[id] {
			if want != got {
				continue
			}
			s.clipboard = id
			s.mimeType = got
			break loop
		}
	}
}

func wlOnRegistryGlobalRemove(data unsafe.Pointer, reg unsafe.Pointer, name uint32) {
	d := callbackLoad(data).(*wlDisplay)
	if s := d.seat; s != nil && name == s.name {
		s.destroy()
		d.seat = nil
	}
	if output, exists := d.outputMap[name]; exists {
		wlOutputDestroy(output)
		delete(d.outputMap, name)
		delete(d.outputConfig, output)
	}
}

func wlOnTouchDown(data unsafe.Pointer, touch unsafe.Pointer, serial, t uint32, surf unsafe.Pointer, id int32, x, y int32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.touchFoci[id] = w
	w.lastTouch = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.ProcessEvent(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Touch,
		Position:  w.lastTouch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

func wlOnTouchUp(data unsafe.Pointer, touch unsafe.Pointer, serial, t uint32, id int32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.touchFoci[id]
	delete(s.touchFoci, id)
	w.ProcessEvent(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Touch,
		Position:  w.lastTouch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

func wlOnTouchMotion(data unsafe.Pointer, touch unsafe.Pointer, t uint32, id int32, x, y int32) {
	s := callbackLoad(data).(*wlSeat)
	w := s.touchFoci[id]
	w.lastTouch = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.ProcessEvent(pointer.Event{
		Kind:      pointer.Move,
		Position:  w.lastTouch,
		Source:    pointer.Touch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

func wlOnTouchFrame(data unsafe.Pointer, touch unsafe.Pointer) {
}

func wlOnTouchCancel(data unsafe.Pointer, touch unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	for id, w := range s.touchFoci {
		delete(s.touchFoci, id)
		w.ProcessEvent(pointer.Event{
			Kind:   pointer.Cancel,
			Source: pointer.Touch,
		})
	}
}

func wlOnPointerEnter(data unsafe.Pointer, pointer unsafe.Pointer, serial uint32, surf unsafe.Pointer, x, y int32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.pointerSerial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.pointerFocus = w
	w.setCursor(pointer, serial)
	w.lastPos = f32.Point{X: fromFixed(x), Y: fromFixed(y)}
}

func wlOnPointerLeave(data unsafe.Pointer, p unsafe.Pointer, serial uint32, surf unsafe.Pointer) {
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.pointerFocus = nil
	if w.inCompositor {
		w.inCompositor = false
		w.ProcessEvent(pointer.Event{Kind: pointer.Cancel})
	}
}

func wlOnPointerMotion(data unsafe.Pointer, p unsafe.Pointer, t uint32, x, y int32) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.resetFling()
	w.onPointerMotion(x, y, t)
}

func wlOnPointerButton(data unsafe.Pointer, p unsafe.Pointer, serial, t, wbtn, state uint32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.pointerFocus
	// From Linux: include/uapi/linux/input-event-codes.h
	const (
		BTN_LEFT   = 0x110
		BTN_RIGHT  = 0x111
		BTN_MIDDLE = 0x112
		BTN_SIDE   = 0x113
		BTN_EXTRA  = 0x114
	)
	var btn pointer.Buttons
	switch wbtn {
	case BTN_LEFT:
		btn = pointer.ButtonPrimary
	case BTN_RIGHT:
		btn = pointer.ButtonSecondary
	case BTN_MIDDLE:
		btn = pointer.ButtonTertiary
	case BTN_SIDE:
		btn = pointer.ButtonQuaternary
	case BTN_EXTRA:
		btn = pointer.ButtonQuinary
	default:
		return
	}
	if state == 1 && btn == pointer.ButtonPrimary {
		if _, edge := w.systemGesture(); edge != 0 {
			w.resize(serial, edge)
			return
		}
		act, ok := w.w.ActionAt(w.lastPos)
		if ok && w.config.Mode == Windowed {
			switch act {
			case system.ActionMove:
				w.move(serial)
				return
			}
		}
	}
	var kind pointer.Kind
	switch state {
	case 0:
		w.pointerBtns &^= btn
		kind = pointer.Release
		// Move or resize gestures no longer applies.
		w.inCompositor = false
	case 1:
		w.pointerBtns |= btn
		kind = pointer.Press
	}
	w.flushScroll()
	w.resetFling()
	w.ProcessEvent(pointer.Event{
		Kind:      kind,
		Source:    pointer.Mouse,
		Buttons:   w.pointerBtns,
		Position:  w.lastPos,
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

func wlOnPointerAxis(data unsafe.Pointer, p unsafe.Pointer, t, axis uint32, value int32) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	v := fromFixed(value)
	w.resetFling()
	if w.scroll.dist == (f32.Point{}) {
		w.scroll.time = time.Duration(t) * time.Millisecond
	}
	switch axis {
	case _WL_POINTER_AXIS_HORIZONTAL_SCROLL:
		w.scroll.dist.X += v
	case _WL_POINTER_AXIS_VERTICAL_SCROLL:
		// horizontal scroll if shift + mousewheel(up/down) pressed.
		if w.disp.xkb.Modifiers() == key.ModShift {
			w.scroll.dist.X += v
		} else {
			w.scroll.dist.Y += v
		}
	}
}

func wlOnPointerFrame(data unsafe.Pointer, p unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	if w == nil {
		return
	}
	w.flushScroll()
	w.flushFling()
}

func (w *window) flushFling() {
	if !w.fling.start {
		return
	}
	w.fling.start = false
	estx, esty := w.fling.xExtrapolation.Estimate(), w.fling.yExtrapolation.Estimate()
	w.fling.xExtrapolation = fling.Extrapolation{}
	w.fling.yExtrapolation = fling.Extrapolation{}
	vel := float32(math.Sqrt(float64(estx.Velocity*estx.Velocity + esty.Velocity*esty.Velocity)))
	_, c := w.getConfig()
	if !w.fling.anim.Start(c, time.Now(), vel) {
		return
	}
	invDist := 1 / vel
	w.fling.dir.X = estx.Velocity * invDist
	w.fling.dir.Y = esty.Velocity * invDist
}

func wlOnPointerAxisSource(data unsafe.Pointer, pointer unsafe.Pointer, source uint32) {
}

func wlOnPointerAxisStop(data unsafe.Pointer, p unsafe.Pointer, t, axis uint32) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.fling.start = true
}

func wlOnPointerAxisDiscrete(data unsafe.Pointer, p unsafe.Pointer, axis uint32, discrete int32) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.resetFling()
	switch axis {
	case _WL_POINTER_AXIS_HORIZONTAL_SCROLL:
		w.scroll.steps.X += int(discrete)
	case _WL_POINTER_AXIS_VERTICAL_SCROLL:
		// horizontal scroll if shift + mousewheel(up/down) pressed.
		if w.disp.xkb.Modifiers() == key.ModShift {
			w.scroll.steps.X += int(discrete)
		} else {
			w.scroll.steps.Y += int(discrete)
		}
	}
}

func (w *window) ReadClipboard() {
	if w.disp.readClipClose != nil {
		return
	}
	w.disp.readClipClose = make(chan struct{})
	r, err := w.disp.readClipboard()
	if r == nil || err != nil {
		return
	}
	// Don't let slow clipboard transfers block event loop.
	go func() {
		defer r.Close()
		data, _ := io.ReadAll(r)
		e := transfer.DataEvent{
			Type: "application/text",
			Open: func() io.ReadCloser {
				return io.NopCloser(bytes.NewReader(data))
			},
		}
		select {
		case w.clipReads <- e:
			w.disp.wakeup()
		case <-w.disp.readClipClose:
		}
	}()
}

func (w *window) WriteClipboard(mime string, s []byte) {
	w.disp.writeClipboard(s)
}

func (w *window) Configure(options []Option) {
	_, cfg := w.getConfig()
	prev := w.config
	cnf := w.config
	cnf.apply(cfg, options)
	w.config.decoHeight = cnf.decoHeight

	switch cnf.Mode {
	case Fullscreen:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Fullscreen
			w.wsize = w.config.Size
			xdgToplevelSetFullscreen(w.topLvl, nil)
		}
	case Minimized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Minimized
			xdgToplevelSetMinimized(w.topLvl)
		}
	case Maximized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Maximized
			w.wsize = w.config.Size
			xdgToplevelSetMaximized(w.topLvl)
			w.setTitle(prev, cnf)
		}
	case Windowed:
		switch prev.Mode {
		case Fullscreen:
			w.config.Mode = Windowed
			w.size = w.wsize.Div(w.scale)
			xdgToplevelUnsetFullscreen(w.topLvl)
		case Minimized:
			w.config.Mode = Windowed
		case Maximized:
			w.config.Mode = Windowed
			w.size = w.wsize.Div(w.scale)
			xdgToplevelUnsetMaximized(w.topLvl)
		}
		w.setTitle(prev, cnf)
		if prev.Size != cnf.Size {
			w.config.Size = cnf.Size
			w.config.Size.Y += int(w.decoHeight()) * w.scale
			w.size = w.config.Size.Div(w.scale)
		}
		w.config.MinSize = cnf.MinSize
		w.config.MaxSize = cnf.MaxSize
		w.setWindowConstraints()
	}
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func (w *window) setWindowConstraints() {
	decoHeight := w.decoHeight()
	if scaled := w.config.MinSize.Div(w.scale); scaled != (image.Point{}) {
		xdgToplevelSetMinSize(w.topLvl, int32(scaled.X), int32(scaled.Y+decoHeight))
	}
	if scaled := w.config.MaxSize.Div(w.scale); scaled != (image.Point{}) {
		xdgToplevelSetMaxSize(w.topLvl, int32(scaled.X), int32(scaled.Y+decoHeight))
	}
}

// decoHeight returns the adjustment for client-side decorations, if applicable.
// The unit is in surface-local coordinates.
func (w *window) decoHeight() int {
	if !w.config.Decorated {
		return int(w.config.decoHeight)
	}
	return 0
}

func (w *window) setTitle(prev, cnf Config) {
	if prev.Title != cnf.Title {
		w.config.Title = cnf.Title
		xdgToplevelSetTitle(w.topLvl, cnf.Title)
	}
}

func (w *window) Perform(actions system.Action) {
	// NB. there is no way for a minimized window to be unminimized.
	// https://wayland.app/protocols/xdg-shell#xdg_toplevel:request:set_minimized
	walkActions(actions, func(action system.Action) {
		switch action {
		case system.ActionClose:
			w.requestClose()
		}
	})
}

func (w *window) move(serial uint32) {
	s := w.disp.seat
	if w.inCompositor || s.pointerFocus != w {
		return
	}
	w.inCompositor = true
	xdgToplevelMove(w.topLvl, s.seat, serial)
}

func (w *window) resize(serial, edge uint32) {
	s := w.disp.seat
	if w.inCompositor || s.pointerFocus != w {
		return
	}
	w.inCompositor = true
	xdgToplevelResize(w.topLvl, s.seat, serial, edge)
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor.cursor = w.loadCursor(cursor)
	w.updateCursor()
}

func (w *window) updateCursor() {
	s := w.disp.seat
	ptr := s.pointer
	if ptr == nil || s.pointerFocus != w {
		return
	}
	w.setCursor(ptr, s.pointerSerial)
}

func (w *window) setCursor(pointer unsafe.Pointer, serial uint32) {
	c := w.cursor.system
	if c == nil {
		c = w.cursor.cursor
	}
	if c == nil {
		wlPointerSetCursor(pointer, serial, nil, 0, 0)
		return
	}
	// Get images[0].
	img := *c.Images
	buf := wlCursorImageGetBuffer(img)
	if buf == nil {
		return
	}
	wlPointerSetCursor(pointer, serial, w.cursor.surf, int32(img.HotspotX/uint32(w.scale)), int32(img.HotspotY/uint32(w.scale)))
	wlSurfaceAttach(w.cursor.surf, buf, 0, 0)
	wlSurfaceDamage(w.cursor.surf, 0, 0, int32(img.Width), int32(img.Height))
	wlSurfaceCommit(w.cursor.surf)
}

func (w *window) resetFling() {
	w.fling.start = false
	w.fling.anim = fling.Animation{}
}

func wlOnKeyboardKeymap(data unsafe.Pointer, keyboard unsafe.Pointer, format uint32, fd int32, size uint32) {
	defer syscall.Close(int(fd))
	s := callbackLoad(data).(*wlSeat)
	s.disp.repeat.Stop(0)
	s.disp.xkb.DestroyKeymapState()
	if format != _WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1 {
		return
	}
	if err := s.disp.xkb.LoadKeymap(int(format), int(fd), int(size)); err != nil {
		// TODO: Do better.
		panic(err)
	}
}

func wlOnKeyboardEnter(data unsafe.Pointer, keyboard unsafe.Pointer, serial uint32, surf unsafe.Pointer, keys unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.keyboardFocus = w
	s.disp.repeat.Stop(0)
	w.config.Focused = true
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func wlOnKeyboardLeave(data unsafe.Pointer, keyboard unsafe.Pointer, serial uint32, surf unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.disp.repeat.Stop(0)
	w := s.keyboardFocus
	w.config.Focused = false
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func wlOnKeyboardKey(data unsafe.Pointer, keyboard unsafe.Pointer, serial, timestamp, keyCode, state uint32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.keyboardFocus
	t := time.Duration(timestamp) * time.Millisecond
	s.disp.repeat.Stop(t)
	w.resetFling()
	kc := mapXKBKeycode(uint32(keyCode))
	ks := mapXKBKeyState(uint32(state))
	_, events := w.disp.xkb.DispatchKey(kc, ks)
	for _, e := range events {
		if ee, ok := e.(key.EditEvent); ok {
			w.w.EditorInsert(ee.Text)
		} else {
			w.ProcessEvent(e)
		}
	}
	if state != _WL_KEYBOARD_KEY_STATE_PRESSED {
		return
	}
	if w.disp.xkb.IsRepeatKey(kc) {
		w.disp.repeat.Start(w, kc, t)
	}
}

func mapXKBKeycode(keyCode uint32) uint32 {
	// According to the xkb_v1 spec: "to determine the xkb keycode, clients must add 8 to the key event keycode."
	return keyCode + 8
}

func mapXKBKeyState(state uint32) key.State {
	switch state {
	case _WL_KEYBOARD_KEY_STATE_RELEASED:
		return key.Release
	default:
		return key.Press
	}
}

func (r *repeatState) Start(w *window, keyCode uint32, t time.Duration) {
	if r.rate <= 0 {
		return
	}
	stopC := make(chan struct{})
	r.start = t
	r.last = 0
	r.now = 0
	r.stopC = stopC
	r.key = keyCode
	r.win = w
	rate, delay := r.rate, r.delay
	go func() {
		timer := time.NewTimer(delay)
		for {
			select {
			case <-timer.C:
			case <-stopC:
				close(stopC)
				return
			}
			r.Advance(delay)
			w.disp.wakeup()
			delay = time.Second / time.Duration(rate)
			timer.Reset(delay)
		}
	}()
}

func (r *repeatState) Stop(t time.Duration) {
	if r.stopC == nil {
		return
	}
	r.stopC <- struct{}{}
	<-r.stopC
	r.stopC = nil
	t -= r.start
	if r.now > t {
		r.now = t
	}
}

func (r *repeatState) Advance(dt time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.now += dt
}

func (r *repeatState) Repeat(d *wlDisplay) {
	if r.rate <= 0 {
		return
	}
	r.mu.Lock()
	now := r.now
	r.mu.Unlock()
	for {
		var delay time.Duration
		if r.last < r.delay {
			delay = r.delay
		} else {
			delay = time.Second / time.Duration(r.rate)
		}
		if r.last+delay > now {
			break
		}
		_, events := d.xkb.DispatchKey(r.key, key.Press)
		for _, e := range events {
			if ee, ok := e.(key.EditEvent); ok {
				r.win.w.EditorInsert(ee.Text)
			} else {
				r.win.ProcessEvent(e)
			}
		}
		r.last += delay
	}
}

func wlOnFrameDone(data unsafe.Pointer, callback unsafe.Pointer, t uint32) {
	wlCallbackDestroy(callback)
	w := callbackLoad(data).(*window)
	if w.lastFrameCallback == callback {
		w.lastFrameCallback = nil
		w.draw(false)
	}
}

func (w *window) close(err error) {
	// View events are queued. Keep the connection alive until the caller has
	// handled the invalid view and calls Event again to acknowledge it.
	w.teardown.begin(func() {
		w.ProcessEvent(WaylandViewEvent{})
	}, func() {
		// Destroy the GPU before the native surface, but deliver DestroyEvent
		// only after all native resources have been released.
		w.runFuncs.close()
		close(w.done)
		w.ProcessEvent(DestroyEvent{Err: err})
		w.destroy()
		w.disp.destroy()
		w.disp = nil
	})
}

func (w *window) dispatch() {
	if w.disp == nil {
		<-w.wakeups
		w.w.Invalidate()
		return
	}
	if err := w.disp.dispatch(); err != nil || w.closing {
		w.close(err)
		return
	}
	select {
	case e := <-w.clipReads:
		w.disp.readClipClose = nil
		w.ProcessEvent(e)
	case <-w.wakeups:
		w.w.Invalidate()
	default:
	}
}

func (w *window) ProcessEvent(e event.Event) {
	w.w.ProcessEvent(e)
}

func (w *window) Event() event.Event {
	w.teardown.complete()
	for {
		w.runFuncs.drain()
		if w.delivered != nil {
			e := w.delivered
			w.delivered = nil
			if !e.abort {
				w.closing = true
				w.disp.wakeup()
			}
		}
		evt, ok := w.w.nextEvent()
		if !ok {
			w.dispatch()
			continue
		}
		if e, ok := evt.(*ClosingEvent); ok && e == w.closeEvent {
			w.delivered = w.closeEvent
			w.closeEvent = nil
		}
		if view, ok := evt.(WaylandViewEvent); ok && !view.Valid() {
			w.teardown.viewReturned()
		}
		return evt
	}
}

func (w *window) requestClose() {
	if w.closeEvent != nil {
		return
	}
	e := &ClosingEvent{}
	w.closeEvent = e
	w.ProcessEvent(e)
}

func (w *window) Invalidate() {
	select {
	case w.wakeups <- struct{}{}:
	default:
		return
	}
	w.disp.wakeup()
}

func (w *window) Run(f func()) {
	w.runFuncs.add(f, w.runWakeup)
}

func (w *window) Done() <-chan struct{} { return w.done }

func (w *window) Frame(frame *op.Ops) {
	w.w.ProcessFrame(frame, nil)
}

// bindDataDevice initializes the dataDev field if and only if both
// the seat and dataDeviceManager fields are initialized.
func (d *wlDisplay) bindDataDevice() {
	if d.seat != nil && d.dataDeviceManager != nil {
		d.seat.dataDev = wlDataDeviceManagerGetDataDevice(d.dataDeviceManager, d.seat.seat)
		if d.seat.dataDev == nil {
			return
		}
		callbackStore(unsafe.Pointer(d.seat.dataDev), d.seat)
		wlAddListener(d.seat.dataDev, lk_data_device, unsafe.Pointer(d.seat.dataDev))
	}
}

func (d *wlDisplay) dispatch() error {
	// wl_display_prepare_read records the current thread for
	// use in wl_display_read_events or wl_display_cancel_events.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dispfd := wlDisplayGetFd(d.disp)
	// Poll for events and notifications.
	pollfds := append(d.poller.pollfds[:0],
		syscall.PollFd{Fd: int32(dispfd), Events: syscall.POLLIN | syscall.POLLERR},
		syscall.PollFd{Fd: int32(d.notify.read), Events: syscall.POLLIN | syscall.POLLERR},
	)
	for wlDisplayPrepareRead(d.disp) != 0 {
		wlDisplayDispatchPending(d.disp)
	}
	dispFd := &pollfds[0]
	if ret, err := wlDisplayFlush(d.disp); ret < 0 {
		if err != syscall.EAGAIN {
			return fmt.Errorf("wayland: wl_display_flush failed: %v", err)
		}
		// EAGAIN means the output buffer was full. Poll for
		// POLLOUT to know when we can write again.
		dispFd.Events |= syscall.POLLOUT
	}
	if _, err := syscall.Poll(pollfds, -1); err != nil && err != syscall.EINTR {
		wlDisplayCancelRead(d.disp)
		return fmt.Errorf("wayland: poll failed: %v", err)
	}
	if dispFd.Revents&(syscall.POLLERR|syscall.POLLHUP) != 0 {
		wlDisplayCancelRead(d.disp)
		return errors.New("wayland: display file descriptor gone")
	}
	// Handle events.
	if dispFd.Revents&syscall.POLLIN != 0 {
		if ret, err := wlDisplayReadEvents(d.disp); ret < 0 {
			return fmt.Errorf("wayland: wl_display_read_events failed: %v", err)
		}
		wlDisplayDispatchPending(d.disp)
	} else {
		wlDisplayCancelRead(d.disp)
	}
	// Clear notifications.
	for {
		_, err := syscall.Read(d.notify.read, d.poller.buf[:])
		if err == syscall.EAGAIN {
			break
		}
		if err != nil {
			return fmt.Errorf("wayland: read from notify pipe failed: %v", err)
		}
	}
	d.repeat.Repeat(d)
	return nil
}

func (w *window) SetAnimating(anim bool) {
	w.animating = anim
	if anim {
		w.draw(false)
	}
}

// Wakeup wakes up the event loop through the notification pipe.
func (d *wlDisplay) wakeup() {
	oneByte := make([]byte, 1)
	if _, err := syscall.Write(d.notify.write, oneByte); err != nil && err != syscall.EAGAIN {
		panic(fmt.Errorf("failed to write to pipe: %v", err))
	}
}

func (w *window) destroy() {
	if s := w.disp.seat; s != nil && s.imFocus == w {
		s.imState.clearPreedit(w.w)
		s.imFocus = nil
		s.resetTextInputState()
	}
	if w.lastFrameCallback != nil {
		wlCallbackDestroy(w.lastFrameCallback)
		w.lastFrameCallback = nil
	}
	if w.cursor.surf != nil {
		wlSurfaceDestroy(w.cursor.surf)
	}
	if w.cursor.theme != nil {
		wlCursorThemeDestroy(w.cursor.theme)
	}
	if w.topLvl != nil {
		xdgToplevelDestroy(w.topLvl)
	}
	if w.surf != nil {
		wlSurfaceDestroy(w.surf)
	}
	if w.wmSurf != nil {
		xdgSurfaceDestroy(w.wmSurf)
	}
	if w.decor != nil {
		zxdgToplevelDecorationV1Destroy(w.decor)
	}
	callbackDelete(unsafe.Pointer(w.surf))
}

func wlOnKeyboardModifiers(data unsafe.Pointer, keyboard unsafe.Pointer, serial, depressed, latched, locked, group uint32) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	d := s.disp
	d.repeat.Stop(0)
	if d.xkb == nil {
		return
	}
	d.xkb.UpdateMask(uint32(depressed), uint32(latched), uint32(locked), uint32(group), uint32(group), uint32(group))
}

func wlOnKeyboardRepeatInfo(data unsafe.Pointer, keyboard unsafe.Pointer, rate, delay int32) {
	s := callbackLoad(data).(*wlSeat)
	d := s.disp
	d.repeat.Stop(0)
	d.repeat.rate = int(rate)
	d.repeat.delay = time.Duration(delay) * time.Millisecond
}

func (s *wlSeat) sendSurroundingText(snapshot textInputSnapshot) {
	zwpTextInputV3SetSurroundingText(s.im, snapshot.text, int32(snapshot.cursor), int32(snapshot.anchor))
	s.imState.sent = snapshot
	s.imState.sentValid = true
}

// refreshTextInput reports the current editor state to the input method.
func (s *wlSeat) refreshTextInput(w *window) {
	snapshot, status := prepareSurroundingText(w.w)
	if status == surroundingAwaitingSnippet {
		return
	}
	if status == surroundingReady {
		if !s.imState.sentValid {
			// The session was enabled without surrounding text, so the
			// compositor may ignore it now. Restart the session instead.
			s.activateTextInput(w)
			return
		}
		s.sendSurroundingText(snapshot)
	}
	cause := uint32(_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_INPUT_METHOD)
	if s.imState.external {
		cause = uint32(_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER)
	}
	zwpTextInputV3SetTextChangeCause(s.im, uint32(cause))
	s.sendCursorRectangle(w)
	s.commitTextInput()
	s.imState.dirty = false
	s.imState.external = false
}

func (s *wlSeat) sendCursorRectangle(w *window) {
	sel := w.w.EditorState().Selection
	top := sel.Transform.Transform(sel.Caret.Pos.Sub(f32.Pt(0, sel.Caret.Ascent)))
	bottom := sel.Transform.Transform(sel.Caret.Pos.Add(f32.Pt(0, sel.Caret.Descent)))
	scale := float32(max(1, w.scale))
	minX, maxX := min(top.X, bottom.X)/scale, max(top.X, bottom.X)/scale
	minY, maxY := min(top.Y, bottom.Y)/scale, max(top.Y, bottom.Y)/scale
	x, y := math.Floor(float64(minX)), math.Floor(float64(minY))
	maxXI, maxYI := math.Ceil(float64(maxX)), math.Ceil(float64(maxY))
	zwpTextInputV3SetCursorRectangle(s.im,
		int32(x), int32(y), int32(maxXI-x), int32(maxYI-y))
}

func textInputContentType(hint key.InputHint) (uint32, uint32) {
	var hints, purpose uint32
	switch hint {
	case key.HintText:
		hints = _ZWP_TEXT_INPUT_V3_CONTENT_HINT_COMPLETION | _ZWP_TEXT_INPUT_V3_CONTENT_HINT_SPELLCHECK
	case key.HintNumeric:
		purpose = _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NUMBER
	case key.HintEmail:
		purpose = _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_EMAIL
	case key.HintURL:
		purpose = _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_URL
	case key.HintTelephone:
		purpose = _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PHONE
	case key.HintPassword:
		hints = _ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT | _ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA
		purpose = _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD
	}
	return hints, purpose
}

func (s *wlSeat) activateTextInput(w *window) {
	snapshot, status := prepareSurroundingText(w.w)
	if status == surroundingAwaitingSnippet {
		return
	}
	if s.imActive {
		s.imState.clearPreedit(w.w)
	}
	s.resetTextInputState()
	zwpTextInputV3Disable(s.im)
	zwpTextInputV3Enable(s.im)
	hints, purpose := textInputContentType(w.inputHint)
	zwpTextInputV3SetContentType(s.im, uint32(hints), uint32(purpose))
	// Enabling without surrounding text tells the compositor we can't
	// provide it. sentValid records that.
	if status == surroundingReady {
		s.sendSurroundingText(snapshot)
	}
	s.sendCursorRectangle(w)
	s.commitTextInput()
	s.imActive = true
	s.imHint = w.inputHint
}

func (s *wlSeat) deactivateTextInput(w *window) {
	if !s.imActive {
		return
	}
	s.imState.clearPreedit(w.w)
	zwpTextInputV3Disable(s.im)
	s.commitTextInput()
	s.resetTextInputState()
}

func wlOnTextInputEnter(data unsafe.Pointer, im unsafe.Pointer, surf unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.imFocus = w
	s.resetTextInputState()
	if w.textInput {
		s.activateTextInput(w)
	} else {
		zwpTextInputV3Disable(s.im)
		s.commitTextInput()
	}
}

func wlOnTextInputLeave(data unsafe.Pointer, im unsafe.Pointer, surf unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	if s.imFocus == nil || s.imFocus.surf != surf {
		return
	}
	s.imState.clearPreedit(s.imFocus.w)
	s.imFocus = nil
	s.resetTextInputState()
}

func wlOnTextInputPreeditString(data unsafe.Pointer, im unsafe.Pointer, text string, set bool, begin, end int32) {
	// The protocol allows a null string; set reports a non-null one.
	s := callbackLoad(data).(*wlSeat)
	s.imState.pending.preedit = text
	s.imState.pending.preeditSet = set
	s.imState.pending.preeditSelection = key.Range{Start: int(begin), End: int(end)}
}

func wlOnTextInputCommitString(data unsafe.Pointer, im unsafe.Pointer, text string, set bool) {
	// The protocol allows a null string; set reports a non-null one.
	s := callbackLoad(data).(*wlSeat)
	s.imState.pending.commit = text
	s.imState.pending.commitSet = set
}

func wlOnTextInputDeleteSurroundingText(data unsafe.Pointer, im unsafe.Pointer, before, after uint32) {
	s := callbackLoad(data).(*wlSeat)
	s.imState.pending.deleteBefore = uint32(before)
	s.imState.pending.deleteAfter = uint32(after)
}

func wlOnTextInputDone(data unsafe.Pointer, im unsafe.Pointer, serial uint32) {
	s := callbackLoad(data).(*wlSeat)
	if s.imFocus == nil {
		s.imState.pending = textInputUpdate{}
		return
	}
	if s.imState.applyDone(s.imFocus.w, uint32(serial)) && s.imActive {
		if s.imHint != s.imFocus.inputHint {
			s.activateTextInput(s.imFocus)
			return
		}
		s.refreshTextInput(s.imFocus)
	}
}

func wlOnDataSourceTarget(data unsafe.Pointer, source unsafe.Pointer, mime string) {
}

func wlOnDataSourceSend(data unsafe.Pointer, source unsafe.Pointer, mime string, fd int32) {
	s := callbackLoad(data).(*wlSeat)
	content := s.content
	go func() {
		defer syscall.Close(int(fd))
		syscall.Write(int(fd), content)
	}()
}

func wlOnDataSourceCancelled(data unsafe.Pointer, source unsafe.Pointer) {
	s := callbackLoad(data).(*wlSeat)
	if s.source == source {
		s.content = nil
		s.source = nil
	}
	wlDataSourceDestroy(source)
}

func wlOnDataSourceDNDDropPerformed(data unsafe.Pointer, source unsafe.Pointer) {
}

func wlOnDataSourceDNDFinished(data unsafe.Pointer, source unsafe.Pointer) {
}

func wlOnDataSourceAction(data unsafe.Pointer, source unsafe.Pointer, act uint32) {
}

func (w *window) flushScroll() {
	var fling f32.Point
	if w.fling.anim.Active() {
		dist := float32(w.fling.anim.Tick(time.Now()))
		fling = w.fling.dir.Mul(dist)
	}
	// The Wayland reported scroll distance for
	// discrete scroll axes is only 10 pixels, where
	// 100 seems more appropriate.
	const discreteScale = 10
	if w.scroll.steps.X != 0 {
		w.scroll.dist.X *= discreteScale
	}
	if w.scroll.steps.Y != 0 {
		w.scroll.dist.Y *= discreteScale
	}
	total := w.scroll.dist.Add(fling)
	if total == (f32.Point{}) {
		return
	}
	if w.scroll.steps == (image.Point{}) {
		w.fling.xExtrapolation.SampleDelta(w.scroll.time, -w.scroll.dist.X)
		w.fling.yExtrapolation.SampleDelta(w.scroll.time, -w.scroll.dist.Y)
	}
	// Zero scroll distance prior to calling ProcessEvent, otherwise we may recursively
	// re-process the scroll distance.
	w.scroll.dist = f32.Point{}
	w.scroll.steps = image.Point{}
	w.ProcessEvent(pointer.Event{
		Kind:      pointer.Scroll,
		Source:    pointer.Mouse,
		Buttons:   w.pointerBtns,
		Position:  w.lastPos,
		Scroll:    total,
		Time:      w.scroll.time,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

func (w *window) onPointerMotion(x, y int32, t uint32) {
	w.flushScroll()
	w.lastPos = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.ProcessEvent(pointer.Event{
		Kind:      pointer.Move,
		Position:  w.lastPos,
		Buttons:   w.pointerBtns,
		Source:    pointer.Mouse,
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
	c, _ := w.systemGesture()
	if c != w.cursor.system {
		w.cursor.system = c
		w.updateCursor()
	}
}

// updateCursor updates the system gesture cursor according to the pointer
// position.
func (w *window) systemGesture() (*wlCursor, uint32) {
	if w.config.Mode != Windowed || w.config.Decorated {
		return nil, 0
	}
	_, cfg := w.getConfig()
	border := cfg.Dp(3)
	x, y, size := int(w.lastPos.X), int(w.lastPos.Y), w.config.Size
	north := y <= border
	south := y >= size.Y-border
	west := x <= border
	east := x >= size.X-border

	switch {
	default:
		fallthrough
	case !north && !south && !west && !east:
		return nil, 0
	case north && west:
		return w.cursor.cursors.resizeNorthWest, _XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT
	case north && east:
		return w.cursor.cursors.resizeNorthEast, _XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT
	case south && west:
		return w.cursor.cursors.resizeSouthWest, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT
	case south && east:
		return w.cursor.cursors.resizeSouthEast, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT
	case north:
		return w.cursor.cursors.resizeNorth, _XDG_TOPLEVEL_RESIZE_EDGE_TOP
	case south:
		return w.cursor.cursors.resizeSouth, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM
	case west:
		return w.cursor.cursors.resizeWest, _XDG_TOPLEVEL_RESIZE_EDGE_LEFT
	case east:
		return w.cursor.cursors.resizeEast, _XDG_TOPLEVEL_RESIZE_EDGE_RIGHT
	}
}

func (w *window) updateOpaqueRegion() {
	reg := wlCompositorCreateRegion(w.disp.compositor)
	wlRegionAdd(reg, 0, 0, int32(w.size.X), int32(w.size.Y))
	wlSurfaceSetOpaqueRegion(w.surf, reg)
	wlRegionDestroy(reg)
}

func (w *window) updateOutputs() {
	scale := 1
	var found bool
	for _, conf := range w.disp.outputConfig {
		for _, w2 := range conf.windows {
			if w2 == w {
				found = true
				if conf.scale > scale {
					scale = conf.scale
				}
			}
		}
	}
	if found && scale != w.scale {
		w.scale = scale
		wlSurfaceSetBufferScale(w.surf, int32(w.scale))
		w.draw(true)
	}
	if found {
		w.draw(true)
	}
}

func (w *window) getConfig() (image.Point, unit.Metric) {
	size := w.size.Mul(w.scale)
	return size, unit.Metric{
		PxPerDp: w.ppdp * float32(w.scale),
		PxPerSp: w.ppsp * float32(w.scale),
	}
}

func (w *window) draw(sync bool) {
	if !w.configured {
		return
	}
	w.flushScroll()
	size, cfg := w.getConfig()
	if cfg == (unit.Metric{}) {
		return
	}
	if size != w.config.Size {
		w.config.Size = size
		w.ProcessEvent(ConfigEvent{Config: w.config})
	}
	anim := w.animating || w.fling.anim.Active()
	// Draw animation only when not waiting for frame callback.
	redrawAnim := anim && w.lastFrameCallback == nil
	if !redrawAnim && !sync {
		return
	}
	if anim {
		w.lastFrameCallback = wlSurfaceFrame(w.surf)
		// Use the surface as listener data for gio_onFrameDone.
		wlAddListener(w.lastFrameCallback, lk_callback, unsafe.Pointer(w.surf))
	}
	w.ProcessEvent(frameEvent{
		FrameEvent: FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Metric: cfg,
		},
		Sync: sync,
	})
}

func (w *window) display() unsafe.Pointer {
	return w.disp.disp
}

func (w *window) surface() (unsafe.Pointer, int, int) {
	sz, _ := w.getConfig()
	return w.surf, sz.X, sz.Y
}

func (w *window) ShowTextInput(show bool) {
	w.textInput = show
	s := w.disp.seat
	if s == nil || s.im == nil || s.imFocus != w {
		return
	}
	if show {
		if !s.imActive {
			s.activateTextInput(w)
		}
	} else {
		s.deactivateTextInput(w)
	}
}

func (w *window) SetInputHint(hint key.InputHint) {
	if w.inputHint == hint {
		return
	}
	w.inputHint = hint
	s := w.disp.seat
	if s != nil && s.im != nil && s.imFocus == w && s.imActive {
		s.activateTextInput(w)
	}
}

func (w *window) EditorStateChanged(old, new editorState) {
	s := w.disp.seat
	if s == nil || s.im == nil || s.imFocus != w || !w.textInput {
		return
	}
	if !s.imActive {
		s.activateTextInput(w)
		return
	}
	if s.imHint != w.inputHint {
		s.activateTextInput(w)
		return
	}
	selectionChanged := old.Selection.Range != new.Selection.Range
	textChanged := selectionChanged || !areSnippetsConsistent(old.Snippet, new.Snippet)
	if s.imState.stale {
		// The last done serial didn't match a commit; defer protocol
		// updates until a matching done, then report all changes.
		s.imState.dirty = true
		s.imState.external = s.imState.external || textChanged
		return
	}
	if s.imState.dirty {
		s.imState.external = s.imState.external || textChanged
		s.refreshTextInput(w)
		return
	}
	cancelled := shouldCancelComposition(old, new) && new.compose.Start != -1
	if cancelled {
		s.imState.clearPreedit(w.w)
	}
	surroundingChanged := cancelled || selectionChanged || old.Snippet != new.Snippet
	caretChanged := old.Selection.Caret != new.Selection.Caret || old.Selection.Transform != new.Selection.Transform
	needsCommit := false
	if surroundingChanged {
		snapshot, status := prepareSurroundingText(w.w)
		switch status {
		case surroundingReady:
			if !s.imState.sentValid {
				s.activateTextInput(w)
				return
			}
			s.sendSurroundingText(snapshot)
			if textChanged {
				zwpTextInputV3SetTextChangeCause(s.im, _ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER)
			}
			needsCommit = true
		}
	}
	if caretChanged {
		s.sendCursorRectangle(w)
		needsCommit = true
	}
	if needsCommit {
		s.commitTextInput()
		s.imState.dirty = false
		s.imState.external = false
	}
}

func (w *window) NewContext() (context, error) {
	var firstErr error
	if f := newWaylandEGLContext; f != nil {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if f := newWaylandVulkanContext; f != nil {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errors.New("wayland: no available GPU backends")
}

// detectUIScale reports the system UI scale, or 1.0 if it fails.
func detectUIScale() float32 {
	// TODO: What about other window environments?
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "text-scaling-factor").Output()
	if err != nil {
		return 1.0
	}
	scale, err := strconv.ParseFloat(string(bytes.TrimSpace(out)), 32)
	if err != nil {
		return 1.0
	}
	return float32(scale)
}

func newWLDisplay() (*wlDisplay, error) {
	d := &wlDisplay{
		outputMap:    make(map[uint32]unsafe.Pointer),
		outputConfig: make(map[unsafe.Pointer]*wlOutput),
	}
	pipe := make([]int, 2)
	if err := syscall.Pipe2(pipe, syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		return nil, fmt.Errorf("wayland: failed to create pipe: %v", err)
	}
	d.notify.read = pipe[0]
	d.notify.write = pipe[1]
	xkb, err := xkb.New()
	if err != nil {
		d.destroy()
		return nil, fmt.Errorf("wayland: %v", err)
	}
	d.xkb = xkb
	d.disp, err = wlDisplayConnect()
	if d.disp == nil {
		d.destroy()
		return nil, fmt.Errorf("wayland: wl_display_connect failed: %v", err)
	}
	callbackMap.Store(unsafe.Pointer(d.disp), d)
	d.reg = wlDisplayGetRegistry(d.disp)
	if d.reg == nil {
		d.destroy()
		return nil, errors.New("wayland: wl_display_get_registry failed")
	}
	wlAddListener(d.reg, lk_registry, unsafe.Pointer(d.disp))
	// Wait for the server to register all its globals to the
	// registry listener (gio_onRegistryGlobal).
	wlDisplayRoundtrip(d.disp)
	// Configuration listeners are added to outputs by gio_onRegistryGlobal.
	// We need another roundtrip to get the initial output configurations
	// through the gio_onOutput* callbacks.
	wlDisplayRoundtrip(d.disp)
	return d, nil
}

func (d *wlDisplay) destroy() {
	if d.readClipClose != nil {
		close(d.readClipClose)
		d.readClipClose = nil
	}
	if d.notify.write != 0 {
		syscall.Close(d.notify.write)
		d.notify.write = 0
	}
	if d.notify.read != 0 {
		syscall.Close(d.notify.read)
		d.notify.read = 0
	}
	d.repeat.Stop(0)
	if d.xkb != nil {
		d.xkb.Destroy()
		d.xkb = nil
	}
	if d.seat != nil {
		d.seat.destroy()
		d.seat = nil
	}
	if d.imm != nil {
		zwpTextInputManagerV3Destroy(d.imm)
	}
	if d.decor != nil {
		zxdgDecorationManagerV1Destroy(d.decor)
	}
	if d.shm != nil {
		wlShmDestroy(d.shm)
	}
	if d.compositor != nil {
		wlCompositorDestroy(d.compositor)
	}
	if d.wm != nil {
		xdgWmBaseDestroy(d.wm)
	}
	for _, output := range d.outputMap {
		wlOutputDestroy(output)
	}
	if d.reg != nil {
		wlRegistryDestroy(d.reg)
	}
	if d.disp != nil {
		wlDisplayDisconnect(d.disp)
		callbackDelete(unsafe.Pointer(d.disp))
		d.disp = nil
	}
}

// fromFixed converts a Wayland wl_fixed_t 23.8 number to float32.
func fromFixed(v int32) float32 {
	// Convert to float64 to avoid overflow.
	// From wayland-util.h.
	b := ((1023 + 44) << 52) + (1 << 51) + uint64(v)
	f := math.Float64frombits(b) - (3 << 43)
	return float32(f)
}
