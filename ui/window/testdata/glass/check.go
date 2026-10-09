//go:build darwin && !ios && !nometal

package main

import (
	"math"

	"github.com/dyike/keel/ui/internal/appkit"
)

func subviews(view appkit.ID) []appkit.ID {
	list := appkit.Send(view, "subviews")
	out := make([]appkit.ID, int(appkit.Send(list, "count")))
	for i := range out {
		out[i] = appkit.Send(list, "objectAtIndex:", uintptr(i))
	}
	return out
}

func isKind(obj appkit.ID, class string) bool {
	c := appkit.Class(class)
	return obj != 0 && c != 0 && appkit.SendBool(obj, "isKindOfClass:", uintptr(c))
}

func checkGlass(liquid bool, width, height float64) bool {
	ok := false
	appkit.MainSync(func() {
		windows := appkit.Send(appkit.App(), "windows")
		for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
			window := appkit.Send(windows, "objectAtIndex:", i)
			if !appkit.Equal(appkit.Send(window, "title"), "glass acceptance") || !appkit.SendBool(window, "isVisible") || appkit.SendBool(window, "isMiniaturized") {
				continue
			}
			gio := appkit.Send(window, "contentView")
			if appkit.GoString(appkit.ID(appkit.Call(appkit.Sym("NSStringFromClass"), uintptr(appkit.Send(gio, "class"))))) != "GioView" {
				continue
			}
			alpha := appkit.MsgFloat(appkit.Send(window, "backgroundColor"), appkit.Sel("alphaComponent"))
			if appkit.SendBool(window, "isOpaque") || appkit.SendBool(appkit.Send(gio, "layer"), "isOpaque") || alpha != 0 {
				continue
			}
			bounds := appkit.MsgRect(gio, appkit.Sel("bounds")).Size
			if math.Abs(bounds.Width-width) > 1 || math.Abs(bounds.Height-height) > 1 {
				continue
			}
			var base appkit.ID
			for _, child := range subviews(gio) {
				if isKind(child, "NSVisualEffectView") {
					base = child
				}
			}
			if base == 0 || appkit.Send(base, "blendingMode") != 0 { // NSVisualEffectBlendingModeBehindWindow
				continue
			}
			var host, glass appkit.ID
			for _, child := range subviews(base) {
				if isKind(child, "NSGlassEffectView") {
					glass = child
				} else if isKind(appkit.Send(child, "layer"), "CAMetalLayer") {
					host = child
				}
			}
			if liquid {
				if glass == 0 {
					continue
				}
				host = appkit.Send(glass, "valueForKey:", uintptr(appkit.String("contentView")))
				if appkit.MsgRect(glass, appkit.Sel("bounds")).Size != bounds {
					continue
				}
			} else if glass != 0 {
				continue
			}
			if host == 0 || !isKind(appkit.Send(host, "layer"), "CAMetalLayer") {
				continue
			}
			layer := appkit.Send(host, "layer")
			if appkit.SendBool(layer, "isOpaque") || appkit.Send(layer, "device") == 0 || appkit.MsgRect(host, appkit.Sel("bounds")).Size != bounds {
				continue
			}
			scale := appkit.MsgFloat(window, appkit.Sel("backingScaleFactor"))
			drawable := appkit.MsgSize(layer, appkit.Sel("drawableSize"))
			if math.Abs(drawable.Width-width*scale) > 2 || math.Abs(drawable.Height-height*scale) > 2 {
				continue
			}
			// The native effect must not consume Gio input events.
			point := appkit.MsgPointFromView(gio, appkit.Sel("convertPoint:toView:"), appkit.Point{X: 30, Y: 80}, appkit.Send(gio, "superview"))
			if appkit.MsgIDForPoint(gio, appkit.Sel("hitTest:"), point) != gio {
				continue
			}
			ok = true
		}
	})
	return ok
}
