//go:build windows

package sys

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dyike/keel/native"
)

// Windows needs no privacy grants for these APIs; coordinates are logical
// points scaled by the primary display's DPI, as on macOS.

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	gdi32  = windows.NewLazySystemDLL("gdi32.dll")
	shcore = windows.NewLazySystemDLL("shcore.dll")

	enumDisplayMonitors    = user32.NewProc("EnumDisplayMonitors")
	getMonitorInfo         = user32.NewProc("GetMonitorInfoW")
	getCursorPos           = user32.NewProc("GetCursorPos")
	setCursorPos           = user32.NewProc("SetCursorPos")
	sendInput              = user32.NewProc("SendInput")
	getDC                  = user32.NewProc("GetDC")
	releaseDC              = user32.NewProc("ReleaseDC")
	registerHotKey         = user32.NewProc("RegisterHotKey")
	unregisterHotKey       = user32.NewProc("UnregisterHotKey")
	getMessage             = user32.NewProc("GetMessageW")
	postThreadMessage      = user32.NewProc("PostThreadMessageW")
	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject           = gdi32.NewProc("SelectObject")
	bitBlt                 = gdi32.NewProc("BitBlt")
	getDIBits              = gdi32.NewProc("GetDIBits")
	deleteObject           = gdi32.NewProc("DeleteObject")
	deleteDC               = gdi32.NewProc("DeleteDC")
	getDpiForMonitor       = shcore.NewProc("GetDpiForMonitor")
)

func Permission(kind int, request bool) (bool, error) { return true, nil }

type rect struct{ left, top, right, bottom int32 }

type monitorInfo struct {
	size          uint32
	monitor, work rect
	flags         uint32
}

type monitor struct {
	handle  uintptr
	bounds  rect
	primary bool
	scale   float64
}

// The enumeration callback is made once: Windows allows a process only a
// limited number of callbacks and never frees them.
var (
	monitorMu   sync.Mutex
	monitorList []monitor
	monitorProc = syscall.NewCallback(func(h, _, _, _ uintptr) uintptr {
		info := monitorInfo{size: uint32(unsafe.Sizeof(monitorInfo{}))}
		if r, _, _ := getMonitorInfo.Call(h, uintptr(unsafe.Pointer(&info))); r == 0 {
			return 1
		}
		scale := 1.0
		if getDpiForMonitor.Find() == nil {
			var dx, dy uint32
			if r, _, _ := getDpiForMonitor.Call(h, 0, uintptr(unsafe.Pointer(&dx)), uintptr(unsafe.Pointer(&dy))); r == 0 && dx > 0 {
				scale = float64(dx) / 96
			}
		}
		monitorList = append(monitorList, monitor{h, info.monitor, info.flags&1 != 0, scale})
		return 1
	})
)

// monitors lists the displays in the order Windows enumerates them.
func monitors() ([]monitor, error) {
	monitorMu.Lock()
	defer monitorMu.Unlock()
	monitorList = nil
	if r, _, err := enumDisplayMonitors.Call(0, 0, monitorProc, 0); r == 0 {
		return nil, fmt.Errorf("%w: EnumDisplayMonitors: %v", native.ErrFailed, err)
	}
	return monitorList, nil
}

// primaryScale is the primary display's pixels per point.
func primaryScale() float64 {
	ms, _ := monitors()
	for _, m := range ms {
		if m.primary {
			return m.scale
		}
	}
	return 1
}

func Displays() ([]Display, error) {
	ms, err := monitors()
	if err != nil {
		return nil, err
	}
	s := primaryScale()
	out := make([]Display, len(ms))
	for i, m := range ms {
		b := m.bounds
		out[i] = Display{ID: uint32(i + 1), X: float64(b.left) / s, Y: float64(b.top) / s,
			Width: float64(b.right-b.left) / m.scale, Height: float64(b.bottom-b.top) / m.scale,
			PixelWidth: int(b.right - b.left), PixelHeight: int(b.bottom - b.top), Primary: m.primary}
	}
	return out, nil
}

type bitmapInfoHeader struct {
	size                   uint32
	width, height          int32
	planes, bitCount       uint16
	compression, sizeImage uint32
	xPels, yPels           int32
	clrUsed, clrImportant  uint32
}

