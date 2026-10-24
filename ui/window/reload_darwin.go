//go:build darwin && !ios

package window

import "github.com/dyike/keel/ui/internal/appkit"

// windowOrigin asks AppKit for the frame origin of view's window. It waits
// for the main thread: never call it under the frame lock.
func windowOrigin(view uintptr) *[2]float64 {
	var origin *[2]float64
	appkit.MainSync(func() {
		window := appkit.Send(appkit.ID(view), "window")
		if window == 0 || appkit.SendBool(window, "isMiniaturized") || uintptr(appkit.Send(window, "styleMask"))&nsWindowStyleMaskFullScreen != 0 {
			return
		}
		frame := appkit.MsgRect(window, appkit.Sel("frame"))
		origin = &[2]float64{frame.Origin.X, frame.Origin.Y}
	})
	return origin
}
