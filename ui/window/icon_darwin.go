//go:build darwin && !ios && cgo

package window

/*
#cgo LDFLAGS: -framework AppKit
#include <stddef.h>
void keel_set_app_icon(const void *png, size_t len);
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/dyike/keel/internal/appicon"
)

// platformSetIcon supplies the Dock icon, preserving finished CLI artwork.
func platformSetIcon(art image.Image, finished bool) {
	data := encodePNG(renderIcon(art, appicon.MacOS, 1024, finished))
	C.keel_set_app_icon(unsafe.Pointer(&data[0]), C.size_t(len(data)))
}

func iconWindowEvent(w *Window, e any) {}
