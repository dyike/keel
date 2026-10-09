//go:build darwin && !ios

package sys

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func status(code int) error {
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
		return fmt.Errorf("%w: status %d", native.ErrFailed, code)
	}
}

// Permission kinds: accessibility, screen recording, input monitoring.
func Permission(kind int, request bool) (bool, error) {
	load()
	var granted uintptr
	switch kind {
	case 0:
		if !request {
			granted = call("AXIsProcessTrusted")
			break
		}
		withPool(func() {
			yes := send(class("NSNumber"), "numberWithBool:", 1)
			options := send(class("NSDictionary"), "dictionaryWithObject:forKey:", uintptr(yes), uintptr(constant("kAXTrustedCheckOptionPrompt")))
			granted = call("AXIsProcessTrustedWithOptions", uintptr(options))
		})
	case 1:
		granted = call(pick(request, "CGRequestScreenCaptureAccess", "CGPreflightScreenCaptureAccess"))
	case 2:
		granted = call(pick(request, "CGRequestListenEventAccess", "CGPreflightListenEventAccess"))
	default:
		return false, status(3)
	}
	return byte(granted) != 0, nil
}

func pick(first bool, a, b string) string {
	if first {
		return a
	}
	return b
}

func Displays() ([]Display, error) {
	load()
	var n uint32
	if err := call("CGGetActiveDisplayList", 0, 0, uintptr(unsafe.Pointer(&n))); int32(err) != 0 {
		return nil, status(100 + int(int32(err)))
	}
	if n == 0 {
		return []Display{}, nil
	}
	ids := make([]uint32, n)
	if err := call("CGGetActiveDisplayList", uintptr(n), uintptr(unsafe.Pointer(&ids[0])), uintptr(unsafe.Pointer(&n))); int32(err) != 0 {
		return nil, status(100 + int(int32(err)))
	}
	out := make([]Display, int(n))
	for i, d := range ids[:n] {
		b := cgDisplayBounds(d)
		out[i] = Display{ID: d, X: b.Origin.X, Y: b.Origin.Y, Width: b.Size.Width, Height: b.Size.Height,
			PixelWidth: int(call("CGDisplayPixelsWide", uintptr(d))), PixelHeight: int(call("CGDisplayPixelsHigh", uintptr(d))),
			Primary: uint32(call("CGDisplayIsMain", uintptr(d))) != 0}
	}
	return out, nil
}

const nsBitmapImageFileTypePNG = 4

// Capture takes a PNG of a display with ScreenCaptureKit (macOS 14+). It
// waits for the system, so it must not run on the main thread.
func Capture(displayID uint32) ([]byte, error) {
	load()
	if isMainThread() {
		return nil, status(4)
	}
	if byte(call("CGPreflightScreenCaptureAccess")) == 0 {
		return nil, status(1)
	}
	if uint32(call("CGDisplayIsActive", uintptr(displayID))) == 0 {
		return nil, status(3)
	}
	if class("SCShareableContent") == 0 || class("SCScreenshotManager") == 0 {
		return nil, status(2)
	}
	type result struct {
		png  []byte
		code int
	}
	// Buffered: the system may answer after the timeout.
	done := make(chan result, 1)
	captured := objc.NewBlock(func(_ objc.Block, image uintptr, err id) {
		r := result{code: 100}
		if image != 0 && err == 0 {
			withPool(func() {
				rep := send(send(class("NSBitmapImageRep"), "alloc"), "initWithCGImage:", image)
				if png := send(rep, "representationUsingType:properties:", nsBitmapImageFileTypePNG, uintptr(send(class("NSDictionary"), "dictionary"))); png != 0 {
					r = result{png: goBytes(png)}
				}
				release(rep)
			})
		}
		done <- r
	})
	content := objc.NewBlock(func(_ objc.Block, content, err id) {
		if content == 0 || err != 0 {
			done <- result{code: 100}
			return
		}
		withPool(func() {
			var display id
			displays := send(content, "displays")
			for i := uintptr(0); i < uintptr(send(displays, "count")); i++ {
				d := send(displays, "objectAtIndex:", i)
				if uint32(send(d, "displayID")) == displayID {
					display = d
					break
				}
			}
			if display == 0 {
				done <- result{code: 3}
				return
			}
			filter := send(send(class("SCContentFilter"), "alloc"), "initWithDisplay:excludingWindows:", uintptr(display), uintptr(send(class("NSArray"), "array")))
			config := send(class("SCStreamConfiguration"), "new")
			send(config, "setWidth:", call("CGDisplayPixelsWide", uintptr(displayID)))
			send(config, "setHeight:", call("CGDisplayPixelsHigh", uintptr(displayID)))
			send(config, "setShowsCursor:", 0)
			send(class("SCScreenshotManager"), "captureImageWithFilter:configuration:completionHandler:", uintptr(filter), uintptr(config), uintptr(captured))
			release(filter)
			release(config)
		})
	})
	withPool(func() {
		send(class("SCShareableContent"), "getShareableContentExcludingDesktopWindows:onScreenWindowsOnly:completionHandler:", 0, 1, uintptr(content))
	})
	// The system copied the blocks; ours are no longer needed once they ran.
	defer content.Release()
	select {
	case r := <-done:
		// The capture block may still be referenced until the system
		// returns from calling it; release after a grace period.
		go func() { time.Sleep(time.Second); captured.Release() }()
		if r.code != 0 {
			return nil, status(r.code)
		}
		return r.png, nil
	case <-time.After(10 * time.Second):
		return nil, status(5)
	}
}

