// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && (!cgo || purego)

package app

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	syscall "golang.org/x/sys/unix"
)

// The libwayland calls of os_wayland.go without cgo. libwayland-client and
// libwayland-cursor are loaded at run time; the xdg-shell, xdg-decoration
// and text-input interfaces are built in C memory from
// wayland_protocols_purego.go. Every proxy delivers its events to one
// dispatcher (wl_proxy_add_dispatcher), created once, which decodes the
// arguments and calls the wlOn* handlers. Keel patch.

var wl struct {
	once sync.Once
	err  error

	displayConnect, displayDisconnect, displayGetFD, displayFlush uintptr
	displayRoundtrip, displayDispatchPending, displayPrepareRead  uintptr
	displayReadEvents, displayCancelRead                          uintptr
	proxyMarshalArrayFlags, proxyGetVersion, proxyDestroy         uintptr
	proxyAddDispatcher, proxyGetUserData                          uintptr
	cursorThemeLoad, cursorThemeGetCursor, cursorImageGetBuffer   uintptr
	cursorThemeDestroy                                            uintptr
	malloc, free, errnoLocation                                   uintptr
	dispatcher                                                    uintptr
}

func loadWayland() error {
	wl.once.Do(func() { wl.err = initWayland() })
	return wl.err
}

func initWayland() error {
	client, err := openLib("libwayland-client.so.0", "libwayland-client.so")
	if err != nil {
		return fmt.Errorf("wayland: %w", err)
	}
	cursor, err := openLib("libwayland-cursor.so.0", "libwayland-cursor.so")
	if err != nil {
		return fmt.Errorf("wayland: %w", err)
	}
	libc, err := openLib("libc.so.6", "libc.so")
	if err != nil {
		return fmt.Errorf("wayland: %w", err)
	}
	for _, s := range []struct {
		lib  uintptr
		name string
		addr *uintptr
	}{
		{client, "wl_display_connect", &wl.displayConnect},
		{client, "wl_display_disconnect", &wl.displayDisconnect},
		{client, "wl_display_get_fd", &wl.displayGetFD},
		{client, "wl_display_flush", &wl.displayFlush},
		{client, "wl_display_roundtrip", &wl.displayRoundtrip},
		{client, "wl_display_dispatch_pending", &wl.displayDispatchPending},
		{client, "wl_display_prepare_read", &wl.displayPrepareRead},
		{client, "wl_display_read_events", &wl.displayReadEvents},
		{client, "wl_display_cancel_read", &wl.displayCancelRead},
		{client, "wl_proxy_marshal_array_flags", &wl.proxyMarshalArrayFlags},
		{client, "wl_proxy_get_version", &wl.proxyGetVersion},
		{client, "wl_proxy_destroy", &wl.proxyDestroy},
		{client, "wl_proxy_add_dispatcher", &wl.proxyAddDispatcher},
		{client, "wl_proxy_get_user_data", &wl.proxyGetUserData},
		{cursor, "wl_cursor_theme_load", &wl.cursorThemeLoad},
		{cursor, "wl_cursor_theme_get_cursor", &wl.cursorThemeGetCursor},
		{cursor, "wl_cursor_image_get_buffer", &wl.cursorImageGetBuffer},
		{cursor, "wl_cursor_theme_destroy", &wl.cursorThemeDestroy},
		{libc, "malloc", &wl.malloc},
		{libc, "free", &wl.free},
		{libc, "__errno_location", &wl.errnoLocation},
	} {
		if *s.addr, err = purego.Dlsym(s.lib, s.name); err != nil {
			return fmt.Errorf("wayland: %w", err)
		}
	}
	// The core interfaces are data symbols of libwayland-client.
	core := map[string]*unsafe.Pointer{
		"wl_registry": &wlIface.wl_registry, "wl_compositor": &wlIface.wl_compositor,
		"wl_surface": &wlIface.wl_surface, "wl_region": &wlIface.wl_region,
		"wl_callback": &wlIface.wl_callback, "wl_seat": &wlIface.wl_seat,
		"wl_pointer": &wlIface.wl_pointer, "wl_keyboard": &wlIface.wl_keyboard,
		"wl_touch": &wlIface.wl_touch, "wl_output": &wlIface.wl_output,
		"wl_shm": &wlIface.wl_shm, "wl_data_device_manager": &wlIface.wl_data_device_manager,
		"wl_data_device": &wlIface.wl_data_device, "wl_data_source": &wlIface.wl_data_source,
	}
	ifaces := map[string]unsafe.Pointer{}
	for name, dst := range core {
		addr, err := purego.Dlsym(client, name+"_interface")
		if err != nil {
			return fmt.Errorf("wayland: %w", err)
		}
		*dst = cptr(addr)
		ifaces[name] = *dst
	}
	// wl_output and wl_seat are referenced by the protocols' types.
	if err := buildProtocols(ifaces); err != nil {
		return err
	}
	for name, dst := range map[string]*unsafe.Pointer{
		"xdg_wm_base": &wlIface.xdg_wm_base, "xdg_surface": &wlIface.xdg_surface,
		"xdg_toplevel":                &wlIface.xdg_toplevel,
		"zxdg_decoration_manager_v1":  &wlIface.zxdg_decoration_manager_v1,
		"zxdg_toplevel_decoration_v1": &wlIface.zxdg_toplevel_decoration_v1,
		"zwp_text_input_manager_v3":   &wlIface.zwp_text_input_manager_v3,
		"zwp_text_input_v3":           &wlIface.zwp_text_input_v3,
	} {
		if *dst = ifaces[name]; *dst == nil {
			return errors.New("wayland: interface " + name + " missing")
		}
	}
	wl.dispatcher = purego.NewCallback(wlDispatch)
	return nil
}

