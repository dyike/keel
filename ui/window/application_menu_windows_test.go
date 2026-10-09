//go:build windows

package window

import (
	"github.com/dyike/keel/ui/internal/loop"
	"golang.org/x/sys/windows"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// This target-platform test exercises real HMENU handles and WM_COMMAND on a
// hidden Win32 window, without requiring a GPU or displaying a test application.
func TestWin32MenuHandlesAndCommands(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	class, _ := windows.UTF16PtrFromString("STATIC")
	hwnd, _, err := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(class)), 0, 0, 0, 0, 32, 32, 0, 0, 0, 0)
	if hwnd == 0 {
		t.Fatal("CreateWindowExW", err)
	}
	defer user32.NewProc("DestroyWindow").Call(hwnd)
	calls := 0
	bar, e := NewMenuBar(MenuItem{ID: "file", Title: "File", Children: []MenuItem{
		{ID: "run", Title: "Run", Checked: true, Shortcut: "ctrl+t", OnSelect: func() { calls++ }},
		{ID: "blocked", Title: "Blocked", Disabled: true, OnSelect: func() { calls += 100 }},
		{Separator: true}, {ID: "sub", Title: "Nested", Children: []MenuItem{{ID: "nested", Title: "Nested Action", OnSelect: func() { calls += 10 }}}},
	}})
	if e != nil {
		t.Fatal(e)
	}
	applicationMenu.Lock()
	oldBar, oldGeneration := applicationMenu.current, applicationMenu.generation
	applicationMenu.current = bar
	applicationMenu.generation++
	generation := applicationMenu.generation
	applicationMenu.Unlock()
	defer func() {
		applicationMenu.Lock()
		applicationMenu.current = oldBar
		applicationMenu.generation = oldGeneration
		applicationMenu.Unlock()
	}()
	winMenus.Lock()
	oldModel := winMenus.model
	winMenus.model = menuWire{Generation: generation, Items: wireMenuItems(bar.Items())}
	winMenus.Unlock()
	defer func() { winMenus.Lock(); winMenus.model = oldModel; delete(winMenus.windows, hwnd); winMenus.Unlock() }()
	target := &nativeMenuWindow{hwnd: hwnd, w: newWindow(Options{})}
	winMenus.Lock()
	winMenus.windows[hwnd] = target
	winMenus.Unlock()
	proc := menuSetProc
	if unsafe.Sizeof(uintptr(0)) == 4 {
		proc = user32.NewProc("SetWindowLongW")
	}
	target.previous, _, _ = proc.Call(hwnd, ^uintptr(3), winMenuProc)
	if target.previous == 0 {
		t.Fatal("subclass failed")
	}
	defer proc.Call(hwnd, ^uintptr(3), target.previous)
	rebuildNativeMenu(target)
	installed, _, _ := user32.NewProc("GetMenu").Call(hwnd)
	if installed == 0 || installed != target.menu {
		t.Fatal("menu not attached")
	}
	submenu, _, _ := user32.NewProc("GetSubMenu").Call(installed, 0)
	count, _, _ := user32.NewProc("GetMenuItemCount").Call(submenu)
	if count != 4 {
		t.Fatalf("menu count %d", count)
	}
	checked, _, _ := user32.NewProc("GetMenuState").Call(submenu, 0, 0x400)
	disabled, _, _ := user32.NewProc("GetMenuState").Call(submenu, 1, 0x400)
	if checked&8 == 0 || disabled&3 == 0 {
		t.Fatalf("checked=%x disabled=%x", checked, disabled)
	}
	send := user32.NewProc("SendMessageW")
	for number := range target.commands {
		send.Call(hwnd, 0x111, number, 0)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		loop.Lock()
		loop.Drain()
		done := calls == 11
		loop.Unlock()
		if done {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if calls != 11 {
		t.Fatalf("native command dispatch calls=%d want=11", calls)
	}
	for _, tc := range []struct {
		name      string
		frameless bool
		display   MenuDisplay
		native    bool
	}{
		{"native", false, MenuDisplayAuto, true},
		{"frameless-auto", true, MenuDisplayAuto, false},
		{"window", false, MenuDisplayWindow, false},
		{"hidden", false, MenuDisplayHidden, false},
		{"native-restored", false, MenuDisplayAuto, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target.w.opts.Frameless = tc.frameless
			target.w.opts.MenuDisplay = tc.display
			rebuildNativeMenu(target)
			installed, _, _ := user32.NewProc("GetMenu").Call(hwnd)
			if (installed != 0) != tc.native {
				t.Fatalf("native menu=%x, want attached=%v", installed, tc.native)
			}
		})
	}
}
