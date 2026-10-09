//go:build linux && !android && !nowayland

package wayland

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var lib struct {
	once   sync.Once
	ok     bool
	handle uintptr

	dispatcher uintptr

	marshalArrayFlags, getVersion, addDispatcher, getUserData uintptr
	destroy, createQueue, destroyQueue, createWrapper         uintptr
	destroyWrapper, setQueue, roundtripQueue, flush           uintptr
	dispatchQueuePending                                      uintptr
}

// Load opens libwayland-client 1.20+ once; false when it is missing.
func Load() bool {
	lib.once.Do(func() {
		h, err := purego.Dlopen("libwayland-client.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		lib.handle = h
		for _, s := range []struct {
			dst  *uintptr
			name string
		}{
			{&lib.marshalArrayFlags, "wl_proxy_marshal_array_flags"},
			{&lib.getVersion, "wl_proxy_get_version"},
			{&lib.addDispatcher, "wl_proxy_add_dispatcher"},
			{&lib.getUserData, "wl_proxy_get_user_data"},
			{&lib.destroy, "wl_proxy_destroy"},
			{&lib.createQueue, "wl_display_create_queue"},
			{&lib.destroyQueue, "wl_event_queue_destroy"},
			{&lib.createWrapper, "wl_proxy_create_wrapper"},
			{&lib.destroyWrapper, "wl_proxy_wrapper_destroy"},
			{&lib.setQueue, "wl_proxy_set_queue"},
			{&lib.roundtripQueue, "wl_display_roundtrip_queue"},
			{&lib.flush, "wl_display_flush"},
			{&lib.dispatchQueuePending, "wl_display_dispatch_queue_pending"},
		} {
			p, err := purego.Dlsym(h, s.name)
			if err != nil {
				return // older than 1.20, or not libwayland
			}
			*s.dst = p
		}
		lib.dispatcher = purego.NewCallback(dispatch)
		lib.ok = true
	})
	return lib.ok
}

// Interface returns the address of a core protocol interface exported by
// libwayland, such as "wl_seat_interface", or 0.
func Interface(symbol string) uintptr {
	if !Load() {
		return 0
	}
	p, err := purego.Dlsym(lib.handle, symbol)
	if err != nil {
		return 0
	}
	return p
}

func call(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// Marshal sends a request, returning the new proxy of a constructor. args
// are wl_argument unions: integers, objects, fds, C strings (see CString)
// and 0 for a new_id placeholder.
func Marshal(proxy uintptr, opcode uint32, iface uintptr, version uint32, flags uint32, args ...uintptr) uintptr {
	var a [8]uintptr
	copy(a[:], args)
	r := call(lib.marshalArrayFlags, proxy, uintptr(opcode), iface, uintptr(version), uintptr(flags), uintptr(unsafe.Pointer(&a)))
	runtime.KeepAlive(&a)
	return r
}

func Version(proxy uintptr) uint32        { return uint32(call(lib.getVersion, proxy)) }
func Destroy(proxy uintptr)               { call(lib.destroy, proxy) }
func CreateQueue(display uintptr) uintptr { return call(lib.createQueue, display) }
func DestroyQueue(queue uintptr)          { call(lib.destroyQueue, queue) }
func DestroyWrapper(wrapper uintptr)      { call(lib.destroyWrapper, wrapper) }

// Wrapper returns a wrapper of display whose new objects use queue.
func Wrapper(display, queue uintptr) uintptr {
	w := call(lib.createWrapper, display)
	if w != 0 {
		call(lib.setQueue, w, queue)
	}
	return w
}

func Roundtrip(display, queue uintptr) bool {
	return int32(call(lib.roundtripQueue, display, queue)) >= 0
}
func Flush(display uintptr) bool             { return int32(call(lib.flush, display)) >= 0 }
func DispatchPending(display, queue uintptr) { call(lib.dispatchQueuePending, display, queue) }

// CString returns a NUL-terminated copy; keep it alive until Marshal returns.
func CString(s string) []byte { return append([]byte(s), 0) }

// Ptr is the address of a CString for Marshal.
func Ptr(b []byte) uintptr { return uintptr(unsafe.Pointer(&b[0])) }

func ptr(p uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&p)) }

// GoString copies a C string from an event argument.
func GoString(p uintptr) string {
	if p == 0 {
		return ""
	}
	n := 0
	for *(*byte)(ptr(p + uintptr(n))) != 0 {
		n++
	}
	return string(unsafe.Slice((*byte)(ptr(p)), n))
}

// Handler receives an event of a proxy listened to with Listen: kind is the
// value given to Listen, args the event's wl_argument array (see Arg).
type Handler func(kind, proxy uintptr, opcode uint32, args uintptr)

// Arg returns argument i of an event.
func Arg(args uintptr, i int) uintptr { return *(*uintptr)(ptr(args + uintptr(i)*8)) }

// Handlers by token: proxies carry the token as user data, never a Go
// pointer, and one purego callback serves every proxy.
var handlers struct {
	sync.Mutex
	next uintptr
	m    map[uintptr]Handler
}

// Register returns the token that routes events to h.
func Register(h Handler) uintptr {
	handlers.Lock()
	defer handlers.Unlock()
	if handlers.m == nil {
		handlers.m = map[uintptr]Handler{}
	}
	handlers.next++
	handlers.m[handlers.next] = h
	return handlers.next
}

// Unregister drops a token; events still queued for it are ignored.
func Unregister(token uintptr) {
	handlers.Lock()
	delete(handlers.m, token)
	handlers.Unlock()
}

// Listen routes the events of proxy to the handler of token, with kind.
func Listen(proxy, token, kind uintptr) {
	call(lib.addDispatcher, proxy, lib.dispatcher, kind, token)
}

// dispatch runs on the goroutine dispatching the queue.
func dispatch(kind, proxy, opcode, _, args uintptr) uintptr {
	handlers.Lock()
	h := handlers.m[call(lib.getUserData, proxy)]
	handlers.Unlock()
	if h != nil {
		h(kind, proxy, uint32(opcode), args)
	}
	return 0
}

// Request opcodes from wayland.xml shared by users of this package.
const (
	DisplayGetRegistry = 1
	RegistryBind       = 0
	RegistryGlobal     = 0 // event
	MarshalFlagDestroy = 1
)