// wl_interface and wl_message, for C memory written from Go.
type cWLInterface struct {
	name                 unsafe.Pointer
	version, methodCount int32
	methods              unsafe.Pointer
	eventCount           int32
	events               unsafe.Pointer
}

type cWLMessage struct {
	name, signature, types unsafe.Pointer
}

// cAlloc returns zeroed C memory, never freed: interfaces live as long as
// the process.
func cAlloc(n uintptr) unsafe.Pointer {
	p := cptr(xcall(wl.malloc, n))
	clear(unsafe.Slice((*byte)(p), n))
	return p
}

// buildProtocols builds the protocols' interfaces, adding them to ifaces.
// Interfaces are allocated before their messages, which refer to them.
func buildProtocols(ifaces map[string]unsafe.Pointer) error {
	for _, p := range wlProtocols {
		for _, i := range p.interfaces {
			ifaces[i.name] = cAlloc(unsafe.Sizeof(cWLInterface{}))
		}
	}
	for _, p := range wlProtocols {
		types := (*[1 << 16]unsafe.Pointer)(cAlloc(uintptr(len(p.types)) * unsafe.Sizeof(unsafe.Pointer(nil))))[:len(p.types):len(p.types)]
		for k, name := range p.types {
			if name == "" {
				continue
			}
			if types[k] = ifaces[name]; types[k] == nil {
				return errors.New("wayland: protocol type " + name + " missing")
			}
		}
		messages := func(ms []wlProtocolMessage) unsafe.Pointer {
			if len(ms) == 0 {
				return nil
			}
			arr := (*[1 << 12]cWLMessage)(cAlloc(uintptr(len(ms)) * unsafe.Sizeof(cWLMessage{})))[:len(ms):len(ms)]
			for k, m := range ms {
				arr[k] = cWLMessage{wlCString(m.name), wlCString(m.signature), unsafe.Pointer(&types[m.types])}
			}
			return unsafe.Pointer(&arr[0])
		}
		for _, i := range p.interfaces {
			*(*cWLInterface)(ifaces[i.name]) = cWLInterface{
				name:        wlCString(i.name),
				version:     i.version,
				methodCount: int32(len(i.requests)),
				methods:     messages(i.requests),
				eventCount:  int32(len(i.events)),
				events:      messages(i.events),
			}
		}
	}
	return nil
}

