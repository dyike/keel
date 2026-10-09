//go:build linux && !android && !nowayland

// Package wlclip reads the Wayland clipboard through an application's own
// wl_display. Wayland offers the selection only to the client with keyboard
// focus, so a separate connection would see nothing; the caller passes the
// display of its focused window. libwayland-client is loaded at run time
// through purego, without cgo; build with -tags nowayland to leave it out.
//
// Only the core protocol is used, marshalled by opcode from wayland.xml with
// wl_proxy_marshal_array_flags (libwayland 1.20+), so no generated protocol
// code is needed. Events go through one dispatcher for every proxy
// (wl_proxy_add_dispatcher), created once: purego callbacks are never freed.
// Everything runs on the caller's goroutine, on a private event queue, so
// Gio's own event handling is untouched.
package wlclip

import (
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const maxMIMEs = 64

// Request opcodes and versions from wayland.xml.
const (
	displayGetRegistry   = 1
	registryBind         = 0
	managerGetDataDevice = 1
	deviceRelease        = 2
	offerReceive         = 1
	offerDestroy         = 2
	marshalFlagDestroy   = 1
)

// Event opcodes.
const (
	registryGlobal    = 0
	deviceDataOffer   = 0
	deviceSelection   = 5
	offerOffer        = 0
	dispatchRegistry  = 1 // dispatcher data: which kind of proxy sent it
	dispatchDevice    = 2
	dispatchOffer     = 3
	dispatchedHandled = 0
)

var lib struct {
	once sync.Once
	ok   bool

	registryIface, seatIface, managerIface, deviceIface uintptr
	dispatcher                                          uintptr

	marshalArrayFlags, getVersion, addDispatcher, getUserData uintptr
	destroy, createQueue, destroyQueue, createWrapper         uintptr
	destroyWrapper, setQueue, roundtripQueue, flush           uintptr
}

func load() bool {
	lib.once.Do(func() {
		h, err := purego.Dlopen("libwayland-client.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		sym := func(name string) uintptr {
			p, err := purego.Dlsym(h, name)
			if err != nil {
				return 0
			}
			return p
		}
		for _, s := range []struct {
			dst  *uintptr
			name string
		}{
			{&lib.registryIface, "wl_registry_interface"},
			{&lib.seatIface, "wl_seat_interface"},
			{&lib.managerIface, "wl_data_device_manager_interface"},
			{&lib.deviceIface, "wl_data_device_interface"},
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
		} {
			if *s.dst = sym(s.name); *s.dst == 0 {
				return // older than 1.20, or not libwayland
			}
		}
		lib.dispatcher = purego.NewCallback(dispatch)
		lib.ok = true
	})
	return lib.ok
}

func c(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// marshal sends a request; args are wl_argument unions, 8 bytes each. Go
// memory they point to (strings) must stay alive and unmoved until it
// returns, which libwayland's copy into its buffer guarantees.
func marshal(proxy uintptr, opcode uint32, iface uintptr, version uint32, flags uint32, args ...uintptr) uintptr {
	var a [8]uintptr
	copy(a[:], args)
	r := c(lib.marshalArrayFlags, proxy, uintptr(opcode), iface, uintptr(version), uintptr(flags), uintptr(unsafe.Pointer(&a)))
	runtime.KeepAlive(&a)
	return r
}

func version(proxy uintptr) uint32 { return uint32(c(lib.getVersion, proxy)) }

// cString returns a NUL-terminated copy, pinned by the caller's KeepAlive.
func cString(s string) []byte { return append([]byte(s), 0) }

// ptr converts an address libwayland gave us.
func ptr(p uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&p)) }

func goString(p uintptr) string {
	if p == 0 {
		return ""
	}
	return unix.BytePtrToString((*byte)(ptr(p)))
}

// Open selections by token: proxies carry the token as user data, never a
// Go pointer.
var selections struct {
	sync.Mutex
	next uintptr
	m    map[uintptr]*Selection
}

type offer struct {
	proxy uintptr
	mimes []string
}

// Selection is the clipboard's offer at the time of Open.
type Selection struct {
	token                             uintptr
	display, queue, wrapper, registry uintptr
	seat, manager, device             uintptr
	managerVersion                    uint32
	offers                            []*offer // every offer announced, owned until Close
	selection                         *offer   // the clipboard's, or nil when empty
	seatIfaceName, managerIfaceName   []byte
}

func (s *Selection) find(proxy uintptr) *offer {
	for _, o := range s.offers {
		if o.proxy == proxy {
			return o
		}
	}
	return nil
}

func (s *Selection) listen(proxy, kind uintptr) {
	c(lib.addDispatcher, proxy, lib.dispatcher, kind, s.token)
}

// dispatch receives every event of the proxies wlclip listens to, on the
// goroutine that called wl_display_roundtrip_queue.
func dispatch(kind, proxy, opcode, _, args uintptr) uintptr {
	selections.Lock()
	s := selections.m[c(lib.getUserData, proxy)]
	selections.Unlock()
	if s == nil {
		return dispatchedHandled
	}
	arg := func(i int) uintptr { return *(*uintptr)(ptr(args + uintptr(i)*8)) }
	switch {
	case kind == dispatchRegistry && opcode == registryGlobal:
		name, iface, v := uint32(arg(0)), goString(arg(1)), uint32(arg(2))
		if s.seat == 0 && iface == "wl_seat" {
			// Version 1 is enough to name the seat and needs no listener.
			s.seat = marshal(proxy, registryBind, lib.seatIface, 1, 0,
				uintptr(name), uintptr(unsafe.Pointer(&s.seatIfaceName[0])), 1, 0)
		} else if s.manager == 0 && iface == "wl_data_device_manager" {
			s.managerVersion = min(v, 3)
			s.manager = marshal(proxy, registryBind, lib.managerIface, s.managerVersion, 0,
				uintptr(name), uintptr(unsafe.Pointer(&s.managerIfaceName[0])), uintptr(s.managerVersion), 0)
		}
	case kind == dispatchDevice && opcode == deviceDataOffer:
		o := &offer{proxy: arg(0)}
		s.offers = append(s.offers, o)
		s.listen(o.proxy, dispatchOffer)
	case kind == dispatchDevice && opcode == deviceSelection:
		s.selection = nil
		if p := arg(0); p != 0 {
			s.selection = s.find(p)
		}
	case kind == dispatchOffer && opcode == offerOffer:
		if o := s.find(proxy); o != nil && len(o.mimes) < maxMIMEs {
			o.mimes = append(o.mimes, goString(arg(0)))
		}
	}
	return dispatchedHandled
}

