//go:build !(darwin && !ios) && !windows && !(linux && !android)

package window

// Other platforms retain the Go model and callback shortcuts. Query the
// capability to render an application menu using Go UI components if desired.
func platformNativeApplicationMenu() bool   { return false }
func platformInstallApplicationMenu([]byte) {}
func platformMenuEdit(MenuAction) bool      { return false }

func platformDrawApplicationMenu(*Window) bool { return false }
func applicationMenuWindowEvent(*Window, any)  {}

func platformNativeMenuShortcuts() bool { return false }
