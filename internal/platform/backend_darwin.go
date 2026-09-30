//go:build darwin && cgo

package platform

/*
#cgo CFLAGS: -x objective-c -fobjc-arc -mmacosx-version-min=14.0
#cgo LDFLAGS: -framework AppKit -framework CoreGraphics -framework ApplicationServices -framework Carbon -framework ScreenCaptureKit
#include <stdlib.h>
#include "native_darwin.h"
*/
import "C"
import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
)

type darwinBackend struct{}

func New() driver.Backend { return darwinBackend{} }
func nativeError(code C.int) error {
	switch code {
	case 0:
		return nil
	case 1:
		return capability.ErrPermissionDenied
	case 2:
		return capability.ErrUnsupported
	case 3:
		return capability.ErrInvalidArgument
	case 4:
		return capability.ErrNotReady
	case 5:
		return capability.ErrTimeout
	case 6:
		return capability.ErrConflict
	default:
		return fmt.Errorf("%w: status %d", capability.ErrNative, int(code))
	}
}
func (darwinBackend) Check(k driver.Permission, request bool) (driver.Status, error) {
	r := C.int(0)
	if request {
		r = 1
	}
	s := C.keel_permission(C.int(k), r)
	if s < 0 {
		return driver.NotGranted, nativeError(-s)
	}
	if s == 1 {
		return driver.Granted, nil
	}
	return driver.NotGranted, nil
}
func (darwinBackend) Displays() ([]driver.Display, error) {
	var p *C.keel_display
	var n C.uint32_t
	if err := nativeError(C.keel_displays(&p, &n)); err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(p))
	out := make([]driver.Display, int(n))
	for i, d := range unsafe.Slice(p, int(n)) {
		out[i] = driver.Display{ID: uint32(d.id), Bounds: driver.Rect{X: float64(d.x), Y: float64(d.y), Width: float64(d.w), Height: float64(d.h)}, PixelWidth: int(d.pw), PixelHeight: int(d.ph), Primary: d.primary != 0}
	}
	return out, nil
}
func (darwinBackend) Capture(id uint32) ([]byte, error) {
	var p unsafe.Pointer
	var n C.size_t
	if err := nativeError(C.keel_capture(C.uint32_t(id), &p, &n)); err != nil {
		return nil, err
	}
	defer C.free(p)
	if uint64(n) > math.MaxInt32 {
		return nil, fmt.Errorf("%w: screenshot too large", capability.ErrNative)
	}
	return C.GoBytes(p, C.int(n)), nil
}
func (darwinBackend) Position() (driver.Point, error) {
	var x, y C.double
	err := nativeError(C.keel_position(&x, &y))
	return driver.Point{X: float64(x), Y: float64(y)}, err
}
func (darwinBackend) Move(p driver.Point) error {
	return nativeError(C.keel_move(C.double(p.X), C.double(p.Y)))
}
func (darwinBackend) Click(b uint8) error { return nativeError(C.keel_click(C.int(b))) }
func (darwinBackend) Key(key string, down bool) error {
	code, ok := keyCodes[key]
	if !ok {
		return fmt.Errorf("%w: unknown key %q", capability.ErrInvalidArgument, key)
	}
	d := C.int(0)
	if down {
		d = 1
	}
	return nativeError(C.keel_key(C.ushort(code), d))
}
func (darwinBackend) Window(h unsafe.Pointer, op string, v int) error {
	n := map[string]int{"top": 0, "click": 1, "vibrancy": 2}
	o, ok := n[op]
	if !ok {
		return capability.ErrInvalidArgument
	}
	return nativeError(C.keel_window(h, C.int(o), C.int(v)))
}

// Carbon virtual key positions are independent of the active keyboard layout.
var keyCodes = map[string]uint16{
	"a": 0, "s": 1, "d": 2, "f": 3, "h": 4, "g": 5, "z": 6, "x": 7, "c": 8, "v": 9, "b": 11, "q": 12, "w": 13, "e": 14, "r": 15, "y": 16, "t": 17,
	"1": 18, "2": 19, "3": 20, "4": 21, "6": 22, "5": 23, "equal": 24, "9": 25, "7": 26, "minus": 27, "8": 28, "0": 29, "rightbracket": 30, "o": 31, "u": 32, "leftbracket": 33, "i": 34, "p": 35,
	"enter": 36, "l": 37, "j": 38, "quote": 39, "k": 40, "semicolon": 41, "backslash": 42, "comma": 43, "slash": 44, "n": 45, "m": 46, "period": 47, "tab": 48, "space": 49, "backquote": 50, "backspace": 51, "escape": 53,
	"super": 55, "shift": 56, "alt": 58, "control": 59, "f1": 122, "f2": 120, "f3": 99, "f4": 118, "f5": 96, "f6": 97, "f7": 98, "f8": 100, "f9": 101, "f10": 109, "f11": 103, "f12": 111, "home": 115, "pageup": 116, "delete": 117, "end": 119, "pagedown": 121, "left": 123, "right": 124, "down": 125, "up": 126,
}

type hotkeyCallback struct {
	events  chan struct{}
	done    chan struct{}
	handler func()
}

var hotkeys = struct {
	sync.Mutex
	next      uint32
	callbacks map[uint32]*hotkeyCallback
}{callbacks: make(map[uint32]*hotkeyCallback)}

//export keelHotkeyFired
func keelHotkeyFired(id C.uint32_t) {
	hotkeys.Lock()
	defer hotkeys.Unlock()
	if cb := hotkeys.callbacks[uint32(id)]; cb != nil {
		select {
		case cb.events <- struct{}{}:
		default:
		}
	}
}
func (darwinBackend) Register(c driver.Chord, handler func()) (func() error, error) {
	code, ok := keyCodes[c.Key]
	if !ok {
		return nil, fmt.Errorf("%w: unknown key %q", capability.ErrInvalidArgument, c.Key)
	}
	cb := &hotkeyCallback{make(chan struct{}, 1), make(chan struct{}), handler}
	hotkeys.Lock()
	if hotkeys.next == math.MaxUint32 {
		hotkeys.Unlock()
		return nil, capability.ErrNative
	}
	hotkeys.next++
	id := hotkeys.next
	hotkeys.callbacks[id] = cb
	hotkeys.Unlock()
	var ref unsafe.Pointer
	if err := nativeError(C.keel_hotkey_register(C.uint32_t(id), C.ushort(code), C.uint(c.Modifiers), &ref)); err != nil {
		hotkeys.Lock()
		delete(hotkeys.callbacks, id)
		hotkeys.Unlock()
		return nil, err
	}
	go func() {
		for {
			select {
			case <-cb.done:
				return
			case <-cb.events:
				select {
				case <-cb.done:
					return
				default:
				}
				cb.handler()
			}
		}
	}()
	var mu sync.Mutex
	closed := false
	return func() error {
		mu.Lock()
		defer mu.Unlock()
		if closed {
			return nil
		}
		if err := nativeError(C.keel_hotkey_unregister(ref)); err != nil {
			return err
		}
		hotkeys.Lock()
		delete(hotkeys.callbacks, id)
		close(cb.done)
		hotkeys.Unlock()
		closed = true
		return nil
	}, nil
}
