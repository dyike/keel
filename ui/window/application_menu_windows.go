//go:build windows

package window

import (
	"encoding/json"
	gioapp "github.com/dyike/keel/gio/app"
	"github.com/dyike/keel/ui/core"
	"golang.org/x/sys/windows"
	"sync"
	"unsafe"
)

var (
	menuCreate   = user32.NewProc("CreateMenu")
	menuPopup    = user32.NewProc("CreatePopupMenu")
	menuAppend   = user32.NewProc("AppendMenuW")
	menuSet      = user32.NewProc("SetMenu")
	menuDestroy  = user32.NewProc("DestroyMenu")
	menuDraw     = user32.NewProc("DrawMenuBar")
	menuSetProc  = user32.NewProc("SetWindowLongPtrW")
	menuCallProc = user32.NewProc("CallWindowProcW")
	menuDefProc  = user32.NewProc("DefWindowProcW")
	winMenuProc  = windows.NewCallback(menuWindowProc)
	winMenus     = struct {
		sync.Mutex
		model   menuWire
		windows map[uintptr]*nativeMenuWindow
	}{windows: map[uintptr]*nativeMenuWindow{}}
)

type nativeMenuCommand struct {
	generation uint64
	id         string
}
type nativeMenuWindow struct {
	w                    *Window
	hwnd, previous, menu uintptr
	commands             map[uintptr]nativeMenuCommand // native thread only
	bitmaps              []uintptr                     // released after the menu that refers to them
}

func platformMenuFeatures() MenuFeatures {
	return MenuFeatures{Native: true, Icons: true, CancelTracking: true}
}
func platformCancelMenuTracking(generation uint64, animated bool) {
	winMenus.Lock()
	defer winMenus.Unlock()
	if winMenus.model.Generation == generation {
		for hwnd := range winMenus.windows {
			postMessage.Call(hwnd, 0x1f, 0, 0) // WM_CANCELMODE; no main-thread wait.
		}
	}
}

func platformNativeApplicationMenu() bool        { return true }
func platformDrawApplicationMenu(w *Window) bool { return w.opts.Frameless }
func platformMenuEdit(action MenuAction) bool    { return requestMenuEdit(action) }

func platformInstallApplicationMenu(data []byte) {
	var model menuWire
	if json.Unmarshal(data, &model) != nil {
		return
	}
	winMenus.Lock()
	winMenus.model = model
	targets := make([]*nativeMenuWindow, 0, len(winMenus.windows))
	for _, target := range winMenus.windows {
		targets = append(targets, target)
	}
	winMenus.Unlock()
	for _, target := range targets {
		refreshNativeMenu(target)
	}
}

func applicationMenuWindowEvent(w *Window, e any) {
	ev, ok := e.(gioapp.Win32ViewEvent)
	if !ok {
		return
	}
	if ev.HWND == 0 {
		return
	} // WM_NCDESTROY releases thread-owned resources.
	target := &nativeMenuWindow{w: w, hwnd: ev.HWND}
	winMenus.Lock()
	if winMenus.windows[ev.HWND] != nil {
		winMenus.Unlock()
		return
	}
	winMenus.windows[ev.HWND] = target
	winMenus.Unlock()
	// Run waits for the native thread: never call it while holding a frame lock.
	go w.win.Run(func() {
		proc := menuSetProc
		if unsafe.Sizeof(uintptr(0)) == 4 {
			proc = user32.NewProc("SetWindowLongW")
		}
		target.previous, _, _ = proc.Call(target.hwnd, ^uintptr(3), winMenuProc) // GWLP_WNDPROC = -4
		if target.previous == 0 {
			winMenus.Lock()
			delete(winMenus.windows, target.hwnd)
			winMenus.Unlock()
			return
		}
		rebuildNativeMenu(target)
	})
}

func refreshNativeMenu(target *nativeMenuWindow) {
	go func() {
		if target.w.isClosed() {
			return
		}
		target.w.win.Run(func() {
			winMenus.Lock()
			current := winMenus.windows[target.hwnd] == target
			winMenus.Unlock()
			if current && target.previous != 0 {
				rebuildNativeMenu(target)
			}
		})
	}()
}

