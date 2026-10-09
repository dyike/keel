//go:build linux && !android

package window

// Gio's X11 and Wayland surfaces use Keel-rendered in-window menus.
func platformNativeApplicationMenu() bool      { return false }
func platformDrawApplicationMenu(*Window) bool { return true }
func platformInstallApplicationMenu([]byte)    {}
func platformMenuEdit(action MenuAction) bool  { return requestMenuEdit(action) }
func applicationMenuWindowEvent(*Window, any)  {}

func platformNativeMenuShortcuts() bool { return false }
