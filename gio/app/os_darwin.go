// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && (ios || (cgo && !purego))

package app

/*
#include <Foundation/Foundation.h>

__attribute__ ((visibility ("hidden"))) void gio_runOnMain(uintptr_t h);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createDisplayLink(void);
__attribute__ ((visibility ("hidden"))) void gio_releaseDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) int gio_startDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) int gio_stopDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) void gio_setDisplayLinkDisplay(CFTypeRef dl, uint64_t did);
__attribute__ ((visibility ("hidden"))) void gio_hideCursor();
__attribute__ ((visibility ("hidden"))) void gio_showCursor();
__attribute__ ((visibility ("hidden"))) void gio_setCursor(NSUInteger curID);

static bool isMainThread() {
	return [NSThread isMainThread];
}

static NSUInteger nsstringLength(CFTypeRef cstr) {
	NSString *str = (__bridge NSString *)cstr;
	return [str length];
}

static void nsstringGetCharacters(CFTypeRef cstr, unichar *chars, NSUInteger loc, NSUInteger length) {
	NSString *str = (__bridge NSString *)cstr;
	[str getCharacters:chars range:NSMakeRange(loc, length)];
}

static CFTypeRef newNSString(unichar *chars, NSUInteger length) {
	@autoreleasepool {
		NSString *s = [NSString string];
		if (length > 0) {
			s = [NSString stringWithCharacters:chars length:length];
		}
		return CFBridgingRetain(s);
	}
}
*/
import "C"

import (
	"runtime/cgo"
	"unicode/utf16"
	"unsafe"
)

func isMainThread() bool {
	return bool(C.isMainThread())
}

// runOnMain runs the function on the main thread.
func runOnMain(f func()) {
	if isMainThread() {
		f()
		return
	}
	C.gio_runOnMain(C.uintptr_t(cgo.NewHandle(f)))
}

//export gio_runFunc
func gio_runFunc(h C.uintptr_t) {
	handle := cgo.Handle(h)
	defer handle.Delete()
	f := handle.Value().(func())
	f()
}

// nsstringToString converts a NSString to a Go string.
func nsstringToString(str C.CFTypeRef) string {
	if str == 0 {
		return ""
	}
	n := C.nsstringLength(str)
	chars := make([]uint16, n)
	C.nsstringGetCharacters(str, (*C.unichar)(unsafe.Pointer(unsafe.SliceData(chars))), 0, n)
	utf8 := utf16.Decode(chars)
	return string(utf8)
}

// stringToNSString converts a Go string to a retained NSString.
func stringToNSString(str string) C.CFTypeRef {
	u16 := utf16.Encode([]rune(str))
	chars := (*C.unichar)(unsafe.Pointer(unsafe.SliceData(u16)))
	return C.newNSString(chars, C.NSUInteger(len(u16)))
}

func createDisplayLink() uintptr    { return uintptr(C.gio_createDisplayLink()) }
func releaseDisplayLink(dl uintptr) { C.gio_releaseDisplayLink(C.CFTypeRef(dl)) }
func startDisplayLink(dl uintptr)   { C.gio_startDisplayLink(C.CFTypeRef(dl)) }
func stopDisplayLink(dl uintptr)    { C.gio_stopDisplayLink(C.CFTypeRef(dl)) }
func setDisplayLinkDisplay(dl uintptr, did uint64) {
	C.gio_setDisplayLinkDisplay(C.CFTypeRef(dl), C.uint64_t(did))
}
func hideCursor()       { C.gio_hideCursor() }
func showCursor()       { C.gio_showCursor() }
func setCursor(id uint) { C.gio_setCursor(C.NSUInteger(id)) }

//export gio_onFrameCallback
func gio_onFrameCallback(ref C.CFTypeRef) {
	onFrameCallback(uintptr(ref))
}