func rebuildNativeMenu(target *nativeMenuWindow) {
	winMenus.Lock()
	model := winMenus.model
	winMenus.Unlock()
	if target.w.opts.MenuDisplay != MenuDisplayAuto || platformDrawApplicationMenu(target.w) {
		model.Items = nil
	}
	commands := map[uintptr]nativeMenuCommand{}
	var bitmaps []uintptr
	committed := false
	defer func() {
		if !committed {
			freeMenuBitmaps(bitmaps)
		}
	}()
	next := uintptr(1)
	var build func([]menuWireItem, bool, bool) uintptr
	build = func(items []menuWireItem, popup, blocked bool) uintptr {
		proc := menuCreate
		if popup {
			proc = menuPopup
		}
		handle, _, _ := proc.Call()
		if handle == 0 {
			return 0
		}
		for position, item := range items {
			flags := uintptr(0)
			id := next
			next++
			disabled := blocked || item.Disabled
			if disabled {
				flags |= 0x3
			}
			if item.Checked {
				flags |= 0x8
			}
			title := item.Title
			if len(item.Children) > 0 {
				flags |= 0x10
				id = build(item.Children, true, disabled)
				if id == 0 {
					menuDestroy.Call(handle)
					return 0
				}
			} else if item.Separator {
				flags = 0x800
				id = 0
			} else {
				commands[id] = nativeMenuCommand{model.Generation, item.ID}
				if item.Key != "" {
					title += "\t" + menuWireShortcut(item)
				}
			}
			value, err := windows.UTF16PtrFromString(title)
			if err != nil {
				menuDestroy.Call(handle)
				return 0
			}
			ok, _, _ := menuAppend.Call(handle, flags, id, uintptr(unsafe.Pointer(value)))
			if ok == 0 {
				if flags&0x10 != 0 {
					menuDestroy.Call(id)
				}
				menuDestroy.Call(handle)
				return 0
			}
			if len(item.PNG) > 0 {
				bitmap, err := createMenuBitmap(item.PNG, nativeMenuIconSize(target.hwnd))
				if err != nil {
					menuDestroy.Call(handle)
					return 0
				}
				bitmaps = append(bitmaps, bitmap)
				if !setMenuBitmap(handle, uint32(position), bitmap) {
					menuDestroy.Call(handle)
					return 0
				}
			}
		}
		return handle
	}
	handle := build(model.Items, false, false)
	if handle == 0 {
		return
	}
	if len(model.Items) == 0 {
		menuDestroy.Call(handle)
		handle = 0
	}
	ok, _, _ := menuSet.Call(target.hwnd, handle)
	if ok == 0 {
		if handle != 0 {
			menuDestroy.Call(handle)
		}
		return
	}
	old, oldBitmaps := target.menu, target.bitmaps
	target.menu = handle
	target.bitmaps = bitmaps
	committed = true
	target.commands = commands
	menuDraw.Call(target.hwnd)
	if old != 0 {
		menuDestroy.Call(old)
	}
	freeMenuBitmaps(oldBitmaps)
}

func menuWindowProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	winMenus.Lock()
	target := winMenus.windows[hwnd]
	winMenus.Unlock()
	if target == nil {
		result, _, _ := menuDefProc.Call(hwnd, msg, wparam, lparam)
		return result
	}
	if msg == 0x111 && lparam == 0 { // WM_COMMAND: menus and accelerators only.
		if command, ok := target.commands[wparam&0xffff]; ok {
			go core.Update(func() {
				if !target.w.closed {
					defer core.SetCurrentWindow(target.w)()
					invokeNativeMenu(command.generation, command.id)
				}
			})
			return 0
		}
	}
	if msg == 0x82 { // WM_NCDESTROY; destroy detached old menus, current menu is OS-owned.
		winMenus.Lock()
		delete(winMenus.windows, hwnd)
		winMenus.Unlock()
	}
	result, _, _ := menuCallProc.Call(target.previous, hwnd, msg, wparam, lparam)
	if msg == 0x82 {
		freeMenuBitmaps(target.bitmaps)
		target.bitmaps = nil
	}
	return result
}

func platformNativeMenuShortcuts() bool { return false }

func nativeMenuIconSize(hwnd uintptr) int {
	dpi := uintptr(96)
	if getDpiForWindow.Find() == nil {
		if n, _, _ := getDpiForWindow.Call(hwnd); n != 0 {
			dpi = n
		}
	}
	return max(1, int(16*dpi/96))
}
