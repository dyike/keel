// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android) || freebsd || openbsd

package app

import "unsafe"

// C string helpers for the X11 and Wayland code. Keel patch.

// cString returns s NUL-terminated in Go memory.
func cString(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// goStringN copies n bytes at p, which points outside the Go heap.
func goStringN(p unsafe.Pointer, n int) string {
	if p == nil || n == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(p), n))
}

// goString copies the NUL-terminated string at p.
func goString(p unsafe.Pointer) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}
	return goStringN(p, n)
}
