//go:build windows

package sys

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dyike/keel/native"
)

// Windows notifications are tray balloons (Shell_NotifyIcon with NIF_INFO),
// which Windows 10 and 11 present as system toasts and which need no
// AppUserModelID or packaging. The tray icon exists only while a balloon is
// up. One locked thread owns the callback window and runs its message loop.

var (
	shell32          = windows.NewLazySystemDLL("shell32.dll")
	shellNotifyIcon  = shell32.NewProc("Shell_NotifyIconW")
	registerClassEx  = user32.NewProc("RegisterClassExW")
	createWindowEx   = user32.NewProc("CreateWindowExW")
	defWindowProc    = user32.NewProc("DefWindowProcW")
	dispatchMessage  = user32.NewProc("DispatchMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	postMessage      = user32.NewProc("PostMessageW")
	loadIcon         = user32.NewProc("LoadIconW")
	getModuleHandle  = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW")
	balloons         = &balloonCenter{}
	balloonWindow    struct {
		once     sync.Once
		hwnd     uintptr
		err      error
		ready    chan struct{}
		requests chan func()
		shown    bool   // the tray icon is added
		posted   uint64 // the center's seq of the last balloon handed to the shell
		seq      uint64 // the balloon the shell last reported showing
	}
)

const (
	wmTray       = wmApp + 1
	wmRequest    = wmApp + 2
	wmLButtonUp  = 0x0202
	ninShow      = 0x0402
	ninHide      = 0x0403
	ninTimeout   = 0x0404
	ninUserClick = 0x0405

	nimAdd, nimModify, nimDelete, nimSetVersion = 0, 1, 2, 4
	nifMessage, nifIcon, nifTip, nifInfo        = 0x1, 0x2, 0x4, 0x10
	niifInfo                                    = 0x1
	notifyIconVersion4                          = 4
)

type notifyIconData struct {
	size            uint32
	hwnd            uintptr
	id              uint32
	flags           uint32
	callbackMessage uint32
	icon            uintptr
	tip             [128]uint16
	state           uint32
	stateMask       uint32
	info            [256]uint16
	version         uint32 // union with uTimeout
	infoTitle       [64]uint16
	infoFlags       uint32
	guid            windows.GUID
	balloonIcon     uintptr
}

type windowsBalloonShell struct{}

func (windowsBalloonShell) show(title, body string) error {
	var err error
	onBalloonThread(func() {
		w := &balloonWindow
		d := trayData(nifInfo | nifMessage | nifIcon | nifTip)
		copyUTF16(d.infoTitle[:], title)
		copyUTF16(d.info[:], body)
		copyUTF16(d.tip[:], title)
		d.infoFlags = niifInfo
		if !w.shown {
			if r, _, e := shellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(d))); r == 0 {
				err = fmt.Errorf("%w: Shell_NotifyIcon add: %v", native.ErrFailed, e)
				return
			}
			v := trayData(0)
			v.version = notifyIconVersion4
			shellNotifyIcon.Call(nimSetVersion, uintptr(unsafe.Pointer(v)))
			w.shown = true
		} else if r, _, e := shellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(d))); r == 0 {
			err = fmt.Errorf("%w: Shell_NotifyIcon modify: %v", native.ErrFailed, e)
			return
		}
		w.posted = balloons.seq + 1 // post increments it after show succeeds
	})
	return err
}

func (windowsBalloonShell) hide() error {
	// Called with the center locked, never on the window thread.
	hideTray := func() {
		if balloonWindow.shown {
			shellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(trayData(0))))
			balloonWindow.shown = false
		}
	}
	onBalloonThread(hideTray)
	return nil
}

func trayData(flags uint32) *notifyIconData {
	d := &notifyIconData{hwnd: balloonWindow.hwnd, id: 1, flags: flags, callbackMessage: wmTray}
	d.size = uint32(unsafe.Sizeof(*d))
	if flags&nifIcon != 0 {
		module, _, _ := getModuleHandle.Call(0)
		// The executable's first icon resource, else the generic app icon.
		if d.icon, _, _ = loadIcon.Call(module, 1); d.icon == 0 {
			d.icon, _, _ = loadIcon.Call(0, 32512) // IDI_APPLICATION
		}
	}
	return d
}

