//go:build linux && !android && nowayland

package window

import "unsafe"

func waylandActivate(display, surface unsafe.Pointer, token string) int { return 1 }