func wlCString(s string) unsafe.Pointer {
	p := cptr(xcall(wl.malloc, uintptr(len(s)+1)))
	b := unsafe.Slice((*byte)(p), len(s)+1)
	copy(b, s)
	b[len(s)] = 0
	return p
}

func wlFree(p unsafe.Pointer) { xcall(wl.free, ptr(p)) }

// errno reads the calling thread's errno; callers stay on their thread.
func errno() syscall.Errno {
	return syscall.Errno(*(*int32)(cptr(xcall(wl.errnoLocation))))
}

func wlDisplayConnect() (unsafe.Pointer, error) {
	if err := loadWayland(); err != nil {
		return nil, err
	}
	d := cptr(xcall(wl.displayConnect, 0))
	if d == nil {
		return nil, errors.New("wl_display_connect failed")
	}
	return d, nil
}

func wlDisplayDisconnect(d unsafe.Pointer)        { xcall(wl.displayDisconnect, ptr(d)) }
func wlDisplayGetFd(d unsafe.Pointer) int32       { return i32(xcall(wl.displayGetFD, ptr(d))) }
func wlDisplayRoundtrip(d unsafe.Pointer)         { xcall(wl.displayRoundtrip, ptr(d)) }
func wlDisplayDispatchPending(d unsafe.Pointer)   { xcall(wl.displayDispatchPending, ptr(d)) }
func wlDisplayPrepareRead(d unsafe.Pointer) int32 { return i32(xcall(wl.displayPrepareRead, ptr(d))) }
func wlDisplayCancelRead(d unsafe.Pointer)        { xcall(wl.displayCancelRead, ptr(d)) }
func wlProxyGetVersion(p unsafe.Pointer) uint32   { return uint32(xcall(wl.proxyGetVersion, ptr(p))) }
func wlProxyDestroy(p unsafe.Pointer)             { xcall(wl.proxyDestroy, ptr(p)) }
func wlCursorThemeDestroy(theme unsafe.Pointer)   { xcall(wl.cursorThemeDestroy, ptr(theme)) }
func wlProxyGetUserData(p unsafe.Pointer) unsafe.Pointer {
	return cptr(xcall(wl.proxyGetUserData, ptr(p)))
}

// wlDisplayFlush and wlDisplayReadEvents return errno with a negative
// result; dispatch runs on a locked thread.
func wlDisplayFlush(d unsafe.Pointer) (int32, error) {
	if r := i32(xcall(wl.displayFlush, ptr(d))); r < 0 {
		return r, errno()
	}
	return 0, nil
}

func wlDisplayReadEvents(d unsafe.Pointer) (int32, error) {
	if r := i32(xcall(wl.displayReadEvents, ptr(d))); r < 0 {
		return r, errno()
	}
	return 0, nil
}

func wlProxyMarshalArrayFlags(p unsafe.Pointer, op uint32, iface unsafe.Pointer, version, flags uint32, args []uintptr) unsafe.Pointer {
	return cptr(xcall(wl.proxyMarshalArrayFlags, ptr(p), uintptr(op), ptr(iface), uintptr(version), uintptr(flags), uintptr(unsafe.Pointer(&args[0]))))
}

func wlAddListener(p unsafe.Pointer, kind int, data unsafe.Pointer) {
	// The implementation pointer carries the listener kind, plus one: never
	// NULL.
	xcall(wl.proxyAddDispatcher, ptr(p), wl.dispatcher, uintptr(kind)+1, ptr(data))
}

func wlCursorThemeLoad(name string, size int, shm unsafe.Pointer) unsafe.Pointer {
	// An empty name, as os_wayland_cgo.go passes, selects the default theme.
	cname := wlCString(name)
	defer wlFree(cname)
	return cptr(xcall(wl.cursorThemeLoad, ptr(cname), uintptr(size), ptr(shm)))
}

