// SPDX-License-Identifier: Unlicense OR MIT

//go:build (freebsd || openbsd) && !nox11

package app

import "unsafe"

const x11FramelessSupported = false

func x11UngrabPointer(unsafe.Pointer, uintptr) {}
