// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin

package app

import (
	"runtime"
	"structs"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The shared Darwin runtime (macOS and iOS) without cgo: Objective-C runtime
// messaging, autorelease pools, main-thread dispatch and NSString conversion.
// Keel patch.
//
// Native callbacks are created once (purego never frees them) and route by
// token or handle. Integer and pointer messages go through objc_msgSend by
// address; float and struct messages use wrappers registered once.

type nsPoint struct {
	_    structs.HostLayout
	X, Y float64
}

type nsRect struct {
	_            structs.HostLayout
	Origin, Size nsPoint
}

var rt struct {
	msg, msgStret, msgSuper, msgSuperStret uintptr
	poolPush, poolPop                      uintptr
	retain, release                        uintptr
	getAssoc, setAssoc                     uintptr
	objectGetClass, classAddMethod         uintptr
	mainThread                             uintptr
	dispatchAsync, mainQueue               uintptr
}

// Shared typed messages, registered once: float and struct arguments and
// results need the C calling convention that RegisterFunc implements.
var (
	msgF64      func(id, sel uintptr) float64
	msgSetF64   func(id, sel uintptr, v float64)
	msgSetSize  func(id, sel uintptr, s nsPoint)
	msgRect     func(id, sel uintptr) nsRect
	msgSetRect  func(id, sel uintptr, r nsRect)
	msgInitRect func(id, sel uintptr, r nsRect) uintptr
)

func init() {
	if err := loadRuntime(); err != nil {
		panic(err)
	}
	if err := loadPlatformRuntime(); err != nil {
		panic(err)
	}
}

func loadRuntime() error {
	for _, path := range []string{
		"/usr/lib/libobjc.A.dylib",
		"/usr/lib/libSystem.B.dylib",
		"/System/Library/Frameworks/Foundation.framework/Foundation",
		"/System/Library/Frameworks/QuartzCore.framework/QuartzCore",
		"/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics",
	} {
		if _, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
			return err
		}
	}
	for _, s := range []struct {
		name string
		addr *uintptr
	}{
		{"objc_msgSend", &rt.msg},
		{"objc_msgSendSuper2", &rt.msgSuper},
		{"objc_autoreleasePoolPush", &rt.poolPush},
		{"objc_autoreleasePoolPop", &rt.poolPop},
		{"objc_retain", &rt.retain},
		{"objc_release", &rt.release},
		{"objc_getAssociatedObject", &rt.getAssoc},
		{"objc_setAssociatedObject", &rt.setAssoc},
		{"object_getClass", &rt.objectGetClass},
		{"class_addMethod", &rt.classAddMethod},
		{"pthread_main_np", &rt.mainThread},
		{"dispatch_async_f", &rt.dispatchAsync},
		{"_dispatch_main_q", &rt.mainQueue},
	} {
		addr, err := purego.Dlsym(purego.RTLD_DEFAULT, s.name)
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
	for _, f := range []any{&msgF64, &msgSetF64, &msgSetSize, &msgSetRect, &msgInitRect} {
		purego.RegisterFunc(f, rt.msg)
	}
	purego.RegisterFunc(&msgRect, rt.msgStret)

	classNSString = class("NSString")
	classNSAttributedStr = class("NSAttributedString")
	classNSArray = class("NSArray")
	classNSNotificationCenter = class("NSNotificationCenter")
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

// registerClass registers a new Objective-C subclass with the runtime.
func registerClass(name, super string, protocols []*objc.Protocol, methods []objc.MethodDef) (uintptr, error) {
	c, err := objc.RegisterClass(name, objc.Class(class(super)), protocols, nil, methods)
	return uintptr(c), err
}

// addClassMethod adds a class method (+method) to cls's metaclass.
func addClassMethod(cls, sel uintptr, fn any, types string) {
	meta, _, _ := purego.SyscallN(rt.objectGetClass, cls)
	imp := objc.NewIMP(fn)
	cTypes := append([]byte(types), 0)
	purego.SyscallN(rt.classAddMethod, meta, sel, uintptr(imp), uintptr(unsafe.Pointer(&cTypes[0])))
	runtime.KeepAlive(cTypes)
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
	selLength               = sel("length")
	selGetCharacters        = sel("getCharacters:range:")
	selAlloc                = sel("alloc")
	selInit                 = sel("init")
	selInitWithChars        = sel("initWithCharacters:length:")
	selInitWithFrame        = sel("initWithFrame:")
	selIsKindOfClass        = sel("isKindOfClass:")
	selString               = sel("string")
	selRespondsToSel        = sel("respondsToSelector:")
	selBounds               = sel("bounds")
	selFrame                = sel("frame")
	selSetFrame             = sel("setFrame:")
	selLayer                = sel("layer")
	selWindow               = sel("window")
	selMainScreen           = sel("mainScreen")
	selScreen               = sel("screen")
	selContentsScale        = sel("contentsScale")
	selSetContentsScale     = sel("setContentsScale:")
	selDefaultCenter        = sel("defaultCenter")
	selAddObserver          = sel("addObserver:selector:name:object:")
	selBecomeFirstResponder = sel("becomeFirstResponder")
	selResignFirstResponder = sel("resignFirstResponder")
	selDealloc              = sel("dealloc")
	selObjectForKey         = sel("objectForKey:")
	selCount                = sel("count")
	selObjectAtIndex        = sel("objectAtIndex:")
	selAbsoluteString       = sel("absoluteString")
	selTimestamp            = sel("timestamp")
	selGeneralPasteboard    = sel("generalPasteboard")
)

// Shared Foundation classes, resolved by loadRuntime once Foundation is loaded.
var (
	classNSString             uintptr
	classNSAttributedStr      uintptr
	classNSArray              uintptr
	classNSNotificationCenter uintptr
)

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
