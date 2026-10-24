// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nox11 && (!cgo || purego)

package app

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The Xlib calls of os_x11.go without cgo: libX11, libX11-xcb, libXcursor,
// libXfixes and libxkbcommon-x11, loaded at run time. Every argument and
// result is an integer or a pointer. Keel patch.

var x11 struct {
	initThreads, rmInitialize, openDisplay, closeDisplay          uintptr
	kbQueryExtension, kbSelectEvents                              uintptr
	defaultRootWindow, defaultScreen, displayWidth, displayHeight uintptr
	createWindow, destroyWindow, mapWindow, mapRaised             uintptr
	iconifyWindow, resizeWindow, moveResizeWindow                 uintptr
	getWindowAttributes, setWMNormalHints, setWMHints             uintptr
	setClassHint, setWMProtocols, storeName                       uintptr
	setTextProperty, getTextProperty, internAtom, deleteProperty  uintptr
	convertSelection, setSelectionOwner, changeProperty           uintptr
	sendEvent, pending, nextEvent, filterEvent, connectionNumber  uintptr
	translateCoordinates, defineCursor, displayString             uintptr
	resourceManagerString, rmGetStringDatabase                    uintptr
	rmGetResource, rmDestroyDatabase                              uintptr
	getXCBConnection, fixesHideCursor, cursorLibraryLoadCursor    uintptr
	kbX11CoreDevice, kbX11KeymapNew, kbX11StateNew, kbKeymapUnref uintptr
}

// openLib opens the first of names that loads.
func openLib(names ...string) (uintptr, error) {
	var err error
	for _, name := range names {
		var h uintptr
		if h, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
			return h, nil
		}
	}
	return 0, err
}

func loadX11() error {
	libs := map[string][]string{
		"x11":     {"libX11.so.6", "libX11.so"},
		"xcb":     {"libX11-xcb.so.1", "libX11-xcb.so"},
		"xcursor": {"libXcursor.so.1", "libXcursor.so"},
		"xfixes":  {"libXfixes.so.3", "libXfixes.so"},
		"xkbx11":  {"libxkbcommon-x11.so.0", "libxkbcommon-x11.so"},
		"xkb":     {"libxkbcommon.so.0", "libxkbcommon.so"},
	}
	handles := map[string]uintptr{}
	for k, names := range libs {
		h, err := openLib(names...)
		if err != nil {
			return fmt.Errorf("x11: %w", err)
		}
		handles[k] = h
	}
	for _, s := range []struct {
		lib, name string
		addr      *uintptr
	}{
		{"x11", "XInitThreads", &x11.initThreads},
		{"x11", "XrmInitialize", &x11.rmInitialize},
		{"x11", "XOpenDisplay", &x11.openDisplay},
		{"x11", "XCloseDisplay", &x11.closeDisplay},
		{"x11", "XkbQueryExtension", &x11.kbQueryExtension},
		{"x11", "XkbSelectEvents", &x11.kbSelectEvents},
		{"x11", "XDefaultRootWindow", &x11.defaultRootWindow},
		{"x11", "XDefaultScreen", &x11.defaultScreen},
		{"x11", "XDisplayWidth", &x11.displayWidth},
		{"x11", "XDisplayHeight", &x11.displayHeight},
		{"x11", "XCreateWindow", &x11.createWindow},
		{"x11", "XDestroyWindow", &x11.destroyWindow},
		{"x11", "XMapWindow", &x11.mapWindow},
		{"x11", "XMapRaised", &x11.mapRaised},
		{"x11", "XIconifyWindow", &x11.iconifyWindow},
		{"x11", "XResizeWindow", &x11.resizeWindow},
		{"x11", "XMoveResizeWindow", &x11.moveResizeWindow},
		{"x11", "XGetWindowAttributes", &x11.getWindowAttributes},
		{"x11", "XSetWMNormalHints", &x11.setWMNormalHints},
		{"x11", "XSetWMHints", &x11.setWMHints},
		{"x11", "XSetClassHint", &x11.setClassHint},
		{"x11", "XSetWMProtocols", &x11.setWMProtocols},
		{"x11", "XStoreName", &x11.storeName},
		{"x11", "XSetTextProperty", &x11.setTextProperty},
		{"x11", "XGetTextProperty", &x11.getTextProperty},
		{"x11", "XInternAtom", &x11.internAtom},
		{"x11", "XDeleteProperty", &x11.deleteProperty},
		{"x11", "XConvertSelection", &x11.convertSelection},
		{"x11", "XSetSelectionOwner", &x11.setSelectionOwner},
		{"x11", "XChangeProperty", &x11.changeProperty},
		{"x11", "XSendEvent", &x11.sendEvent},
		{"x11", "XPending", &x11.pending},
		{"x11", "XNextEvent", &x11.nextEvent},
		{"x11", "XFilterEvent", &x11.filterEvent},
		{"x11", "XConnectionNumber", &x11.connectionNumber},
		{"x11", "XTranslateCoordinates", &x11.translateCoordinates},
		{"x11", "XDefineCursor", &x11.defineCursor},
		{"x11", "XDisplayString", &x11.displayString},
		{"x11", "XResourceManagerString", &x11.resourceManagerString},
		{"x11", "XrmGetStringDatabase", &x11.rmGetStringDatabase},
		{"x11", "XrmGetResource", &x11.rmGetResource},
		{"x11", "XrmDestroyDatabase", &x11.rmDestroyDatabase},
		{"xcb", "XGetXCBConnection", &x11.getXCBConnection},
		{"xfixes", "XFixesHideCursor", &x11.fixesHideCursor},
		{"xcursor", "XcursorLibraryLoadCursor", &x11.cursorLibraryLoadCursor},
		{"xkbx11", "xkb_x11_get_core_keyboard_device_id", &x11.kbX11CoreDevice},
		{"xkbx11", "xkb_x11_keymap_new_from_device", &x11.kbX11KeymapNew},
		{"xkbx11", "xkb_x11_state_new_from_device", &x11.kbX11StateNew},
		{"xkb", "xkb_keymap_unref", &x11.kbKeymapUnref},
	} {
		addr, err := purego.Dlsym(handles[s.lib], s.name)
		if err != nil {
			return fmt.Errorf("x11: %w", err)
		}
		*s.addr = addr
	}
	return nil
}