const (
	cgHIDEventTap       = 0
	cgEventMouseMoved   = 5
	cgMouseClickState   = 1
	carbonEventNotFound = -9874
)

func accessible() bool { load(); return byte(call("AXIsProcessTrusted")) != 0 }

func MousePosition() (float64, float64, error) {
	load()
	e := call("CGEventCreate", 0)
	if e == 0 {
		return 0, 0, status(100)
	}
	p := cgEventGetLocation(e)
	cfRelease(e)
	return p.X, p.Y, nil
}

func MouseMove(x, y float64) error {
	if !accessible() {
		return status(1)
	}
	e := cgEventCreateMouse(0, cgEventMouseMoved, cgPoint{x, y}, 0)
	if e == 0 {
		return status(100)
	}
	call("CGEventPost", cgHIDEventTap, e)
	cfRelease(e)
	return nil
}

func Click(button int) error {
	if !accessible() {
		return status(1)
	}
	if button < 0 || button > 2 {
		return status(3)
	}
	x, y, err := MousePosition()
	if err != nil {
		return err
	}
	down := [...]uint32{1, 3, 25} // left, right, other mouse down
	up := [...]uint32{2, 4, 26}
	d := cgEventCreateMouse(0, down[button], cgPoint{x, y}, uint32(button))
	u := cgEventCreateMouse(0, up[button], cgPoint{x, y}, uint32(button))
	defer cfRelease(d)
	defer cfRelease(u)
	if d == 0 || u == 0 {
		return status(100)
	}
	call("CGEventSetIntegerValueField", d, cgMouseClickState, 1)
	call("CGEventSetIntegerValueField", u, cgMouseClickState, 1)
	call("CGEventPost", cgHIDEventTap, d)
	call("CGEventPost", cgHIDEventTap, u)
	return nil
}

func HasKey(key string) bool { _, ok := keyCodes[key]; return ok }

