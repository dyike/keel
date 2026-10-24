// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (!cgo || purego)

package app

import (
	"runtime"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The darwin runtime of os_darwin.go without cgo: main-thread dispatch,
// NSString conversion, the CVDisplayLink and cursors. Keel patch.
//
// Native callbacks are created once (purego never frees them) and route by
// token or handle. Integer and pointer messages go through objc_msgSend by
// address; float and struct messages use wrappers registered once.

var rt struct {
	msg, msgStret, msgSuper, msgSuperStret uintptr
	poolPush, poolPop                      uintptr
	retain, release                        uintptr
	getAssoc, setAssoc                     uintptr
	mainThread                             uintptr
	dispatchAsync, mainQueue               uintptr
	dlCreateActive, dlCreateDisplay        uintptr
	dlSetCallback, dlStart, dlStop         uintptr
	dlRelease, dlSetDisplay, mainDisplay   uintptr
}

func init() {
	if err := loadRuntime(); err != nil {
		panic(err)
	}
}

func loadRuntime() error {
	libs := map[string]uintptr{}
	for _, path := range []string{
		"/usr/lib/libobjc.A.dylib",
		"/usr/lib/libSystem.B.dylib",
		"/System/Library/Frameworks/Foundation.framework/Foundation",
		"/System/Library/Frameworks/AppKit.framework/AppKit",
		"/System/Library/Frameworks/QuartzCore.framework/QuartzCore",
		"/System/Library/Frameworks/CoreVideo.framework/CoreVideo",
		"/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics",
	} {
		h, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return err
		}
		libs[path] = h
	}
	cv := libs["/System/Library/Frameworks/CoreVideo.framework/CoreVideo"]
	cg := libs["/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics"]
	for _, s := range []struct {
		lib  uintptr
		name string
		addr *uintptr
	}{
		{purego.RTLD_DEFAULT, "objc_msgSend", &rt.msg},
		{purego.RTLD_DEFAULT, "objc_msgSendSuper2", &rt.msgSuper},
		{purego.RTLD_DEFAULT, "objc_autoreleasePoolPush", &rt.poolPush},
		{purego.RTLD_DEFAULT, "objc_autoreleasePoolPop", &rt.poolPop},
		{purego.RTLD_DEFAULT, "objc_retain", &rt.retain},
		{purego.RTLD_DEFAULT, "objc_release", &rt.release},
		{purego.RTLD_DEFAULT, "objc_getAssociatedObject", &rt.getAssoc},
		{purego.RTLD_DEFAULT, "objc_setAssociatedObject", &rt.setAssoc},
		{purego.RTLD_DEFAULT, "pthread_main_np", &rt.mainThread},
		{purego.RTLD_DEFAULT, "dispatch_async_f", &rt.dispatchAsync},
		{purego.RTLD_DEFAULT, "_dispatch_main_q", &rt.mainQueue},
		{cv, "CVDisplayLinkCreateWithActiveCGDisplays", &rt.dlCreateActive},
		{cv, "CVDisplayLinkCreateWithCGDisplay", &rt.dlCreateDisplay},
		{cv, "CVDisplayLinkSetOutputCallback", &rt.dlSetCallback},
		{cv, "CVDisplayLinkStart", &rt.dlStart},
		{cv, "CVDisplayLinkStop", &rt.dlStop},
		{cv, "CVDisplayLinkRelease", &rt.dlRelease},
		{cv, "CVDisplayLinkSetCurrentCGDisplay", &rt.dlSetDisplay},
		{cg, "CGMainDisplayID", &rt.mainDisplay},
	} {
		addr, err := purego.Dlsym(s.lib, s.name)
		if err != nil {
			return err
		}
		*s.addr = addr
	}
	// Structs larger than 16 bytes come back through a hidden pointer on
	// amd64, which objc_msgSend_stret expects; arm64 has no stret variant.
	rt.msgStret, rt.msgSuperStret = rt.msg, rt.msgSuper
	if runtime.GOARCH == "amd64" {
		var err error
		if rt.msgStret, err = purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSend_stret"); err != nil {
			return err
		}
		if rt.msgSuperStret, err = purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSendSuper2_stret"); err != nil {
			return err
		}
	}
	classNSString = class("NSString")
	classNSAttributedStr = class("NSAttributedString")
	classNSCursor = class("NSCursor")
	return nil
}

// sel returns a selector; resolve selectors once, into package variables.
func sel(name string) uintptr { return uintptr(objc.RegisterName(name)) }