// Capture copies the display's pixels with GDI and encodes them as PNG.
func Capture(id uint32) ([]byte, error) {
	ms, err := monitors()
	if err != nil {
		return nil, err
	}
	if id == 0 || int(id) > len(ms) {
		return nil, native.ErrInvalidArgument
	}
	b := ms[id-1].bounds
	w, h := int(b.right-b.left), int(b.bottom-b.top)
	screen, _, _ := getDC.Call(0)
	if screen == 0 {
		return nil, fmt.Errorf("%w: GetDC", native.ErrFailed)
	}
	defer releaseDC.Call(0, screen)
	mem, _, _ := createCompatibleDC.Call(screen)
	defer deleteDC.Call(mem)
	bmp, _, _ := createCompatibleBitmap.Call(screen, uintptr(w), uintptr(h))
	if bmp == 0 {
		return nil, fmt.Errorf("%w: CreateCompatibleBitmap", native.ErrFailed)
	}
	defer deleteObject.Call(bmp)
	old, _, _ := selectObject.Call(mem, bmp)
	const srcCopy, captureBlt = 0x00CC0020, 0x40000000
	r, _, _ := bitBlt.Call(mem, 0, 0, uintptr(w), uintptr(h), screen, uintptr(b.left), uintptr(b.top), srcCopy|captureBlt)
	selectObject.Call(mem, old)
	if r == 0 {
		return nil, fmt.Errorf("%w: BitBlt", native.ErrFailed)
	}
	hdr := bitmapInfoHeader{size: uint32(unsafe.Sizeof(bitmapInfoHeader{})), width: int32(w), height: -int32(h), planes: 1, bitCount: 32}
	pix := make([]byte, w*h*4)
	if r, _, _ := getDIBits.Call(mem, bmp, 0, uintptr(h), uintptr(unsafe.Pointer(&pix[0])), uintptr(unsafe.Pointer(&hdr)), 0); r == 0 {
		return nil, fmt.Errorf("%w: GetDIBits", native.ErrFailed)
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(pix); i += 4 { // BGRA to RGBA, opaque
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = pix[i+2], pix[i+1], pix[i], 0xff
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("%w: %v", native.ErrFailed, err)
	}
	return out.Bytes(), nil
}

func MousePosition() (float64, float64, error) {
	var p struct{ x, y int32 }
	if r, _, err := getCursorPos.Call(uintptr(unsafe.Pointer(&p))); r == 0 {
		return 0, 0, fmt.Errorf("%w: GetCursorPos: %v", native.ErrFailed, err)
	}
	s := primaryScale()
	return float64(p.x) / s, float64(p.y) / s, nil
}

func MouseMove(x, y float64) error {
	s := primaryScale()
	if r, _, err := setCursorPos.Call(uintptr(int32(x*s)), uintptr(int32(y*s))); r == 0 {
		return fmt.Errorf("%w: SetCursorPos: %v", native.ErrFailed, err)
	}
	return nil
}

// SendInput structures. Go aligns the union member like C does, so the
// mouse form has INPUT's size on 32- and 64-bit Windows; the keyboard form
// pads to it.
type mouseInput struct {
	dx, dy                 int32
	mouseData, flags, time uint32
	extra                  uintptr
}
type inputMouse struct {
	typ uint32
	mi  mouseInput
}
type keybdInput struct {
	vk, scan    uint16
	flags, time uint32
	extra       uintptr
}
type inputKey struct {
	typ uint32
	ki  keybdInput
	_   [8]byte
}

func send(p unsafe.Pointer, n int) error {
	if r, _, err := sendInput.Call(uintptr(n), uintptr(p), unsafe.Sizeof(inputMouse{})); int(r) != n {
		return fmt.Errorf("%w: SendInput: %v", native.ErrFailed, err)
	}
	return nil
}

func Click(button int) error {
	flags := [][2]uint32{{0x2, 0x4}, {0x8, 0x10}, {0x20, 0x40}} // left, right, middle: down, up
	if button < 0 || button >= len(flags) {
		return native.ErrInvalidArgument
	}
	in := [2]inputMouse{{mi: mouseInput{flags: flags[button][0]}}, {mi: mouseInput{flags: flags[button][1]}}}
	return send(unsafe.Pointer(&in[0]), 2)
}

func HasKey(key string) bool { _, ok := virtualKeys[key]; return ok }

func Key(key string, down bool) error {
	vk, ok := virtualKeys[key]
	if !ok {
		return fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	var flags uint32
	if extendedKeys[key] {
		flags |= 0x1 // KEYEVENTF_EXTENDEDKEY
	}
	if !down {
		flags |= 0x2 // KEYEVENTF_KEYUP
	}
	in := inputKey{typ: 1, ki: keybdInput{vk: vk, flags: flags}}
	return send(unsafe.Pointer(&in), 1)
}

// virtualKeys maps Keel key names to Windows virtual-key codes, the same
// names as on macOS.
var virtualKeys = func() map[string]uint16 {
	m := map[string]uint16{
		"enter": 0x0D, "tab": 0x09, "space": 0x20, "backspace": 0x08, "escape": 0x1B,
		"shift": 0x10, "ctrl": 0x11, "alt": 0x12, "cmd": 0x5B,
		"home": 0x24, "end": 0x23, "pageup": 0x21, "pagedown": 0x22, "delete": 0x2E,
		"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
		";": 0xBA, "=": 0xBB, ",": 0xBC, "-": 0xBD, ".": 0xBE, "/": 0xBF, "`": 0xC0,
		"[": 0xDB, "\\": 0xDC, "]": 0xDD, "'": 0xDE,
	}
	for c := 'a'; c <= 'z'; c++ {
		m[string(c)] = uint16('A' + c - 'a')
	}
	for c := '0'; c <= '9'; c++ {
		m[string(c)] = uint16(c)
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("f%d", i)] = uint16(0x70 + i - 1)
	}
	return m
}()

var extendedKeys = map[string]bool{"home": true, "end": true, "pageup": true, "pagedown": true, "delete": true,
	"left": true, "up": true, "right": true, "down": true, "cmd": true}

// Hotkeys: RegisterHotKey delivers WM_HOTKEY to the registering thread, so
// one locked thread registers every hotkey and runs their message loop.
var hk struct {
	sync.Mutex
	once     sync.Once
	thread   uint32
	ready    chan struct{}
	requests chan func()
	next     uintptr
	fired    map[uintptr]chan struct{}
}

const (
	wmHotkey = 0x0312
	wmApp    = 0x8000
)

func hotkeyLoop() {
	runtime.LockOSThread()
	hk.thread = windows.GetCurrentThreadId()
	close(hk.ready)
	var msg struct {
		hwnd    uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
		_       uint32
	}
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		switch msg.message {
		case wmHotkey:
			hk.Lock()
			if ch := hk.fired[msg.wParam]; ch != nil {
				select {
				case ch <- struct{}{}:
				default:
				}
			}
			hk.Unlock()
		case wmApp:
			for {
				select {
				case f := <-hk.requests:
					f()
					continue
				default:
				}
				break
			}
		}
	}
}

