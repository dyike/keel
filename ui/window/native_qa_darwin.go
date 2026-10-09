//go:build darwin && !ios && keelnativeqa

package window

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/appkit"
)

// Main thread.
func qaFindMenuItem(menu appkit.ID, name string) appkit.ID {
	if menu == 0 {
		return 0
	}
	items := appkit.Send(menu, "itemArray")
	for i := uintptr(0); i < uintptr(appkit.Send(items, "count")); i++ {
		item := appkit.Send(items, "objectAtIndex:", i)
		if appkit.Equal(appkit.Send(item, "representedObject"), name) {
			return item
		}
		if found := qaFindMenuItem(appkit.Send(item, "submenu"), name); found != 0 {
			return found
		}
	}
	return 0
}

// NativeQAActivateMenu exercises an installed native menu action in development tests.
func NativeQAActivateMenu(id string) {
	appkit.MainAsync(func() {
		app := appkit.App()
		item := qaFindMenuItem(appkit.Send(app, "mainMenu"), id)
		if item != 0 && appkit.SendBool(item, "isEnabled") {
			appkit.Send(app, "sendAction:to:from:", uintptr(appkit.Send(item, "action")), uintptr(appkit.Send(item, "target")), uintptr(item))
		}
	})
}

// NativeQACompose exercises the native input method bridge in development tests.
func NativeQACompose(commit bool) {
	appkit.MainAsync(func() {
		// The fixture may run while another app owns keyboard focus.
		// Deliver input to the application's own window rather than messaging a nil responder.
		app := appkit.App()
		window := appkit.Send(app, "keyWindow")
		if window == 0 {
			windows := appkit.Send(app, "windows")
			for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
				candidate := appkit.Send(windows, "objectAtIndex:", i)
				if appkit.SendBool(candidate, "isVisible") && appkit.SendBool(candidate, "canBecomeKeyWindow") {
					appkit.Send(candidate, "makeKeyAndOrderFront:", 0)
					window = candidate
					break
				}
			}
		}
		view := appkit.Send(window, "firstResponder")
		// NSRange travels as two integers.
		if commit {
			appkit.Send(view, "insertText:replacementRange:", uintptr(appkit.String("你好中文")), appkit.NotFound, 0)
		} else {
			appkit.Send(view, "setMarkedText:selectedRange:replacementRange:", uintptr(appkit.String("imepreedit")), 10, 0, appkit.NotFound, 0)
		}
		go core.Update(func() {})
	})
}

const nsBitmapImageFileTypePNG = 4

// srgbPNG converts an image's TIFF representation to sRGB PNG data.
func srgbPNG(tiff appkit.ID) appkit.ID {
	rep := appkit.Send(appkit.Class("NSBitmapImageRep"), "imageRepWithData:", uintptr(tiff))
	rep = appkit.Send(rep, "bitmapImageRepByConvertingToColorSpace:renderingIntent:", uintptr(appkit.Send(appkit.Class("NSColorSpace"), "sRGBColorSpace")), 0)
	return appkit.Send(rep, "representationUsingType:properties:", nsBitmapImageFileTypePNG, uintptr(appkit.Send(appkit.Class("NSDictionary"), "dictionary")))
}

// NativeQACaptureIcon saves the actual and configured AppKit icons for comparison.
// Call from a background goroutine, outside core.Update.
func NativeQACaptureIcon(path, source, reference string) error {
	ok := false
	appkit.MainSync(func() {
		icon := srgbPNG(appkit.Send(appkit.Send(appkit.App(), "applicationIconImage"), "TIFFRepresentation"))
		ok = appkit.SendBool(icon, "writeToFile:atomically:", uintptr(appkit.String(path)), 1)
		image := appkit.Send(appkit.Send(appkit.Class("NSImage"), "alloc"), "initWithContentsOfFile:", uintptr(appkit.String(source)))
		expected := srgbPNG(appkit.Send(image, "TIFFRepresentation"))
		ok = ok && appkit.SendBool(expected, "writeToFile:atomically:", uintptr(appkit.String(reference)), 1)
		appkit.Release(image)
	})
	if !ok {
		return fmt.Errorf("window: native icon capture failed")
	}
	return nil
}
