//go:build linux && !android && !nowayland

package window

import (
	"runtime"
	"unsafe"

	"github.com/dyike/keel/ui/internal/wayland"
)

// xdg-activation-v1 from wayland-protocols (staging), written out by hand:
// Activate needs one request on Gio's own connection and surface. The tables
// mirror struct wl_message and struct wl_interface; they live in Go globals,
// which never move, and libwayland only reads them.

type wlMessage struct {
	name, signature unsafe.Pointer // C strings
	types           unsafe.Pointer // *[n]*wl_interface
}

type wlInterface struct {
	name                 unsafe.Pointer
	version, methodCount int32
	methods              unsafe.Pointer
	eventCount           int32
	_                    int32
	events               unsafe.Pointer
}

// addr converts the address of a libwayland symbol.
func addr(p uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&p)) }

func cstr(s string) unsafe.Pointer { return unsafe.Pointer(&append([]byte(s), 0)[0]) }

var xdgActivation struct {
	types             [6]unsafe.Pointer
	tokenRequests     [5]wlMessage
	tokenEvents       [1]wlMessage
	requests          [3]wlMessage
	token, activation wlInterface
	ready             bool
}

// Builds the tables once libwayland is loaded; the main loop of Gio's
// Wayland backend calls Activate, so this needs no lock beyond the caller's.
func xdgActivationInterface() uintptr {
	x := &xdgActivation
	if !x.ready {
		types := func(i int) unsafe.Pointer { return unsafe.Pointer(&x.types[i]) }
		surface := addr(wayland.Interface("wl_surface_interface"))
		seat := addr(wayland.Interface("wl_seat_interface"))
		x.types = [6]unsafe.Pointer{
			unsafe.Pointer(&x.token), // get_activation_token: id
			nil, surface,             // activate: token, surface
			nil, seat, // token set_serial: serial, seat
			surface, // token set_surface: surface
		}
		x.tokenRequests = [5]wlMessage{
			{cstr("set_serial"), cstr("uo"), types(3)},
			{cstr("set_app_id"), cstr("s"), types(1)},
			{cstr("set_surface"), cstr("o"), types(5)},
			{cstr("commit"), cstr(""), types(1)},
			{cstr("destroy"), cstr(""), types(1)},
		}
		x.tokenEvents = [1]wlMessage{{cstr("done"), cstr("s"), types(1)}}
		x.token = wlInterface{name: cstr("xdg_activation_token_v1"), version: 1, methodCount: 5,
			methods: unsafe.Pointer(&x.tokenRequests), eventCount: 1, events: unsafe.Pointer(&x.tokenEvents)}
		x.requests = [3]wlMessage{
			{cstr("destroy"), cstr(""), types(1)},
			{cstr("get_activation_token"), cstr("n"), types(0)},
			{cstr("activate"), cstr("so"), types(1)},
		}
		x.activation = wlInterface{name: cstr("xdg_activation_v1"), version: 1, methodCount: 3, methods: unsafe.Pointer(&x.requests)}
		x.ready = true
	}
	return uintptr(unsafe.Pointer(&x.activation))
}

const (
	activationDestroy  = 0
	activationActivate = 2
)

// waylandActivate returns 0 on success, 1 without xdg-activation, -1 on error.
func waylandActivate(display, surface unsafe.Pointer, token string) int {
	if !wayland.Load() || display == nil {
		return -1
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	d := uintptr(display)
	iface := xdgActivationInterface()
	queue := wayland.CreateQueue(d)
	if queue == 0 {
		return -1
	}
	defer wayland.DestroyQueue(queue)
	wrapper := wayland.Wrapper(d, queue)
	if wrapper == 0 {
		return -1
	}
	defer wayland.DestroyWrapper(wrapper)
	var activation uintptr
	name := wayland.CString("xdg_activation_v1")
	listener := wayland.Register(func(_, registry uintptr, opcode uint32, args uintptr) {
		if opcode == wayland.RegistryGlobal && activation == 0 && wayland.GoString(wayland.Arg(args, 1)) == "xdg_activation_v1" {
			activation = wayland.Marshal(registry, wayland.RegistryBind, iface, 1, 0, wayland.Arg(args, 0), wayland.Ptr(name), 1, 0)
		}
	})
	defer wayland.Unregister(listener)
	registry := wayland.Marshal(wrapper, wayland.DisplayGetRegistry, wayland.Interface("wl_registry_interface"), wayland.Version(wrapper), 0, 0)
	if registry == 0 {
		return -1
	}
	defer wayland.Destroy(registry)
	wayland.Listen(registry, listener, 0)
	if !wayland.Roundtrip(d, queue) {
		return -1
	}
	if activation == 0 {
		return 1
	}
	t := wayland.CString(token)
	wayland.Marshal(activation, activationActivate, 0, 1, 0, wayland.Ptr(t), uintptr(surface))
	runtime.KeepAlive(t)
	wayland.Marshal(activation, activationDestroy, 0, 1, wayland.MarshalFlagDestroy)
	runtime.KeepAlive(name)
	if !wayland.Flush(d) {
		return -1
	}
	return 0
}
