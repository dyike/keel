//go:build windows

package sys

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/dyike/keel/native"
	"golang.org/x/sys/windows"
)

var (
	trayCreateIcon      = user32.NewProc("CreateIconFromResourceEx")
	trayDestroyIcon     = user32.NewProc("DestroyIcon")
	trayCreateMenu      = user32.NewProc("CreatePopupMenu")
	trayDestroyMenu     = user32.NewProc("DestroyMenu")
	trayAppendMenu      = user32.NewProc("AppendMenuW")
	trayTrackMenu       = user32.NewProc("TrackPopupMenuEx")
	trayForeground      = user32.NewProc("SetForegroundWindow")
	trayCursor          = user32.NewProc("GetCursorPos")
	trayRegisterMessage = user32.NewProc("RegisterWindowMessageW")
	trayWindowProc      = windows.NewCallback(trayProc)
	trayThread          struct {
		once     sync.Once
		ready    chan struct{}
		requests chan func()
		err      error
		hwnd     uintptr
		taskbar  uintptr
		next     uint32
		items    map[uint32]*windowsTray // locked OS thread only
	}
)

const (
	trayRequest      = 0x8000 + 30
	trayMessage      = 0x8000 + 31
	trayNinSelect    = 0x400
	trayNinKeySelect = 0x401
)

type windowsTray struct {
	id         uint32
	icon, menu uintptr
	options    TrayOptions
}

func trayOnThread(fn func()) error {
	trayThread.once.Do(func() {
		trayThread.ready = make(chan struct{})
		trayThread.requests = make(chan func(), 16)
		go trayLoop()
	})
	<-trayThread.ready
	if trayThread.err != nil {
		return trayThread.err
	}
	done := make(chan struct{})
	trayThread.requests <- func() { fn(); close(done) }
	postMessage.Call(trayThread.hwnd, trayRequest, 0, 0)
	<-done
	return nil
}

func trayLoop() {
	runtime.LockOSThread()
	// Keep the callback thread for the process lifetime; callbacks are allocated once.
	module, _, _ := getModuleHandle.Call(0)
	name, _ := windows.UTF16PtrFromString("KeelTrayWindow")
	wc := struct {
		size, style                   uint32
		proc                          uintptr
		clsExtra, wndExtra            int32
		instance, icon, cursor, brush uintptr
		menuName, className           *uint16
		smallIcon                     uintptr
	}{proc: trayWindowProc, instance: module, className: name}
	wc.size = uint32(unsafe.Sizeof(wc))
	if r, _, e := registerClassEx.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		trayThread.err = fmt.Errorf("%w: RegisterClassEx: %v", native.ErrFailed, e)
		close(trayThread.ready)
		return
	}
	// Hidden top-level HWND receives the shell's TaskbarCreated broadcast.
	hwnd, _, e := createWindowEx.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(name)), 0, 0, 0, 0, 0, 0, 0, module, 0)
	if hwnd == 0 {
		trayThread.err = fmt.Errorf("%w: CreateWindowEx: %v", native.ErrFailed, e)
		close(trayThread.ready)
		return
	}
	taskbar, _ := windows.UTF16PtrFromString("TaskbarCreated")
	trayThread.taskbar, _, _ = trayRegisterMessage.Call(uintptr(unsafe.Pointer(taskbar)))
	trayThread.hwnd = hwnd
	trayThread.items = make(map[uint32]*windowsTray)
	close(trayThread.ready)
	var msg struct {
		hwnd           uintptr
		message        uint32
		wParam, lParam uintptr
		time           uint32
		pt             struct{ x, y int32 }
		private        uint32
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

func TrayOpen(options TrayOptions) (TrayBackend, error) {
	var t *windowsTray
	var err error
	if e := trayOnThread(func() {
		if trayThread.next == 65535 {
			err = native.ErrFailed
			return
		}
		trayThread.next++
		t = &windowsTray{id: trayThread.next}
		err = t.apply(options, true)
		if err == nil {
			trayThread.items[t.id] = t
		}
	}); e != nil {
		return nil, e
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *windowsTray) data() *notifyIconData {
	d := &notifyIconData{hwnd: trayThread.hwnd, id: t.id, flags: nifMessage | nifIcon | nifTip | 0x80, callbackMessage: trayMessage, icon: t.icon}
	d.size = uint32(unsafe.Sizeof(*d))
	copyUTF16(d.tip[:], t.options.Tooltip)
	return d
}

func (t *windowsTray) add() error {
	if r, _, e := shellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(t.data()))); r == 0 {
		return fmt.Errorf("%w: Shell_NotifyIcon: %v", native.ErrFailed, e)
	}
	d := t.data()
	d.version = notifyIconVersion4
	shellNotifyIcon.Call(nimSetVersion, uintptr(unsafe.Pointer(d)))
	return nil
}