// Open binds the seat's data device on display and waits for the current
// selection. Close it when done.
func Open(display unsafe.Pointer) (*Selection, error) {
	if display == nil || !load() {
		return nil, native.ErrUnsupported
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	s := &Selection{display: uintptr(display)}
	s.seatIfaceName, s.managerIfaceName = cString("wl_seat"), cString("wl_data_device_manager")
	selections.Lock()
	if selections.m == nil {
		selections.m = map[uintptr]*Selection{}
	}
	selections.next++
	s.token = selections.next
	selections.m[s.token] = s
	selections.Unlock()
	fail := func(err error) (*Selection, error) { s.Close(); return nil, err }

	if s.queue = c(lib.createQueue, s.display); s.queue == 0 {
		return fail(native.ErrFailed)
	}
	if s.wrapper = c(lib.createWrapper, s.display); s.wrapper == 0 {
		return fail(native.ErrFailed)
	}
	c(lib.setQueue, s.wrapper, s.queue)
	if s.registry = marshal(s.wrapper, displayGetRegistry, lib.registryIface, version(s.wrapper), 0, 0); s.registry == 0 {
		return fail(native.ErrFailed)
	}
	s.listen(s.registry, dispatchRegistry)
	if int32(c(lib.roundtripQueue, s.display, s.queue)) < 0 {
		return fail(native.ErrFailed)
	}
	if s.seat == 0 || s.manager == 0 {
		return fail(native.ErrUnsupported)
	}
	if s.device = marshal(s.manager, managerGetDataDevice, lib.deviceIface, version(s.manager), 0, 0, s.seat); s.device == 0 {
		return fail(native.ErrFailed)
	}
	s.listen(s.device, dispatchDevice)
	// The compositor announces the selection to a focused client's new data
	// device; a second roundtrip lets the offer's MIME types arrive too.
	if int32(c(lib.roundtripQueue, s.display, s.queue)) < 0 || int32(c(lib.roundtripQueue, s.display, s.queue)) < 0 {
		return fail(native.ErrFailed)
	}
	return s, nil
}

// MIMEs lists the offered types; nil when the clipboard is empty.
func (s *Selection) MIMEs() []string {
	if s.selection == nil {
		return nil
	}
	return append([]string(nil), s.selection.mimes...)
}

// Read receives one type, at most limit bytes, by the deadline. A larger
// payload is an error rather than a truncated value.
func (s *Selection) Read(mime string, limit int, deadline time.Time) ([]byte, error) {
	if s.selection == nil {
		return nil, native.ErrFailed
	}
	var fds [2]int
	if err := unix.Pipe2(fds[:], unix.O_CLOEXEC|unix.O_NONBLOCK); err != nil {
		return nil, native.ErrFailed
	}
	name := cString(mime)
	p := s.selection.proxy
	// The owner writes into the pipe; libwayland duplicates the fd it sends.
	marshal(p, offerReceive, 0, version(p), 0, uintptr(unsafe.Pointer(&name[0])), uintptr(fds[1]))
	runtime.KeepAlive(name)
	unix.Close(fds[1])
	if r, _, errno := purego.SyscallN(lib.flush, s.display); int32(r) < 0 && unix.Errno(errno) != unix.EAGAIN {
		unix.Close(fds[0])
		return nil, native.ErrFailed
	}
	f := os.NewFile(uintptr(fds[0]), "wayland-selection")
	defer f.Close()
	f.SetReadDeadline(deadline)
	data, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return nil, native.ErrTimeout
	}
	if err != nil {
		return nil, native.ErrFailed
	}
	if len(data) > limit {
		return nil, errors.New("wayland selection exceeds the limit")
	}
	return data, nil
}

// Close releases the data device and offers.
func (s *Selection) Close() {
	for _, o := range s.offers {
		marshal(o.proxy, offerDestroy, 0, version(o.proxy), marshalFlagDestroy)
	}
	s.offers, s.selection = nil, nil
	if s.device != 0 {
		if s.managerVersion >= 2 {
			marshal(s.device, deviceRelease, 0, version(s.device), marshalFlagDestroy)
		} else {
			c(lib.destroy, s.device)
		}
	}
	for _, p := range []uintptr{s.manager, s.seat, s.registry} {
		if p != 0 {
			c(lib.destroy, p)
		}
	}
	if s.wrapper != 0 {
		c(lib.destroyWrapper, s.wrapper)
	}
	if lib.ok {
		c(lib.flush, s.display)
	}
	if s.queue != 0 {
		c(lib.destroyQueue, s.queue)
	}
	s.device, s.manager, s.seat, s.registry, s.wrapper, s.queue = 0, 0, 0, 0, 0, 0
	selections.Lock()
	delete(selections.m, s.token)
	selections.Unlock()
}
