//go:build darwin && cgo

package sys

/*
#cgo CFLAGS: -x objective-c -fobjc-arc -mmacosx-version-min=14.0
#cgo LDFLAGS: -framework AppKit -framework CoreGraphics -framework ApplicationServices -framework Carbon -framework ScreenCaptureKit
#include <stdlib.h>
#include "sys_darwin.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/dyike/keel/native"
)

func status(code C.int) error {
	switch code {
	case 0:
		return nil
	case 1:
		return native.ErrPermissionDenied
	case 2:
		return native.ErrUnsupported
	case 3:
		return native.ErrInvalidArgument
	case 4:
		return fmt.Errorf("%w: must not run on the main thread", native.ErrFailed)
	case 5:
		return native.ErrTimeout
	case 6:
		return native.ErrConflict
	default:
		return fmt.Errorf("%w: status %d", native.ErrFailed, int(code))
	}
}

func Permission(kind int, request bool) (bool, error) {
	r := C.int(0)
	if request {
		r = 1
	}
	s := C.keel_permission(C.int(kind), r)
	if s < 0 {
		return false, status(-s)
	}
	return s == 1, nil
}

func Displays() ([]Display, error) {
	var p *C.keel_display
	var n C.uint32_t
	if err := status(C.keel_displays(&p, &n)); err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(p))
	out := make([]Display, int(n))
	for i, d := range unsafe.Slice(p, int(n)) {
		out[i] = Display{ID: uint32(d.id), X: float64(d.x), Y: float64(d.y), Width: float64(d.w), Height: float64(d.h),
			PixelWidth: int(d.pw), PixelHeight: int(d.ph), Primary: d.primary != 0}
	}
	return out, nil
}

func Capture(id uint32) ([]byte, error) {
	var p unsafe.Pointer
	var n C.size_t
	if err := status(C.keel_capture(C.uint32_t(id), &p, &n)); err != nil {
		return nil, err
	}
	defer C.free(p)
	return C.GoBytes(p, C.int(n)), nil
}

func MousePosition() (float64, float64, error) {
	var x, y C.double
	err := status(C.keel_position(&x, &y))
	return float64(x), float64(y), err
}
func MouseMove(x, y float64) error { return status(C.keel_move(C.double(x), C.double(y))) }
func Click(button int) error       { return status(C.keel_click(C.int(button))) }

func HasKey(key string) bool { _, ok := keyCodes[key]; return ok }

func Key(key string, down bool) error {
	code, ok := keyCodes[key]
	if !ok {
		return fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	d := C.int(0)
	if down {
		d = 1
	}
	return status(C.keel_key(C.ushort(code), d))
}

// Carbon virtual key codes follow physical positions, independent of keyboard layout.
var keyCodes = map[string]uint16{
	"a": 0, "s": 1, "d": 2, "f": 3, "h": 4, "g": 5, "z": 6, "x": 7, "c": 8, "v": 9, "b": 11, "q": 12, "w": 13, "e": 14, "r": 15, "y": 16, "t": 17,
	"1": 18, "2": 19, "3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "9": 25, "7": 26, "-": 27, "8": 28, "0": 29, "]": 30, "o": 31, "u": 32, "[": 33, "i": 34, "p": 35,
	"enter": 36, "l": 37, "j": 38, "'": 39, "k": 40, ";": 41, "\\": 42, ",": 43, "/": 44, "n": 45, "m": 46, ".": 47, "tab": 48, "space": 49, "`": 50, "backspace": 51, "escape": 53,
	"cmd": 55, "shift": 56, "alt": 58, "ctrl": 59, "f1": 122, "f2": 120, "f3": 99, "f4": 118, "f5": 96, "f6": 97, "f7": 98, "f8": 100, "f9": 101, "f10": 109, "f11": 103, "f12": 111,
	"home": 115, "pageup": 116, "delete": 117, "end": 119, "pagedown": 121, "left": 123, "right": 124, "down": 125, "up": 126,
}

var hotkeys struct {
	sync.Mutex
	next  uint32
	fired map[uint32]chan struct{}
}

//export keelHotkeyFired
func keelHotkeyFired(id C.uint32_t) {
	hotkeys.Lock()
	defer hotkeys.Unlock()
	if ch := hotkeys.fired[uint32(id)]; ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Hotkey registers key+mods; fn runs on its own goroutine and presses that
// arrive while it runs are merged.
func Hotkey(key string, mods uint, fn func()) (func() error, error) {
	code, ok := keyCodes[key]
	if !ok {
		return nil, fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	fired := make(chan struct{}, 1)
	hotkeys.Lock()
	if hotkeys.fired == nil {
		hotkeys.fired = map[uint32]chan struct{}{}
	}
	hotkeys.next++
	id := hotkeys.next
	hotkeys.fired[id] = fired
	hotkeys.Unlock()
	var ref unsafe.Pointer
	if err := status(C.keel_hotkey_register(C.uint32_t(id), C.ushort(code), C.uint(mods), &ref)); err != nil {
		hotkeys.Lock()
		delete(hotkeys.fired, id)
		hotkeys.Unlock()
		return nil, err
	}
	go func() {
		for range fired {
			fn()
		}
	}()
	var once sync.Once
	var err error
	return func() error {
		once.Do(func() {
			err = status(C.keel_hotkey_unregister(ref))
			hotkeys.Lock()
			delete(hotkeys.fired, id)
			close(fired)
			hotkeys.Unlock()
		})
		return err
	}, nil
}