// onHotkeyThread runs f on the hotkey thread and waits for it.
func onHotkeyThread(f func()) {
	hk.once.Do(func() {
		hk.ready = make(chan struct{})
		hk.requests = make(chan func(), 16)
		hk.fired = map[uintptr]chan struct{}{}
		go hotkeyLoop()
	})
	<-hk.ready
	done := make(chan struct{})
	hk.requests <- func() { f(); close(done) }
	postThreadMessage.Call(uintptr(hk.thread), wmApp, 0, 0)
	<-done
}

// Hotkey registers key+mods; fn runs on its own goroutine and presses that
// arrive while it runs are merged.
func Hotkey(key string, mods uint, fn func()) (func() error, error) {
	vk, ok := virtualKeys[key]
	if !ok {
		return nil, fmt.Errorf("%w: unknown key %q", native.ErrInvalidArgument, key)
	}
	var m uintptr = 0x4000 // MOD_NOREPEAT
	if mods&ModAlt != 0 {
		m |= 0x1
	}
	if mods&ModCtrl != 0 {
		m |= 0x2
	}
	if mods&ModShift != 0 {
		m |= 0x4
	}
	if mods&ModCmd != 0 {
		m |= 0x8 // the Windows key
	}
	fired := make(chan struct{}, 1)
	var id uintptr
	var regErr error
	onHotkeyThread(func() {
		hk.Lock()
		hk.next++
		id = hk.next
		hk.fired[id] = fired
		hk.Unlock()
		if r, _, err := registerHotKey.Call(0, id, m, uintptr(vk)); r == 0 {
			hk.Lock()
			delete(hk.fired, id)
			hk.Unlock()
			if errors.Is(err, syscall.Errno(1409)) { // ERROR_HOTKEY_ALREADY_REGISTERED
				regErr = native.ErrConflict
			} else {
				regErr = fmt.Errorf("%w: RegisterHotKey: %v", native.ErrFailed, err)
			}
		}
	})
	if regErr != nil {
		return nil, regErr
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
			onHotkeyThread(func() {
				if r, _, e := unregisterHotKey.Call(0, id); r == 0 {
					err = fmt.Errorf("%w: UnregisterHotKey: %v", native.ErrFailed, e)
				}
				hk.Lock()
				delete(hk.fired, id)
				hk.Unlock()
			})
			close(fired)
		})
		return err
	}, nil
}
