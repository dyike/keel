// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && (!cgo || purego)

package app

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// Calls into C libraries for the X11 and Wayland bindings without cgo.
// Keel patch.

// openLib opens the first of names that loads.
func openLib(names ...string) (uintptr, error) {
	var err error
	for _, name := range names {
		var h uintptr
		if h, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
			return h, nil
		}
	}
	return 0, err
}

// xcall calls a C function; pointers to Go memory stay valid and in place
// for the call.
//
//go:uintptrescapes
func xcall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// ptr converts an opaque C pointer, which the garbage collector ignores.
func ptr(p unsafe.Pointer) uintptr { return uintptr(p) }

// cptr returns a pointer result, which points outside the Go heap.
func cptr(r uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&r)) }

// i32 widens a C int result; the register's upper half is undefined.
func i32(r uintptr) int32 { return int32(r) }

// iarg passes a C int argument.
func iarg(v int32) uintptr { return uintptr(v) }

func bytesArg(b []byte) uintptr { return uintptr(unsafe.Pointer(unsafe.SliceData(b))) }
