//go:build darwin && !ios

package sys

// Objective-C, Core Foundation and libdispatch through purego, without cgo.
// Frameworks load on first use, so programs that never call native code do
// not pay for them. Callbacks into Go are created once per signature.

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type id = objc.ID

type cgPoint struct{ X, Y float64 }
type cgRect struct {
	Origin cgPoint
	Size   struct{ Width, Height float64 }
}

var (
	loadOnce sync.Once

	msgSendAddr uintptr
	poolPushFn  uintptr
	poolPopFn   uintptr
	mainNPFn    uintptr
	cfReleaseFn uintptr
	mainQueue   uintptr
	asyncFFn    uintptr
	syncFFn     uintptr
	dispatchFn  uintptr // one callback for every dispatched Go function

	cgDisplayBounds       func(display uint32) cgRect
	cgEventGetLocation    func(event uintptr) cgPoint
	cgEventCreateMouse    func(source uintptr, typ uint32, at cgPoint, button uint32) uintptr
	cgEventCreateKeyboard func(source uintptr, key uint16, down bool) uintptr
)

// load opens the frameworks once. A framework that is missing (an older
// system) leaves its symbols unresolved; callers check with sym.
func load() {
	loadOnce.Do(func() {
		for _, path := range []string{
			"/usr/lib/libobjc.A.dylib",
			"/System/Library/Frameworks/Foundation.framework/Foundation",
			"/System/Library/Frameworks/AppKit.framework/AppKit",
			"/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics",
			"/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices",
			"/System/Library/Frameworks/Carbon.framework/Carbon",
			"/System/Library/Frameworks/ScreenCaptureKit.framework/ScreenCaptureKit",
			"/System/Library/Frameworks/UserNotifications.framework/UserNotifications",
		} {
			_, _ = purego.Dlopen(path, purego.RTLD_GLOBAL|purego.RTLD_NOW)
		}
		msgSendAddr = mustSym("objc_msgSend")
		poolPushFn = mustSym("objc_autoreleasePoolPush")
		poolPopFn = mustSym("objc_autoreleasePoolPop")
		mainNPFn = mustSym("pthread_main_np")
		cfReleaseFn = mustSym("CFRelease")
		// dispatch_get_main_queue() is a macro for &_dispatch_main_q.
		mainQueue = mustSym("_dispatch_main_q")
		asyncFFn = mustSym("dispatch_async_f")
		syncFFn = mustSym("dispatch_sync_f")
		dispatchFn = purego.NewCallback(runDispatched)
		purego.RegisterFunc(&cgDisplayBounds, mustSym("CGDisplayBounds"))
		purego.RegisterFunc(&cgEventGetLocation, mustSym("CGEventGetLocation"))
		purego.RegisterFunc(&cgEventCreateMouse, mustSym("CGEventCreateMouseEvent"))
		purego.RegisterFunc(&cgEventCreateKeyboard, mustSym("CGEventCreateKeyboardEvent"))
	})
}

// sym returns the address of a C symbol in a loaded image, or 0.
func sym(name string) uintptr {
	p, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if err != nil {
		return 0
	}
	return p
}

func mustSym(name string) uintptr {
	p := sym(name)
	if p == 0 {
		panic("keel: native symbol " + name + " not found")
	}
	return p
}

// call calls a C function taking and returning integers or pointers.
//
//go:uintptrescapes
func call(name string, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(mustSym(name), args...)
	return r
}

// constant reads an exported object constant, such as NSPasteboardTypeString.
func constant(name string) id {
	p := sym(name)
	if p == 0 {
		return 0
	}
	return **(**id)(unsafe.Pointer(&p))
}

var (
	selMu    sync.RWMutex
	selCache = map[string]objc.SEL{}
)

func sel(name string) objc.SEL {
	selMu.RLock()
	s, ok := selCache[name]
	selMu.RUnlock()
	if !ok {
		s = objc.RegisterName(name)
		selMu.Lock()
		selCache[name] = s
		selMu.Unlock()
	}
	return s
}

