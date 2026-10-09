//go:build darwin && !ios

package main

import (
	"fmt"
	"math"
	"os"

	"github.com/dyike/keel/ui/internal/appkit"
)

// Actions of checkNativeButtons.
const (
	checkLayout = iota
	resizeWindow
	clickMinimize
	clickClose
	isMinimized
	deminiaturize
	isRestored
)

const nsWindowStyleMaskFullSizeContentView = 1 << 15

func className(obj appkit.ID) string {
	return appkit.GoString(appkit.ID(appkit.Call(appkit.Sym("NSStringFromClass"), uintptr(appkit.Send(obj, "class")))))
}

func checkNativeButtons(action int, report bool, height, left, offsetY, spacing float64) bool {
	ok := false
	logf := func(format string, args ...any) {
		if report {
			fmt.Fprintf(os.Stderr, format, args...)
		}
	}
	appkit.MainSync(func() {
		windows := appkit.Send(appkit.App(), "windows")
		for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
			window := appkit.Send(windows, "objectAtIndex:", i)
			if !appkit.Equal(appkit.Send(window, "title"), "Native buttons acceptance") {
				continue
			}
			switch action {
			case isMinimized:
				ok = appkit.SendBool(window, "isMiniaturized")
				return
			case deminiaturize:
				appkit.Send(window, "deminiaturize:", 0)
				ok = true
				return
			case isRestored:
				ok = !appkit.SendBool(window, "isMiniaturized")
				return
			}
			ok = appkit.SendBool(window, "titlebarAppearsTransparent") && appkit.Send(window, "titleVisibility") == 1 &&
				uintptr(appkit.Send(window, "styleMask"))&nsWindowStyleMaskFullSizeContentView != 0
			content := appkit.Send(window, "contentView")
			flipped := appkit.SendBool(content, "isFlipped")
			contentHeight := appkit.MsgRect(content, appkit.Sel("bounds")).Size.Height
			frameView := appkit.Send(content, "superview")
			convertFrom, convertTo := appkit.Sel("convertPoint:fromView:"), appkit.Sel("convertPoint:toView:")
			hitTest := func(p appkit.Point, target appkit.ID) appkit.ID {
				hit := appkit.MsgIDForPoint(frameView, appkit.Sel("hitTest:"), appkit.MsgPointFromView(content, convertTo, p, frameView))
				if hit != target && !appkit.SendBool(hit, "isDescendantOf:", uintptr(target)) {
					return hit
				}
				return 0
			}
			for kind := uintptr(0); kind <= 2; kind++ {
				button := appkit.Send(window, "standardWindowButton:", kind)
				if button == 0 || appkit.SendBool(button, "isHiddenOrHasHiddenAncestor") || !appkit.SendBool(button, "isEnabled") ||
					appkit.Send(button, "action") == 0 || appkit.Send(button, "window") != window {
					ok = false
				}
				b := appkit.MsgRect(button, appkit.Sel("bounds"))
				center := appkit.MsgPointFromView(content, convertFrom, appkit.Point{X: b.Origin.X + b.Size.Width/2, Y: b.Origin.Y + b.Size.Height/2}, button)
				top := center.Y
				if !flipped {
					top = contentHeight - center.Y
				}
				expectedX := left + appkit.MsgRect(button, appkit.Sel("frame")).Size.Width/2 + float64(kind)*spacing
				if math.Abs(center.X-expectedX) > .5 || math.Abs(top-height/2-offsetY) > .5 {
					ok = false
					logf("placement %d: actual %.1f,%.1f expected %.1f,%.1f\n", kind, center.X, top, expectedX, height/2+offsetY)
				}
				if hit := hitTest(center, button); hit != 0 {
					ok = false
					logf("button %d is not hittable: %s\n", kind, className(hit))
				}
				logf("button %d: %s / %s, hidden=%v, enabled=%v\n", kind, className(button), className(appkit.Send(button, "cell")),
					appkit.SendBool(button, "isHiddenOrHasHiddenAncestor"), appkit.SendBool(button, "isEnabled"))
			}
			tab := appkit.Point{X: 230, Y: height / 2}
			if !flipped {
				tab.Y = contentHeight - tab.Y
			}
			if hit := hitTest(tab, content); hit != 0 {
				ok = false
				logf("titlebar intercepts content tabs: %s\n", className(hit))
			}
			switch action {
			case resizeWindow:
				appkit.MsgSetSize(window, appkit.Sel("setContentSize:"), appkit.Size{Width: 720, Height: 360})
			case clickMinimize:
				appkit.Send(appkit.Send(window, "standardWindowButton:", 1), "performClick:", 0)
			case clickClose:
				appkit.Send(appkit.Send(window, "standardWindowButton:", 0), "performClick:", 0)
			}
			return
		}
	})
	return ok
}
