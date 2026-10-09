//go:build darwin && !ios

package appkit

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type ID = objc.ID
type SEL = objc.SEL

type Point struct{ X, Y float64 }
type Size struct{ Width, Height float64 }
type Rect struct {
	Origin Point
	Size   Size
}

func (r Rect) MaxY() float64 { return r.Origin.Y + r.Size.Height }

// Contains reports whether p lies in r, as NSPointInRect does.
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Origin.X && p.Y >= r.Origin.Y && p.X < r.Origin.X+r.Size.Width && p.Y < r.MaxY()
}

// NSRange's NotFound location.
const NotFound = uintptr(1<<63 - 1)

var (
	loadOnce sync.Once

	msgSendAddr, poolPushFn, poolPopFn, mainNPFn uintptr
	mainQueue, asyncFFn, syncFFn, dispatchFn     uintptr

	// Typed objc_msgSend variants for float and struct arguments. They use
	// reflection: keep them off per-event paths where possible.
	MsgRect          func(obj ID, sel SEL) Rect
	MsgSetRect       func(obj ID, sel SEL, r Rect)
	MsgSetPoint      func(obj ID, sel SEL, p Point)
	MsgSetSize       func(obj ID, sel SEL, s Size)
	MsgFloat         func(obj ID, sel SEL) float64
	MsgSetFloat      func(obj ID, sel SEL, v float64)
	MsgPoint         func(obj ID, sel SEL) Point
	MsgSize          func(obj ID, sel SEL) Size
	MsgIDForPoint    func(obj ID, sel SEL, p Point) ID
	MsgInitRect      func(obj ID, sel SEL, r Rect) ID
	MsgPointFromView func(obj ID, sel SEL, p Point, view ID) Point
	MsgKeyEvent      func(cls ID, sel SEL, typ uint, loc Point, flags uint, ts float64, wn int, ctx ID, chars, charsIgnoring ID, repeat bool, keyCode uint16) ID
)

func load() {
	loadOnce.Do(func() {
		for _, path := range []string{
			"/usr/lib/libobjc.A.dylib",
			"/System/Library/Frameworks/Foundation.framework/Foundation",
			"/System/Library/Frameworks/AppKit.framework/AppKit",
			"/System/Library/Frameworks/QuartzCore.framework/QuartzCore",
			"/System/Library/Frameworks/Metal.framework/Metal",
		} {
			_, _ = purego.Dlopen(path, purego.RTLD_GLOBAL|purego.RTLD_NOW)
		}
		msgSendAddr = mustSym("objc_msgSend")
		stret := msgSendAddr
		if runtime.GOARCH == "amd64" {
			// Structs larger than 16 bytes come back through memory.
			stret = mustSym("objc_msgSend_stret")
		}
		poolPushFn = mustSym("objc_autoreleasePoolPush")
		poolPopFn = mustSym("objc_autoreleasePoolPop")
		mainNPFn = mustSym("pthread_main_np")
		mainQueue = mustSym("_dispatch_main_q") // dispatch_get_main_queue()
		asyncFFn = mustSym("dispatch_async_f")
		syncFFn = mustSym("dispatch_sync_f")
		dispatchFn = purego.NewCallback(runDispatched)
		purego.RegisterFunc(&MsgRect, stret)
		purego.RegisterFunc(&MsgSetRect, msgSendAddr)
		purego.RegisterFunc(&MsgSetPoint, msgSendAddr)
		purego.RegisterFunc(&MsgSetSize, msgSendAddr)
		purego.RegisterFunc(&MsgFloat, msgSendAddr)
		purego.RegisterFunc(&MsgSetFloat, msgSendAddr)
		purego.RegisterFunc(&MsgPoint, msgSendAddr)
		purego.RegisterFunc(&MsgSize, msgSendAddr)
		purego.RegisterFunc(&MsgIDForPoint, msgSendAddr)
		purego.RegisterFunc(&MsgInitRect, msgSendAddr)
		purego.RegisterFunc(&MsgPointFromView, msgSendAddr)
		purego.RegisterFunc(&MsgKeyEvent, msgSendAddr)
	})
}

// Sym returns the address of a C symbol, or 0.
func Sym(name string) uintptr {
	load()
	p, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if err != nil {
		return 0
	}
	return p
}

func mustSym(name string) uintptr {
	p, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if err != nil {
		panic("keel: native symbol " + name + " not found")
	}
	return p
}

