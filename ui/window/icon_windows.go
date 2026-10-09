//go:build windows

package window

import (
	"image"
	"sync"
	"unsafe"

	gioapp "github.com/dyike/keel/third_party/gio/app"
	"golang.org/x/sys/windows"

	"github.com/dyike/keel/internal/appicon"
)

var (
	user32                   = windows.NewLazySystemDLL("user32.dll")
	createIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	postMessage              = user32.NewProc("PostMessageW")
	getDpiForWindow          = user32.NewProc("GetDpiForWindow")
	getSystemMetricsForDpi   = user32.NewProc("GetSystemMetricsForDpi")
	getSystemMetrics         = user32.NewProc("GetSystemMetrics")
	winIcons                 struct {
		sync.Mutex
		hwnds map[*Window]uintptr
	}
)

const (
	wmSetIcon  = 0x0080
	iconSmall  = 0
	iconBig    = 1
	smCxIcon   = 11
	smCxSmIcon = 49
	iconResVer = 0x00030000
)

func platformSetIcon(art image.Image, finished bool) {
	winIcons.Lock()
	defer winIcons.Unlock()
	for _, hwnd := range winIcons.hwnds {
		setWindowIcon(hwnd, art, finished)
	}
}

// iconWindowEvent keeps each window's HWND and gives new windows the icon.
func iconWindowEvent(w *Window, e any) {
	ev, ok := e.(gioapp.Win32ViewEvent)
	if !ok {
		return
	}
	winIcons.Lock()
	defer winIcons.Unlock()
	if winIcons.hwnds == nil {
		winIcons.hwnds = map[*Window]uintptr{}
	}
	if ev.HWND == 0 {
		delete(winIcons.hwnds, w)
		return
	}
	winIcons.hwnds[w] = ev.HWND
	if art, finished := currentIconState(); art != nil {
		setWindowIcon(ev.HWND, art, finished)
	}
}

// setWindowIcon sets the small (title bar) and big (taskbar, Alt+Tab)
// icons at the window's DPI, each drawn at its size on Microsoft's grid.
// It posts rather than sends WM_SETICON: callers may hold the frame lock
// the window's thread is waiting for.
func setWindowIcon(hwnd uintptr, art image.Image, finished bool) {
	dpi := uintptr(96)
	if getDpiForWindow.Find() == nil {
		if d, _, _ := getDpiForWindow.Call(hwnd); d != 0 {
			dpi = d
		}
	}
	metric := func(index uintptr, fallback int) int {
		if getSystemMetricsForDpi.Find() == nil {
			if v, _, _ := getSystemMetricsForDpi.Call(index, dpi); v != 0 {
				return int(v)
			}
		}
		if v, _, _ := getSystemMetrics.Call(index); v != 0 {
			return int(v)
		}
		return fallback
	}
	for _, icon := range []struct {
		which uintptr
		size  int
	}{{iconSmall, metric(smCxSmIcon, 16)}, {iconBig, metric(smCxIcon, 32)}} {
		data := encodePNG(renderIcon(art, appicon.Windows, icon.size, finished))
		h, _, _ := createIconFromResourceEx.Call(uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), 1, iconResVer, uintptr(icon.size), uintptr(icon.size), 0)
		if h != 0 {
			postMessage.Call(hwnd, wmSetIcon, icon.which, h)
		}
	}
}
