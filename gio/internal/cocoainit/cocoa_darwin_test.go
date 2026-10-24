// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios

package cocoainit

import (
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Importing the package leaves Cocoa in multithreaded mode, with either
// binding.
func TestCocoaIsMultiThreaded(t *testing.T) {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
		t.Fatal(err)
	}
	var isMulti func(objc.ID, objc.SEL) bool
	purego.RegisterLibFunc(&isMulti, purego.RTLD_DEFAULT, "objc_msgSend")
	if !isMulti(objc.ID(objc.GetClass("NSThread")), objc.RegisterName("isMultiThreaded")) {
		t.Fatal("NSThread reports a single-threaded process")
	}
}