// class returns a class, or 0 when this system does not have it.
func class(name string) id { return id(objc.GetClass(name)) }

// send sends a message whose arguments are integers or pointers.
//
//go:uintptrescapes
func send(obj id, selector string, args ...uintptr) id {
	var a [8]uintptr
	a[0], a[1] = uintptr(obj), uintptr(sel(selector))
	n := copy(a[2:], args)
	r, _, _ := purego.SyscallN(msgSendAddr, a[:n+2]...)
	return id(r)
}

//go:uintptrescapes
func sendBool(obj id, selector string, args ...uintptr) bool {
	return byte(send(obj, selector, args...)) != 0
}

func release(obj id) {
	if obj != 0 {
		send(obj, "release")
	}
}

func cfRelease(ref uintptr) {
	if ref != 0 {
		purego.SyscallN(cfReleaseFn, ref)
	}
}

// withPool runs fn in an autorelease pool on a locked thread: a pool must be
// popped on the thread that pushed it.
func withPool(fn func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool, _, _ := purego.SyscallN(poolPushFn)
	defer purego.SyscallN(poolPopFn, pool)
	fn()
}

func isMainThread() bool {
	r, _, _ := purego.SyscallN(mainNPFn)
	return int32(r) != 0
}

const utf8Encoding = 4

// nsString returns an autoreleased NSString.
func nsString(s string) id {
	b := unsafe.StringData(s)
	str := send(send(class("NSString"), "alloc"), "initWithBytes:length:encoding:",
		uintptr(unsafe.Pointer(b)), uintptr(len(s)), utf8Encoding)
	runtime.KeepAlive(s)
	return send(str, "autorelease")
}

// goString copies an NSString.
func goString(str id) string {
	if str == 0 {
		return ""
	}
	p := send(str, "UTF8String")
	if p == 0 {
		return ""
	}
	n := int(send(str, "lengthOfBytesUsingEncoding:", utf8Encoding))
	return string(unsafe.Slice(*(**byte)(unsafe.Pointer(&p)), n))
}

// goBytes copies the contents of an NSData.
func goBytes(data id) []byte {
	n := int(send(data, "length"))
	if n == 0 {
		return []byte{}
	}
	p := send(data, "bytes")
	return append([]byte(nil), unsafe.Slice(*(**byte)(unsafe.Pointer(&p)), n)...)
}

// callBlock invokes a block Apple passed in, with integer arguments.
func callBlock(block uintptr, args ...uintptr) {
	lit := *(**[3]uintptr)(unsafe.Pointer(&block))
	var a [4]uintptr
	a[0] = block
	n := copy(a[1:], args)
	purego.SyscallN(lit[2], a[:n+1]...)
}

// Functions queued for libdispatch, by token: Go pointers never reach C.
var dispatched struct {
	sync.Mutex
	next uintptr
	fns  map[uintptr]func()
}

func runDispatched(token uintptr) {
	dispatched.Lock()
	fn := dispatched.fns[token]
	delete(dispatched.fns, token)
	dispatched.Unlock()
	if fn != nil {
		withPool(fn)
	}
}

func dispatchToken(fn func()) uintptr {
	dispatched.Lock()
	defer dispatched.Unlock()
	if dispatched.fns == nil {
		dispatched.fns = map[uintptr]func(){}
	}
	dispatched.next++
	dispatched.fns[dispatched.next] = fn
	return dispatched.next
}

// mainAsync runs fn on the main queue later.
func mainAsync(fn func()) {
	load()
	purego.SyscallN(asyncFFn, mainQueue, dispatchToken(fn), dispatchFn)
}

// onMain runs fn on the main thread and waits for it.
func onMain(fn func()) {
	load()
	if isMainThread() {
		withPool(fn)
		return
	}
	purego.SyscallN(syncFFn, mainQueue, dispatchToken(fn), dispatchFn)
}
