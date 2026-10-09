//go:build darwin && !ios

package window

import (
	"image"

	"github.com/dyike/keel/internal/appicon"
	"github.com/dyike/keel/ui/internal/appkit"
)

// platformSetIcon supplies the Dock icon, preserving finished CLI artwork.
// It is set on the main queue, once the app object exists.
func platformSetIcon(art image.Image, finished bool) {
	data := encodePNG(renderIcon(art, appicon.MacOS, 1024, finished))
	appkit.MainAsync(func() {
		image := appkit.Send(appkit.Send(appkit.Class("NSImage"), "alloc"), "initWithData:", uintptr(appkit.Data(data)))
		if image != 0 {
			appkit.Send(appkit.Send(appkit.Class("NSApplication"), "sharedApplication"), "setApplicationIconImage:", uintptr(image))
			appkit.Release(image)
		}
	})
}

func iconWindowEvent(w *Window, e any) {}