// Call calls a C function taking and returning integers or pointers.
//
//go:uintptrescapes
func Call(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// Constant reads an exported object constant, such as NSApp or
// NSAppearanceNameDarkAqua; 0 when this system does not have it.
func Constant(name string) ID {
	p := Sym(name)
	if p == 0 {
		return 0
	}
	return **(**ID)(unsafe.Pointer(&p))
}

var (
	selMu    sync.RWMutex
	selCache = map[string]SEL{}
)

func Sel(name string) SEL {
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

// Class returns a class, or 0 when this system does not have it.
func Class(name string) ID { load(); return ID(objc.GetClass(name)) }

// Send sends a message whose arguments are integers or pointers.
//
//go:uintptrescapes
func Send(obj ID, selector string, args ...uintptr) ID {
	load()
	var a [10]uintptr
	a[0], a[1] = uintptr(obj), uintptr(Sel(selector))
	n := copy(a[2:], args)
	r, _, _ := purego.SyscallN(msgSendAddr, a[:n+2]...)
	return ID(r)
}

//go:uintptrescapes
func SendBool(obj ID, selector string, args ...uintptr) bool {
	return byte(Send(obj, selector, args...)) != 0
}

func Bool(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

func Retain(obj ID) ID {
	if obj == 0 {
		return 0
	}
	return Send(obj, "retain")
}

func Release(obj ID) {
	if obj != 0 {
		Send(obj, "release")
	}
}

// App returns NSApp, or 0 before the application object exists.
func App() ID { return Constant("NSApp") }

// Pool runs fn in an autorelease pool on a locked thread: a pool must be
// popped on the thread that pushed it.
func Pool(fn func()) {
	load()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool, _, _ := purego.SyscallN(poolPushFn)
	defer purego.SyscallN(poolPopFn, pool)
	fn()
}

func IsMainThread() bool {
	load()
	r, _, _ := purego.SyscallN(mainNPFn)
	return int32(r) != 0
}

const utf8Encoding = 4

// String returns an autoreleased NSString.
func String(s string) ID {
	b := unsafe.StringData(s)
	str := Send(Send(Class("NSString"), "alloc"), "initWithBytes:length:encoding:",
		uintptr(unsafe.Pointer(b)), uintptr(len(s)), utf8Encoding)
	runtime.KeepAlive(s)
	return Send(str, "autorelease")
}

// GoString copies an NSString.
func GoString(str ID) string {
	if str == 0 {
		return ""
	}
	p := Send(str, "UTF8String")
	if p == 0 {
		return ""
	}
	n := int(Send(str, "lengthOfBytesUsingEncoding:", utf8Encoding))
	return string(unsafe.Slice(*(**byte)(unsafe.Pointer(&p)), n))
}

// Data returns an autoreleased NSData with a copy of b.
func Data(b []byte) ID {
	var p unsafe.Pointer
	if len(b) > 0 {
		p = unsafe.Pointer(&b[0])
	}
	d := Send(Class("NSData"), "dataWithBytes:length:", uintptr(p), uintptr(len(b)))
	runtime.KeepAlive(b)
	return d
}

// Equal reports whether two NSStrings hold the same text.
func Equal(str ID, s string) bool {
	return str != 0 && SendBool(str, "isEqualToString:", uintptr(String(s)))
}

// NewBlock creates a block backed by fn, whose first parameter is
// objc.Block. purego creates one callback per signature, reused.
func NewBlock(fn any) objc.Block { return objc.NewBlock(fn) }

// RegisterClass defines an Objective-C class once; methods are
// objc.MethodDef values. Protocols missing on this system are skipped.
func RegisterClass(name, super string, protocols []string, methods []objc.MethodDef) ID {
	load()
	var ps []*objc.Protocol
	for _, p := range protocols {
		if proto := objc.GetProtocol(p); proto != nil {
			ps = append(ps, proto)
		}
	}
	c, err := objc.RegisterClass(name, objc.GetClass(super), ps, nil, methods)
	if err != nil {
		panic("keel: cannot register class " + name + ": " + err.Error())
	}
	return ID(c)
}

func Method(selector string, fn any) objc.MethodDef {
	return objc.MethodDef{Cmd: Sel(selector), Fn: fn}
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
		Pool(fn)
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

// MainAsync runs fn on the main queue later, never waiting for it: callers
// may hold the frame lock.
func MainAsync(fn func()) {
	load()
	purego.SyscallN(asyncFFn, mainQueue, dispatchToken(fn), dispatchFn)
}

// MainSync runs fn on the main thread and waits. Never call it with the
// frame lock held.
func MainSync(fn func()) {
	load()
	if IsMainThread() {
		Pool(fn)
		return
	}
	purego.SyscallN(syncFFn, mainQueue, dispatchToken(fn), dispatchFn)
}
