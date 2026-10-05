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

// platformSetIcon gives the Dock Apple's plate: 824px with continuous
// corners and the template shadow on a 1024px canvas.
func platformSetIcon(art image.Image) {
	data := encodePNG(appicon.MacOS.Render(art, 1024, true))
	C.keel_set_app_icon(unsafe.Pointer(&data[0]), C.size_t(len(data)))
}

func iconWindowEvent(w *Window, e any) {}