// xcall calls an X function; pointers to Go memory stay valid and in place
// for the call.
//
//go:uintptrescapes
func xcall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// ptr converts an opaque C pointer, which the garbage collector ignores.
func ptr(p unsafe.Pointer) uintptr { return uintptr(p) }

// cptr returns a pointer result, which points outside the Go heap.
func cptr(r uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&r)) }

// i32 widens a C int result; the register's upper half is undefined.
func i32(r uintptr) int32 { return int32(r) }

// iarg passes a C int argument.
func iarg(v int32) uintptr { return uintptr(v) }

func bytesArg(b []byte) uintptr { return uintptr(unsafe.Pointer(unsafe.SliceData(b))) }

func xInitThreads() int32            { return i32(xcall(x11.initThreads)) }
func xrmInitialize()                 { xcall(x11.rmInitialize) }
func xOpenDisplay() unsafe.Pointer   { return cptr(xcall(x11.openDisplay, 0)) }
func xCloseDisplay(d unsafe.Pointer) { xcall(x11.closeDisplay, ptr(d)) }

func xkbQueryExtension(d unsafe.Pointer, major, minor int32) (int32, bool) {
	maj, min := major, minor
	var base int32
	r := xcall(x11.kbQueryExtension, ptr(d), 0, uintptr(unsafe.Pointer(&base)), 0, uintptr(unsafe.Pointer(&maj)), uintptr(unsafe.Pointer(&min)))
	return base, i32(r) == xTrue
}

func xkbSelectEvents(d unsafe.Pointer, device, affect, values uint32) int32 {
	return i32(xcall(x11.kbSelectEvents, ptr(d), uintptr(device), uintptr(affect), uintptr(values)))
}

func xDefaultRootWindow(d unsafe.Pointer) xWindow { return xcall(x11.defaultRootWindow, ptr(d)) }
func xDefaultScreen(d unsafe.Pointer) int32       { return i32(xcall(x11.defaultScreen, ptr(d))) }

func xDisplayWidth(d unsafe.Pointer, screen int32) int32 {
	return i32(xcall(x11.displayWidth, ptr(d), iarg(screen)))
}

func xDisplayHeight(d unsafe.Pointer, screen int32) int32 {
	return i32(xcall(x11.displayHeight, ptr(d), iarg(screen)))
}