func (t *windowsTray) apply(options TrayOptions, add bool) error {
	icon, _, e := trayCreateIcon.Call(uintptr(unsafe.Pointer(&options.PNG[0])), uintptr(len(options.PNG)), 1, 0x30000, 32, 32, 0)
	if icon == 0 {
		return fmt.Errorf("%w: tray icon: %v", native.ErrFailed, e)
	}
	menu, err := makeWindowsTrayMenu(options.Menu)
	if err != nil {
		trayDestroyIcon.Call(icon)
		return err
	}
	oldIcon, oldMenu, oldOptions := t.icon, t.menu, t.options
	t.icon, t.menu, t.options = icon, menu, options
	if add {
		err = t.add()
	} else if r, _, e := shellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(t.data()))); r == 0 {
		err = fmt.Errorf("%w: tray update: %v", native.ErrFailed, e)
	}
	if err != nil {
		t.icon, t.menu, t.options = oldIcon, oldMenu, oldOptions
		trayDestroyIcon.Call(icon)
		if menu != 0 {
			trayDestroyMenu.Call(menu)
		}
		return err
	}
	if oldIcon != 0 {
		trayDestroyIcon.Call(oldIcon)
	}
	if oldMenu != 0 {
		trayDestroyMenu.Call(oldMenu)
	}
	return nil
}

func makeWindowsTrayMenu(items []TrayMenuItem) (uintptr, error) {
	if len(items) == 0 {
		return 0, nil
	}
	menu, _, e := trayCreateMenu.Call()
	if menu == 0 {
		return 0, fmt.Errorf("%w: CreatePopupMenu: %v", native.ErrFailed, e)
	}
	for _, item := range items {
		flags := uintptr(0)
		command := uintptr(item.ID)
		var child uintptr
		if item.Separator {
			flags |= 0x800
		}
		if item.Disabled {
			flags |= 3
		}
		if item.Checked {
			flags |= 8
		}
		if len(item.Children) != 0 {
			var err error
			child, err = makeWindowsTrayMenu(item.Children)
			if err != nil {
				trayDestroyMenu.Call(menu)
				return 0, err
			}
			flags |= 0x10
			command = child
		}
		title, _ := windows.UTF16PtrFromString(item.Title)
		if r, _, e := trayAppendMenu.Call(menu, flags, command, uintptr(unsafe.Pointer(title))); r == 0 {
			if child != 0 {
				trayDestroyMenu.Call(child)
			}
			trayDestroyMenu.Call(menu)
			return 0, fmt.Errorf("%w: AppendMenu: %v", native.ErrFailed, e)
		}
	}
	return menu, nil
}

func (t *windowsTray) Update(options TrayOptions) error {
	var err error
	if e := trayOnThread(func() { err = t.apply(options, false) }); e != nil {
		return e
	}
	return err
}

func (t *windowsTray) Close() error {
	return trayOnThread(func() {
		shellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(t.data())))
		delete(trayThread.items, t.id)
		if t.icon != 0 {
			trayDestroyIcon.Call(t.icon)
			t.icon = 0
		}
		if t.menu != 0 {
			trayDestroyMenu.Call(t.menu)
			t.menu = 0
		}
	})
}

func trayProc(hwnd, message, wParam, lParam uintptr) uintptr {
	if message == trayRequest {
		for {
			select {
			case fn := <-trayThread.requests:
				fn()
			default:
				return 0
			}
		}
	}
	if trayThread.taskbar != 0 && message == trayThread.taskbar {
		for _, t := range trayThread.items {
			_ = t.add()
		}
		return 0
	}
	if message == trayMessage {
		t := trayThread.items[uint32(lParam>>16)]
		if t == nil {
			return 0
		}
		switch lParam & 0xffff {
		case trayNinSelect, trayNinKeySelect:
			t.options.Event(0)
		case 0x7b: // WM_CONTEXTMENU, including keyboard activation
			if t.menu != 0 {
				var point struct{ x, y int32 }
				point.x, point.y = int32(int16(wParam&0xffff)), int32(int16((wParam>>16)&0xffff))
				if point.x == -1 && point.y == -1 {
					trayCursor.Call(uintptr(unsafe.Pointer(&point)))
				}
				trayForeground.Call(hwnd)
				command, _, _ := trayTrackMenu.Call(t.menu, 0x100|0x2, uintptr(point.x), uintptr(point.y), hwnd, 0)
				postMessage.Call(hwnd, 0, 0, 0) // shell dismissal rule
				if command != 0 {
					t.options.Event(uint32(command))
				}
			}
		}
		return 0
	}
	r, _, _ := defWindowProc.Call(hwnd, message, wParam, lParam)
	return r
}

func TrayMain() { select {} }