func Key(key string, down bool) error {
	code, ok := keyCodes[key]
	if !ok {
		return fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	if !accessible() {
		return status(1)
	}
	e := cgEventCreateKeyboard(0, code, down)
	if e == 0 {
		return status(100)
	}
	call("CGEventPost", cgHIDEventTap, e)
	cfRelease(e)
	return nil
}

// Carbon virtual key codes follow physical positions, independent of keyboard layout.
var keyCodes = map[string]uint16{
	"a": 0, "s": 1, "d": 2, "f": 3, "h": 4, "g": 5, "z": 6, "x": 7, "c": 8, "v": 9, "b": 11, "q": 12, "w": 13, "e": 14, "r": 15, "y": 16, "t": 17,
	"1": 18, "2": 19, "3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "9": 25, "7": 26, "-": 27, "8": 28, "0": 29, "]": 30, "o": 31, "u": 32, "[": 33, "i": 34, "p": 35,
	"enter": 36, "l": 37, "j": 38, "'": 39, "k": 40, ";": 41, "\\": 42, ",": 43, "/": 44, "n": 45, "m": 46, ".": 47, "tab": 48, "space": 49, "`": 50, "backspace": 51, "escape": 53,
	"cmd": 55, "shift": 56, "alt": 58, "ctrl": 59, "f1": 122, "f2": 120, "f3": 99, "f4": 118, "f5": 96, "f6": 97, "f7": 98, "f8": 100, "f9": 101, "f10": 109, "f11": 103, "f12": 111,
	"home": 115, "pageup": 116, "delete": 117, "end": 119, "pagedown": 121, "left": 123, "right": 124, "down": 125, "up": 126,
}

// Carbon hotkeys: one event handler for every hotkey, installed while any
// is registered, and one callback for the process.
var hotkeys struct {
	sync.Mutex
	next  uint32
	fired map[uint32]chan struct{}
	// Main thread only.
	handler  uintptr
	count    int
	callback uintptr
}

const (
	hotkeySignature         = 'K'<<24 | 'E'<<16 | 'E'<<8 | 'L'
	kEventClassKeyboard     = 'k'<<24 | 'e'<<16 | 'y'<<8 | 'b'
	kEventHotKeyPressed     = 5
	kEventParamDirectObject = '-'<<24 | '-'<<16 | '-'<<8 | '-'
	typeEventHotKeyID       = 'h'<<24 | 'k'<<16 | 'i'<<8 | 'd'
	eventHotKeyExistsErr    = -9878
)

type eventHotKeyID struct{ Signature, ID uint32 }

func hotkeyEvent(next, event, _ uintptr) uintptr {
	var hk eventHotKeyID
	s := call("GetEventParameter", event, kEventParamDirectObject, typeEventHotKeyID, 0, unsafe.Sizeof(hk), 0, uintptr(unsafe.Pointer(&hk)))
	if int32(s) != 0 || hk.Signature != hotkeySignature {
		notHandled := int32(carbonEventNotFound)
		return uintptr(uint32(notHandled))
	}
	hotkeys.Lock()
	defer hotkeys.Unlock()
	if ch := hotkeys.fired[hk.ID]; ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	return 0
}

var carbonModifiers = [...]uintptr{ModCtrl: 4096, ModAlt: 2048, ModShift: 512, ModCmd: 256}

// Hotkey registers key+mods; fn runs on its own goroutine and presses that
// arrive while it runs are merged.
func Hotkey(key string, mods uint, fn func()) (func() error, error) {
	code, ok := keyCodes[key]
	if !ok {
		return nil, fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	load()
	fired := make(chan struct{}, 1)
	hotkeys.Lock()
	if hotkeys.fired == nil {
		hotkeys.fired = map[uint32]chan struct{}{}
		hotkeys.callback = purego.NewCallback(hotkeyEvent)
	}
	hotkeys.next++
	id := hotkeys.next
	hotkeys.fired[id] = fired
	hotkeys.Unlock()
	var result int32
	var ref uintptr
	onMain(func() {
		if hotkeys.handler == 0 {
			spec := [2]uint32{kEventClassKeyboard, kEventHotKeyPressed}
			result = int32(call("InstallEventHandler", call("GetApplicationEventTarget"), hotkeys.callback, 1, uintptr(unsafe.Pointer(&spec)), 0, uintptr(unsafe.Pointer(&hotkeys.handler))))
			if result != 0 {
				return
			}
		}
		var flags uintptr
		for bit, m := range carbonModifiers {
			if mods&uint(bit) != 0 {
				flags |= m
			}
		}
		// EventHotKeyID is passed by value, packed in one register.
		ident := uintptr(hotkeySignature) | uintptr(id)<<32
		result = int32(call("RegisterEventHotKey", uintptr(code), flags, ident, call("GetApplicationEventTarget"), 0, uintptr(unsafe.Pointer(&ref))))
		if result == 0 {
			hotkeys.count++
		} else if hotkeys.count == 0 {
			call("RemoveEventHandler", hotkeys.handler)
			hotkeys.handler = 0
		}
	})
	if result != 0 {
		hotkeys.Lock()
		delete(hotkeys.fired, id)
		hotkeys.Unlock()
		if result == eventHotKeyExistsErr {
			return nil, status(6)
		}
		return nil, status(int(result))
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
			var result int32
			onMain(func() {
				result = int32(call("UnregisterEventHotKey", ref))
				if result == 0 && hotkeys.count > 0 {
					hotkeys.count--
					if hotkeys.count == 0 {
						call("RemoveEventHandler", hotkeys.handler)
						hotkeys.handler = 0
					}
				}
			})
			err = status(int(result))
			hotkeys.Lock()
			delete(hotkeys.fired, id)
			close(fired)
			hotkeys.Unlock()
		})
		return err
	}, nil
}