func xCreateWindow(d unsafe.Pointer, parent xWindow, x, y int32, width, height, border uint32, depth int32, class uint32, mask uintptr, attrs *xSetWindowAttributes) xWindow {
	return xcall(x11.createWindow, ptr(d), parent, iarg(x), iarg(y), uintptr(width), uintptr(height), uintptr(border), iarg(depth), uintptr(class), 0, mask, uintptr(unsafe.Pointer(attrs)))
}

func xDestroyWindow(d unsafe.Pointer, w xWindow) { xcall(x11.destroyWindow, ptr(d), w) }
func xMapWindow(d unsafe.Pointer, w xWindow)     { xcall(x11.mapWindow, ptr(d), w) }
func xMapRaised(d unsafe.Pointer, w xWindow)     { xcall(x11.mapRaised, ptr(d), w) }

func xIconifyWindow(d unsafe.Pointer, w xWindow, screen int32) {
	xcall(x11.iconifyWindow, ptr(d), w, iarg(screen))
}

func xResizeWindow(d unsafe.Pointer, w xWindow, width, height uint32) {
	xcall(x11.resizeWindow, ptr(d), w, uintptr(width), uintptr(height))
}

func xMoveResizeWindow(d unsafe.Pointer, w xWindow, x, y int32, width, height uint32) {
	xcall(x11.moveResizeWindow, ptr(d), w, iarg(x), iarg(y), uintptr(width), uintptr(height))
}

func xGetWindowAttributes(d unsafe.Pointer, w xWindow, attrs *xWindowAttributes) {
	xcall(x11.getWindowAttributes, ptr(d), w, uintptr(unsafe.Pointer(attrs)))
}

func xSetWMNormalHints(d unsafe.Pointer, w xWindow, h *xSizeHints) {
	xcall(x11.setWMNormalHints, ptr(d), w, uintptr(unsafe.Pointer(h)))
}

func xSetWMHints(d unsafe.Pointer, w xWindow, h *xWMHints) {
	xcall(x11.setWMHints, ptr(d), w, uintptr(unsafe.Pointer(h)))
}

// xSetClassHint takes NUL-terminated strings; Xlib copies them.
func xSetClassHint(d unsafe.Pointer, w xWindow, name, class []byte) {
	h := &xClassHint{unsafe.Pointer(&name[0]), unsafe.Pointer(&class[0])}
	xcall(x11.setClassHint, ptr(d), w, uintptr(unsafe.Pointer(h)))
}

func xSetWMProtocols(d unsafe.Pointer, w xWindow, atoms []xAtom) {
	xcall(x11.setWMProtocols, ptr(d), w, uintptr(unsafe.Pointer(unsafe.SliceData(atoms))), uintptr(len(atoms)))
}

// xStoreName takes a NUL-terminated name.
func xStoreName(d unsafe.Pointer, w xWindow, name []byte) {
	xcall(x11.storeName, ptr(d), w, bytesArg(name))
}

func xSetTextProperty(d unsafe.Pointer, w xWindow, value []byte, encoding xAtom, format int32, prop xAtom) {
	p := &xTextProperty{
		Value:    unsafe.Pointer(unsafe.SliceData(value)),
		Encoding: encoding,
		Format:   format,
		Nitems:   uintptr(len(value)),
	}
	xcall(x11.setTextProperty, ptr(d), w, uintptr(unsafe.Pointer(p)), prop)
}

func xGetTextProperty(d unsafe.Pointer, w xWindow, p *xTextProperty, prop xAtom) int32 {
	return i32(xcall(x11.getTextProperty, ptr(d), w, uintptr(unsafe.Pointer(p)), prop))
}

// xInternAtom takes a NUL-terminated name.
func xInternAtom(d unsafe.Pointer, name []byte, onlyIfExists int32) xAtom {
	return xcall(x11.internAtom, ptr(d), bytesArg(name), iarg(onlyIfExists))
}

func xDeleteProperty(d unsafe.Pointer, w xWindow, prop xAtom) {
	xcall(x11.deleteProperty, ptr(d), w, prop)
}

func xConvertSelection(d unsafe.Pointer, sel, target, prop xAtom, requestor xWindow, time uintptr) {
	xcall(x11.convertSelection, ptr(d), sel, target, prop, requestor, time)
}

func xSetSelectionOwner(d unsafe.Pointer, sel xAtom, owner xWindow, time uintptr) {
	xcall(x11.setSelectionOwner, ptr(d), sel, owner, time)
}

