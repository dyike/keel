//go:build darwin && !ios

package window

import "github.com/dyike/keel/ui/internal/appkit"

const nsWindowStyleMaskFullScreen = 1 << 14

// Called once after the first frame: Gio has finished configuring and
// cascading the native window by then. AppKit work is queued, never awaited
// under the frame lock.
func centerNewWindow(w *Window) {
	if w.nativeView == 0 {
		return
	}
	// Gio keeps the view valid until the next view event; retain it across
	// the asynchronous dispatch.
	view := appkit.Retain(appkit.ID(w.nativeView))
	appkit.MainAsync(func() {
		defer appkit.Release(view)
		window := appkit.Send(view, "window")
		if window == 0 {
			return
		}
		screen := appkit.Send(window, "screen")
		if screen == 0 {
			screen = appkit.Send(appkit.Class("NSScreen"), "mainScreen")
		}
		if screen == 0 || appkit.SendBool(window, "isMiniaturized") || uintptr(appkit.Send(window, "styleMask"))&nsWindowStyleMaskFullScreen != 0 {
			return
		}
		visible := appkit.MsgRect(screen, appkit.Sel("visibleFrame"))
		frame := appkit.MsgRect(window, appkit.Sel("frame"))
		// Include the title bar, preserve the size and the screen origin.
		// Oversized windows keep their top and left edges reachable.
		origin := appkit.Point{
			X: visible.Origin.X + max(0, (visible.Size.Width-frame.Size.Width)/2),
			Y: visible.MaxY() - frame.Size.Height - max(0, (visible.Size.Height-frame.Size.Height)/2),
		}
		appkit.MsgSetPoint(window, appkit.Sel("setFrameOrigin:"), origin)
	})
}
