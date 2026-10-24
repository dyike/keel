// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (!cgo || purego)

// Package cocoainit initializes support for multithreaded
// programs in Cocoa.
package cocoainit

import (
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Keel patch: the purego form of cocoa_darwin.go, see
// docs/desktop-cgo-free.md. Starting any NSThread puts Cocoa in
// multithreaded mode; the running thread retains itself, so the reference
// from new is released, as ARC does in the cgo version.
func init() {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
		panic("cocoainit: " + err.Error())
	}
	thread := objc.ID(objc.GetClass("NSThread")).Send(objc.RegisterName("new"))
	thread.Send(objc.RegisterName("start"))
	thread.Send(objc.RegisterName("release"))
}