// class returns a class, or panics: the frameworks are loaded at init.
func class(name string) uintptr {
	c := uintptr(objc.GetClass(name))
	if c == 0 {
		panic("app: Objective-C class " + name + " not found")
	}
	return c
}

// send sends an integer or pointer message. Pointers to Go memory stay
// valid for the call only.
func send(id, sel uintptr, args ...uintptr) uintptr {
	a := make([]uintptr, 0, 2+len(args))
	a = append(a, id, sel)
	a = append(a, args...)
	r, _, _ := purego.SyscallN(rt.msg, a...)
	return r
}

// objcSuper is struct objc_super for objc_msgSendSuper2: the search starts
// above class, which is the implementing class, never the object's
// dynamic class, which can be a runtime subclass of it.
type objcSuper struct {
	receiver, class uintptr
}

func sendSuper(id, cls, sel uintptr, args ...uintptr) uintptr {
	sup := &objcSuper{id, cls}
	a := make([]uintptr, 0, 2+len(args))
	a = append(a, uintptr(unsafe.Pointer(sup)), sel)
	a = append(a, args...)
	r, _, _ := purego.SyscallN(rt.msgSuper, a...)
	runtime.KeepAlive(sup)
	return r
}

func retainObject(obj uintptr) uintptr {
	if obj != 0 {
		purego.SyscallN(rt.retain, obj)
	}
	return obj
}

func releaseObject(obj uintptr) {
	if obj != 0 {
		purego.SyscallN(rt.release, obj)
	}
}

// withPool runs f inside an autorelease pool on a locked thread.
func withPool(f func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	p, _, _ := purego.SyscallN(rt.poolPush)
	defer purego.SyscallN(rt.poolPop, p)
	f()
}

func associated(obj, key uintptr) uintptr {
	r, _, _ := purego.SyscallN(rt.getAssoc, obj, key)
	return r
}

const (
	objcAssociationAssign          = 0
	objcAssociationRetainNonatomic = 1
)

func setAssociated(obj, key, value, policy uintptr) {
	purego.SyscallN(rt.setAssoc, obj, key, value, policy)
}