func copyUTF16(dst []uint16, s string) {
	u, _ := windows.UTF16FromString(s)
	n := copy(dst[:len(dst)-1], u)
	dst[n] = 0
}

var balloonThreadID uint32

func balloonOnThread() bool { return windows.GetCurrentThreadId() == balloonThreadID }

// onBalloonThread runs f on the window thread and waits for it.
func onBalloonThread(f func()) {
	w := &balloonWindow
	w.once.Do(func() {
		w.ready = make(chan struct{})
		w.requests = make(chan func(), 16)
		go balloonLoop()
	})
	<-w.ready
	if w.err != nil {
		return
	}
	if balloonOnThread() {
		f()
		return
	}
	done := make(chan struct{})
	w.requests <- func() { f(); close(done) }
	postMessage.Call(w.hwnd, wmRequest, 0, 0)
	<-done
}

func balloonLoop() {
	runtime.LockOSThread()
	w := &balloonWindow
	balloonThreadID = windows.GetCurrentThreadId()
	module, _, _ := getModuleHandle.Call(0)
	name, _ := windows.UTF16PtrFromString("KeelNotificationWindow")
	class := struct {
		size, style                   uint32
		proc                          uintptr
		clsExtra, wndExtra            int32
		instance, icon, cursor, brush uintptr
		menuName, className           *uint16
		smallIcon                     uintptr
	}{proc: windows.NewCallback(balloonProc), instance: module, className: name}
	class.size = uint32(unsafe.Sizeof(class))
	registerClassEx.Call(uintptr(unsafe.Pointer(&class)))
	// A hidden top-level window: tray callbacks reach it like any window.
	hwnd, _, e := createWindowEx.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(name)), 0, 0, 0, 0, 0, 0, 0, module, 0)
	if hwnd == 0 {
		w.err = fmt.Errorf("%w: CreateWindowEx: %v", native.ErrFailed, e)
		close(w.ready)
		return
	}
	w.hwnd = hwnd
	close(w.ready)
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
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func balloonProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmRequest:
		for {
			select {
			case f := <-balloonWindow.requests:
				f()
				continue
			default:
			}
			return 0
		}
	case wmTray:
		// NOTIFYICON_VERSION_4 puts the event in the low word of lParam. A
		// replaced balloon's hide can arrive after its successor was posted,
		// so events name the balloon the shell last showed. The center is
		// updated off this thread: a post holding it may be waiting on us.
		w := &balloonWindow
		var e balloonEvent
		switch lParam & 0xffff {
		case ninShow:
			w.seq = w.posted
			return 0
		case ninUserClick, wmLButtonUp:
			e = balloonClicked
		case ninTimeout, ninHide:
			e = balloonClosed
		default:
			return 0
		}
		seq := w.seq
		go func() {
			if fn := balloons.event(e, seq); fn != nil {
				fn()
			}
		}()
		return 0
	}
	r, _, _ := defWindowProc.Call(hwnd, msg, wParam, lParam)
	return r
}

func init() { balloons.shell = windowsBalloonShell{} }

func NotificationAvailable() bool {
	onBalloonThread(func() {})
	return balloonWindow.err == nil
}

// Balloons need no authorization; this only makes sure the window exists.
func NotificationPermission(done func(error)) {
	onBalloonThread(func() {})
	done(balloonWindow.err)
}

func NotificationPost(id, title, body string, done func(error)) {
	NotificationPostInteractive(id, title, body, nil, done)
}

func NotificationPostInteractive(id, title, body string, onClick func(), done func(error)) {
	if !NotificationAvailable() {
		done(balloonWindow.err)
		return
	}
	done(balloons.post(id, title, body, onClick))
}

func NotificationRemove(id string, done func(error)) { done(balloons.remove(id)) }
