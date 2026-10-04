//go:build linux && !android && !nowayland

package window

/*
#cgo pkg-config: wayland-client
#include <stdlib.h>
#include "activation_wayland.h"
*/
import "C"

import "unsafe"

// waylandActivate returns 0 on success, 1 without xdg-activation, -1 on error.
func waylandActivate(display, surface unsafe.Pointer, token string) int {
	cs := C.CString(token)
	defer C.free(unsafe.Pointer(cs))
	return int(C.keel_wayland_activate(display, surface, cs))
}
