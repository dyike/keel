// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android && cgo && giocgo && !purego) || freebsd) && !nowayland

package app

import "unsafe"

// The legacy binding retains integer scaling without adding cgo protocols.
func bindFractionalGlobal(*wlDisplay, unsafe.Pointer, uint32, string) {}
func (*window) initFractionalScale()                                  {}
func (*window) updateViewport()                                       {}
func (*window) destroyFractionalScale()                               {}
func (*wlDisplay) destroyFractionalScale()                            {}

func waylandShellVersion(uint32) uint32 { return 1 }