func wlCursorThemeGetCursor(theme unsafe.Pointer, name string) *wlCursor {
	cname := wlCString(name)
	defer wlFree(cname)
	return (*wlCursor)(cptr(xcall(wl.cursorThemeGetCursor, ptr(theme), ptr(cname))))
}

func wlCursorImageGetBuffer(img *wlCursorImage) unsafe.Pointer {
	return cptr(xcall(wl.cursorImageGetBuffer, uintptr(unsafe.Pointer(img))))
}

// wlArgs reads a union wl_argument array.
type wlArgs struct{ p unsafe.Pointer }

// word reads argument i; a union wl_argument is pointer-sized.
func (a wlArgs) word(i int) uintptr {
	return *(*uintptr)(unsafe.Add(a.p, i*int(unsafe.Sizeof(uintptr(0)))))
}
func (a wlArgs) u(i int) uint32         { return uint32(a.word(i)) }
func (a wlArgs) i(i int) int32          { return int32(uint32(a.word(i))) }
func (a wlArgs) o(i int) unsafe.Pointer { return cptr(a.word(i)) }
func (a wlArgs) s(i int) (string, bool) {
	p := a.o(i)
	return goString(p), p != nil
}
func (a wlArgs) str(i int) string { s, _ := a.s(i); return s }