func boolArg(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

var (
	selLength        = sel("length")
	selGetCharacters = sel("getCharacters:range:")
	selAlloc         = sel("alloc")
	selInit          = sel("init")
	selInitWithChars = sel("initWithCharacters:length:")
	selIsKindOfClass = sel("isKindOfClass:")
	selString        = sel("string")
	selRespondsToSel = sel("respondsToSelector:")
	selSet           = sel("set")
	selHide          = sel("hide")
	selUnhide        = sel("unhide")
)

// Classes, resolved by loadRuntime once the frameworks are loaded.
var classNSString, classNSAttributedStr, classNSCursor uintptr

func isMainThread() bool {
	r, _, _ := purego.SyscallN(rt.mainThread)
	return r != 0
}

// mainFuncs holds the functions posted to the main queue, by token.
var mainFuncs struct {
	sync.Mutex
	next  uintptr
	funcs map[uintptr]func()
}

// mainCallback runs a posted function; one callback serves every post.
var mainCallback = purego.NewCallback(func(token uintptr) {
	mainFuncs.Lock()
	f := mainFuncs.funcs[token]
	delete(mainFuncs.funcs, token)
	mainFuncs.Unlock()
	f()
})

// runOnMain runs the function on the main thread.
func runOnMain(f func()) {
	if isMainThread() {
		f()
		return
	}
	mainFuncs.Lock()
	if mainFuncs.funcs == nil {
		mainFuncs.funcs = make(map[uintptr]func())
	}
	mainFuncs.next++
	token := mainFuncs.next
	mainFuncs.funcs[token] = f
	mainFuncs.Unlock()
	purego.SyscallN(rt.dispatchAsync, rt.mainQueue, token, mainCallback)
}

// nsstringToString converts a NSString to a Go string.
func nsstringToString(str uintptr) string {
	if str == 0 {
		return ""
	}
	n := send(str, selLength)
	if n == 0 {
		return ""
	}
	chars := make([]uint16, n)
	// NSRange is two integers and travels in two registers.
	send(str, selGetCharacters, uintptr(unsafe.Pointer(unsafe.SliceData(chars))), 0, n)
	return string(utf16.Decode(chars))
}

// stringToNSString converts a Go string to a retained NSString.
func stringToNSString(str string) uintptr {
	u16 := utf16.Encode([]rune(str))
	s := send(classNSString, selAlloc)
	if len(u16) == 0 {
		return send(s, selInit)
	}
	return send(s, selInitWithChars, uintptr(unsafe.Pointer(unsafe.SliceData(u16))), uintptr(len(u16)))
}

// plainString returns the NSString of a string or attributed string.
func plainString(obj uintptr) uintptr {
	if obj != 0 && send(obj, selIsKindOfClass, classNSAttributedStr) != 0 {
		return send(obj, selString)
	}
	return obj
}

// displayLinkCallback is the CVDisplayLink output callback for every link.
var displayLinkCallback = purego.NewCallback(func(dl, inNow, inOutputTime uintptr, flagsIn uint64, flagsOut, ctx uintptr) int32 {
	onFrameCallback(dl)
	return 0 // kCVReturnSuccess
})

func createDisplayLink() uintptr {
	// Keel patch (as in os_macos.m): with every display asleep there are no
	// active displays and creation fails. Fall back to the main display; the
	// window moves the link to its own display later.
	var dl uintptr
	if r, _, _ := purego.SyscallN(rt.dlCreateActive, uintptr(unsafe.Pointer(&dl))); int32(r) != 0 {
		dl = 0
		main, _, _ := purego.SyscallN(rt.mainDisplay)
		if r, _, _ := purego.SyscallN(rt.dlCreateDisplay, uintptr(uint32(main)), uintptr(unsafe.Pointer(&dl))); int32(r) != 0 {
			return 0
		}
	}
	purego.SyscallN(rt.dlSetCallback, dl, displayLinkCallback, 0)
	return dl
}

func releaseDisplayLink(dl uintptr) { purego.SyscallN(rt.dlRelease, dl) }
func startDisplayLink(dl uintptr)   { purego.SyscallN(rt.dlStart, dl) }
func stopDisplayLink(dl uintptr)    { purego.SyscallN(rt.dlStop, dl) }

func setDisplayLinkDisplay(dl uintptr, did uint64) {
	purego.SyscallN(rt.dlSetDisplay, dl, uintptr(uint32(did)))
}

func hideCursor() { withPool(func() { send(classNSCursor, selHide) }) }
func showCursor() { withPool(func() { send(classNSCursor, selUnhide) }) }

// cursorSelectors are the NSCursor class methods of os_macos.m's
// gio_setCursor, by cursor ID. A private cursor has a public fallback.
var cursorSelectors = [...]struct{ name, fallback string }{
	0:  {"arrowCursor", ""},
	2:  {"IBeamCursor", ""},
	3:  {"IBeamCursorForVerticalLayout", ""},
	4:  {"pointingHandCursor", ""},
	5:  {"crosshairCursor", ""},
	6:  {"arrowCursor", ""}, // _moveCursor fails on Monterey.
	7:  {"resizeLeftRightCursor", ""},
	8:  {"resizeUpDownCursor", ""},
	9:  {"openHandCursor", ""},
	10: {"closedHandCursor", ""},
	11: {"operationNotAllowedCursor", ""},
	12: {"busyButClickableCursor", "arrowCursor"},
	13: {"busyButClickableCursor", "arrowCursor"},
	14: {"_windowResizeNorthWestCursor", "resizeUpDownCursor"},
	15: {"_windowResizeNorthEastCursor", "resizeUpDownCursor"},
	16: {"_windowResizeSouthWestCursor", "resizeUpDownCursor"},
	17: {"_windowResizeSouthEastCursor", "resizeUpDownCursor"},
	18: {"resizeUpDownCursor", ""},
	19: {"resizeLeftRightCursor", ""},
	20: {"resizeLeftCursor", ""},
	21: {"resizeRightCursor", ""},
	22: {"resizeUpCursor", ""},
	23: {"resizeDownCursor", ""},
	24: {"_windowResizeNorthEastSouthWestCursor", "resizeUpDownCursor"},
	25: {"_windowResizeNorthWestSouthEastCursor", "resizeUpDownCursor"},
}

func setCursor(id uint) {
	withPool(func() {
		c := struct{ name, fallback string }{"arrowCursor", ""}
		if id < uint(len(cursorSelectors)) && cursorSelectors[id].name != "" {
			c = cursorSelectors[id]
		}
		if c.fallback != "" {
			// Some cursors are not public: use one only when NSCursor has it.
			s := sel(c.name)
			if send(classNSCursor, selRespondsToSel, s) != 0 {
				if cur := send(classNSCursor, s); cur != 0 && send(cur, selIsKindOfClass, classNSCursor) != 0 {
					send(cur, selSet)
					return
				}
			}
			c.name = c.fallback
		}
		send(send(classNSCursor, sel(c.name)), selSet)
	})
}