func xChangeProperty(d unsafe.Pointer, w xWindow, prop, typ xAtom, format, mode int32, data unsafe.Pointer, n int32) {
	xcall(x11.changeProperty, ptr(d), w, prop, typ, iarg(format), iarg(mode), uintptr(data), iarg(n))
}

func xSendEvent(d unsafe.Pointer, w xWindow, propagate int32, mask int, ev *xEvent) {
	xcall(x11.sendEvent, ptr(d), w, iarg(propagate), uintptr(mask), uintptr(unsafe.Pointer(ev)))
}

func xPending(d unsafe.Pointer) int32 { return i32(xcall(x11.pending, ptr(d))) }

func xNextEvent(d unsafe.Pointer, ev *xEvent) {
	xcall(x11.nextEvent, ptr(d), uintptr(unsafe.Pointer(ev)))
}

func xFilterEvent(ev *xEvent, w xWindow) int32 {
	return i32(xcall(x11.filterEvent, uintptr(unsafe.Pointer(ev)), w))
}

func xConnectionNumber(d unsafe.Pointer) int32 { return i32(xcall(x11.connectionNumber, ptr(d))) }

func xTranslateCoordinates(d unsafe.Pointer, src, dst xWindow, x, y int32) (int32, int32) {
	var dx, dy int32
	var child xWindow
	xcall(x11.translateCoordinates, ptr(d), src, dst, iarg(x), iarg(y), uintptr(unsafe.Pointer(&dx)), uintptr(unsafe.Pointer(&dy)), uintptr(unsafe.Pointer(&child)))
	return dx, dy
}

func xFixesHideCursor(d unsafe.Pointer, w xWindow) { xcall(x11.fixesHideCursor, ptr(d), w) }

// xcursorLibraryLoadCursor takes a NUL-terminated name.
func xcursorLibraryLoadCursor(d unsafe.Pointer, name []byte) uintptr {
	return xcall(x11.cursorLibraryLoadCursor, ptr(d), bytesArg(name))
}

func xDefineCursor(d unsafe.Pointer, w xWindow, c uintptr) { xcall(x11.defineCursor, ptr(d), w, c) }

func xDisplayString(d unsafe.Pointer) string { return goString(cptr(xcall(x11.displayString, ptr(d)))) }

func xResourceManagerString(d unsafe.Pointer) unsafe.Pointer {
	return cptr(xcall(x11.resourceManagerString, ptr(d)))
}

func xrmGetStringDatabase(s unsafe.Pointer) unsafe.Pointer {
	return cptr(xcall(x11.rmGetStringDatabase, ptr(s)))
}

func xrmGetResource(db unsafe.Pointer, name, class string) (string, xrmValue, bool) {
	var t unsafe.Pointer
	var v xrmValue
	cname, cclass := cString(name), cString(class)
	r := xcall(x11.rmGetResource, ptr(db), bytesArg(cname), bytesArg(cclass), uintptr(unsafe.Pointer(&t)), uintptr(unsafe.Pointer(&v)))
	if i32(r) == xFalse || t == nil {
		return "", xrmValue{}, false
	}
	return goString(t), v, true
}

func xrmDestroyDatabase(db unsafe.Pointer) { xcall(x11.rmDestroyDatabase, ptr(db)) }

func xGetXCBConnection(d unsafe.Pointer) unsafe.Pointer {
	return cptr(xcall(x11.getXCBConnection, ptr(d)))
}

func xkbX11GetCoreKeyboardDeviceID(xcb unsafe.Pointer) int32 {
	return i32(xcall(x11.kbX11CoreDevice, ptr(xcb)))
}

func xkbX11KeymapNewFromDevice(ctx, xcb unsafe.Pointer, dev int32) unsafe.Pointer {
	return cptr(xcall(x11.kbX11KeymapNew, ptr(ctx), ptr(xcb), iarg(dev), 0)) // XKB_KEYMAP_COMPILE_NO_FLAGS
}

func xkbX11StateNewFromDevice(km, xcb unsafe.Pointer, dev int32) unsafe.Pointer {
	return cptr(xcall(x11.kbX11StateNew, ptr(km), ptr(xcb), iarg(dev)))
}

func xkbKeymapUnref(km unsafe.Pointer) { xcall(x11.kbKeymapUnref, ptr(km)) }