// wlDispatch receives every event of every proxy: implementation is the
// listener kind, args a union wl_argument array.
func wlDispatch(implementation, proxy uintptr, opcode uint32, msg, args uintptr) int32 {
	p, a := cptr(proxy), wlArgs{cptr(args)}
	data := wlProxyGetUserData(p)
	switch implementation - 1 {
	case lk_registry:
		switch opcode {
		case 0:
			wlOnRegistryGlobal(data, p, a.u(0), a.str(1), a.u(2))
		case 1:
			wlOnRegistryGlobalRemove(data, p, a.u(0))
		}
	case lk_surface:
		switch opcode {
		case 0:
			wlOnSurfaceEnter(data, p, a.o(0))
		case 1:
			wlOnSurfaceLeave(data, p, a.o(0))
		}
	case lk_xdg_surface:
		if opcode == 0 {
			wlOnXdgSurfaceConfigure(data, p, a.u(0))
		}
	case lk_xdg_toplevel:
		switch opcode {
		case 0:
			wlOnToplevelConfigure(data, p, a.i(0), a.i(1), a.o(2))
		case 1:
			wlOnToplevelClose(data, p)
		}
	case lk_zxdg_toplevel_decoration_v1:
		if opcode == 0 {
			wlOnToplevelDecorationConfigure(data, p, a.u(0))
		}
	case lk_xdg_wm_base:
		if opcode == 0 { // ping
			xdgWmBasePong(p, a.u(0))
		}
	case lk_callback:
		if opcode == 0 {
			wlOnFrameDone(data, p, a.u(0))
		}
	case lk_output:
		switch opcode {
		case 0:
			wlOnOutputGeometry(data, p, a.i(0), a.i(1), a.i(2), a.i(3), a.i(4), a.str(5), a.str(6), a.i(7))
		case 1:
			wlOnOutputMode(data, p, a.u(0), a.i(1), a.i(2), a.i(3))
		case 2:
			wlOnOutputDone(data, p)
		case 3:
			wlOnOutputScale(data, p, a.i(0))
		}
	case lk_seat:
		switch opcode {
		case 0:
			wlOnSeatCapabilities(data, p, a.u(0))
		case 1:
			wlOnSeatName(data, p, a.str(0))
		}
	case lk_pointer:
		switch opcode {
		case 0:
			wlOnPointerEnter(data, p, a.u(0), a.o(1), a.i(2), a.i(3))
		case 1:
			wlOnPointerLeave(data, p, a.u(0), a.o(1))
		case 2:
			wlOnPointerMotion(data, p, a.u(0), a.i(1), a.i(2))
		case 3:
			wlOnPointerButton(data, p, a.u(0), a.u(1), a.u(2), a.u(3))
		case 4:
			wlOnPointerAxis(data, p, a.u(0), a.u(1), a.i(2))
		case 5:
			wlOnPointerFrame(data, p)
		case 6:
			wlOnPointerAxisSource(data, p, a.u(0))
		case 7:
			wlOnPointerAxisStop(data, p, a.u(0), a.u(1))
		case 8:
			wlOnPointerAxisDiscrete(data, p, a.u(0), a.i(1))
		}
	case lk_touch:
		switch opcode {
		case 0:
			wlOnTouchDown(data, p, a.u(0), a.u(1), a.o(2), a.i(3), a.i(4), a.i(5))
		case 1:
			wlOnTouchUp(data, p, a.u(0), a.u(1), a.i(2))
		case 2:
			wlOnTouchMotion(data, p, a.u(0), a.i(1), a.i(2), a.i(3))
		case 3:
			wlOnTouchFrame(data, p)
		case 4:
			wlOnTouchCancel(data, p)
		}
	case lk_keyboard:
		switch opcode {
		case 0:
			wlOnKeyboardKeymap(data, p, a.u(0), a.i(1), a.u(2))
		case 1:
			wlOnKeyboardEnter(data, p, a.u(0), a.o(1), a.o(2))
		case 2:
			wlOnKeyboardLeave(data, p, a.u(0), a.o(1))
		case 3:
			wlOnKeyboardKey(data, p, a.u(0), a.u(1), a.u(2), a.u(3))
		case 4:
			wlOnKeyboardModifiers(data, p, a.u(0), a.u(1), a.u(2), a.u(3), a.u(4))
		case 5:
			wlOnKeyboardRepeatInfo(data, p, a.i(0), a.i(1))
		}
	case lk_zwp_text_input_v3:
		switch opcode {
		case 0:
			wlOnTextInputEnter(data, p, a.o(0))
		case 1:
			wlOnTextInputLeave(data, p, a.o(0))
		case 2:
			text, set := a.s(0)
			wlOnTextInputPreeditString(data, p, text, set, a.i(1), a.i(2))
		case 3:
			text, set := a.s(0)
			wlOnTextInputCommitString(data, p, text, set)
		case 4:
			wlOnTextInputDeleteSurroundingText(data, p, a.u(0), a.u(1))
		case 5:
			wlOnTextInputDone(data, p, a.u(0))
		}
	case lk_data_device:
		switch opcode {
		case 0:
			wlOnDataDeviceOffer(data, p, a.o(0))
		case 1:
			wlOnDataDeviceEnter(data, p, a.u(0), a.o(1), a.i(2), a.i(3), a.o(4))
		case 2:
			wlOnDataDeviceLeave(data, p)
		case 3:
			wlOnDataDeviceMotion(data, p, a.u(0), a.i(1), a.i(2))
		case 4:
			wlOnDataDeviceDrop(data, p)
		case 5:
			wlOnDataDeviceSelection(data, p, a.o(0))
		}
	case lk_data_offer:
		switch opcode {
		case 0:
			wlOnDataOfferOffer(data, p, a.str(0))
		case 1:
			wlOnDataOfferSourceActions(data, p, a.u(0))
		case 2:
			wlOnDataOfferAction(data, p, a.u(0))
		}
	case lk_data_source:
		switch opcode {
		case 0:
			wlOnDataSourceTarget(data, p, a.str(0))
		case 1:
			wlOnDataSourceSend(data, p, a.str(0), a.i(1))
		case 2:
			wlOnDataSourceCancelled(data, p)
		case 3:
			wlOnDataSourceDNDDropPerformed(data, p)
		case 4:
			wlOnDataSourceDNDFinished(data, p)
		case 5:
			wlOnDataSourceAction(data, p, a.u(0))
		}
	}
	return 0
}
