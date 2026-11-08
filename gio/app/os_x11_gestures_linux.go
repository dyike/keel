// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nox11

package app

import (
	"github.com/ebitengine/purego"
	"sync"
	"unsafe"
)

const x11FramelessSupported = true

var x11Pointer struct {
	sync.Once
	ungrab uintptr
}

func x11UngrabPointer(display unsafe.Pointer, time uintptr) {
	x11Pointer.Do(func() {
		h, err := purego.Dlopen("libX11.so.6", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err == nil {
			x11Pointer.ungrab, _ = purego.Dlsym(h, "XUngrabPointer")
		}
	})
	if x11Pointer.ungrab != 0 {
		purego.SyscallN(x11Pointer.ungrab, uintptr(display), time)
	}
}
