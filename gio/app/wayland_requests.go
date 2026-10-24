// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import (
	"structs"
	"unsafe"
)

// The Wayland requests os_wayland.go sends, marshalled with
// wl_proxy_marshal_array_flags as wayland-scanner's inline functions
// marshal them with wl_proxy_marshal_flags. Keel patch: both bindings use
// these; os_wayland_cgo.go checks the opcodes and values below against the
// headers.

// Request opcodes.
const (
	wlDisplayGetRegistryOp = 1

	wlRegistryBindOp = 0

	wlCompositorCreateSurfaceOp = 0
	wlCompositorCreateRegionOp  = 1

	wlSurfaceDestroyOp          = 0
	wlSurfaceAttachOp           = 1
	wlSurfaceDamageOp           = 2
	wlSurfaceFrameOp            = 3
	wlSurfaceSetOpaqueRegionOp  = 4
	wlSurfaceCommitOp           = 6
	wlSurfaceSetBufferScaleOp   = 8
	wlRegionDestroyOp           = 0
	wlRegionAddOp               = 1
	wlSeatGetPointerOp          = 0
	wlSeatGetKeyboardOp         = 1
	wlSeatGetTouchOp            = 2
	wlSeatReleaseOp             = 3
	wlPointerSetCursorOp        = 0
	wlPointerReleaseOp          = 1
	wlKeyboardReleaseOp         = 0
	wlTouchReleaseOp            = 0
	wlDataDeviceManagerCreateOp = 0
	wlDataDeviceManagerGetOp    = 1
	wlDataDeviceSetSelectionOp  = 1
	wlDataDeviceReleaseOp       = 2
	wlDataSourceOfferOp         = 0
	wlDataSourceDestroyOp       = 1
	wlDataOfferReceiveOp        = 1
	wlDataOfferDestroyOp        = 2

	xdgWmBaseDestroyOp       = 0
	xdgWmBaseGetXdgSurfaceOp = 2
	xdgWmBasePongOp          = 3

	xdgSurfaceDestroyOp      = 0
	xdgSurfaceGetToplevelOp  = 1
	xdgSurfaceAckConfigureOp = 4

	xdgToplevelDestroyOp         = 0
	xdgToplevelSetTitleOp        = 2
	xdgToplevelSetAppIdOp        = 3
	xdgToplevelMoveOp            = 5
	xdgToplevelResizeOp          = 6
	xdgToplevelSetMaxSizeOp      = 7
	xdgToplevelSetMinSizeOp      = 8
	xdgToplevelSetMaximizedOp    = 9
	xdgToplevelUnsetMaximizedOp  = 10
	xdgToplevelSetFullscreenOp   = 11
	xdgToplevelUnsetFullscreenOp = 12
	xdgToplevelSetMinimizedOp    = 13

	zxdgDecorationManagerDestroyOp = 0
	zxdgDecorationManagerGetOp     = 1
	zxdgToplevelDecorationDestroy  = 0

	zwpTextInputV3DestroyOp             = 0
	zwpTextInputV3EnableOp              = 1
	zwpTextInputV3DisableOp             = 2
	zwpTextInputV3SetSurroundingTextOp  = 3
	zwpTextInputV3SetTextChangeCauseOp  = 4
	zwpTextInputV3SetContentTypeOp      = 5
	zwpTextInputV3SetCursorRectangleOp  = 6
	zwpTextInputV3CommitOp              = 7
	zwpTextInputManagerV3DestroyOp      = 0
	zwpTextInputManagerV3GetTextInputOp = 1
	wlMarshalFlagDestroy                = 1 << 0
)

// Protocol values.
const (
	_WL_SEAT_CAPABILITY_POINTER        = 1
	_WL_SEAT_CAPABILITY_KEYBOARD       = 2
	_WL_SEAT_CAPABILITY_TOUCH          = 4
	_WL_POINTER_AXIS_VERTICAL_SCROLL   = 0
	_WL_POINTER_AXIS_HORIZONTAL_SCROLL = 1
	_WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1  = 1
	_WL_KEYBOARD_KEY_STATE_RELEASED    = 0
	_WL_KEYBOARD_KEY_STATE_PRESSED     = 1
	_WL_OUTPUT_MODE_CURRENT            = 1

	_XDG_TOPLEVEL_RESIZE_EDGE_TOP          = 1
	_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM       = 2
	_XDG_TOPLEVEL_RESIZE_EDGE_LEFT         = 4
	_XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT     = 5
	_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT  = 6
	_XDG_TOPLEVEL_RESIZE_EDGE_RIGHT        = 8
	_XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT    = 9
	_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT = 10

	_ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE = 1
	_ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE = 2

	_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_INPUT_METHOD   = 0
	_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER          = 1
	_ZWP_TEXT_INPUT_V3_CONTENT_HINT_COMPLETION     = 0x1
	_ZWP_TEXT_INPUT_V3_CONTENT_HINT_SPELLCHECK     = 0x2
	_ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT    = 0x40
	_ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA = 0x80
	_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NUMBER      = 4
	_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PHONE       = 5
	_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_URL         = 6
	_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_EMAIL       = 7
	_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD    = 9
)

// Listener kinds: each binding maps a kind to its listener table or to its
// event dispatcher.
const (
	lk_registry = iota
	lk_surface
	lk_xdg_surface
	lk_xdg_toplevel
	lk_zxdg_toplevel_decoration_v1
	lk_xdg_wm_base
	lk_callback
	lk_output
	lk_seat
	lk_pointer
	lk_touch
	lk_keyboard
	lk_zwp_text_input_v3
	lk_data_device
	lk_data_offer
	lk_data_source
)

// wlIface holds the wl_interface pointers the bindings provide.
var wlIface struct {
	wl_registry, wl_compositor, wl_surface, wl_region, wl_callback unsafe.Pointer
	wl_seat, wl_pointer, wl_keyboard, wl_touch, wl_output, wl_shm  unsafe.Pointer
	wl_data_device_manager, wl_data_device, wl_data_source         unsafe.Pointer
	xdg_wm_base, xdg_surface, xdg_toplevel                         unsafe.Pointer
	zxdg_decoration_manager_v1, zxdg_toplevel_decoration_v1        unsafe.Pointer
	zwp_text_input_manager_v3, zwp_text_input_v3                   unsafe.Pointer
}

// wlCursor and wlCursorImage are libwayland-cursor's struct wl_cursor and
// struct wl_cursor_image, in C memory.
type wlCursor struct {
	_          structs.HostLayout
	ImageCount uint32
	Images     **wlCursorImage
	Name       unsafe.Pointer
}

type wlCursorImage struct {
	_                  structs.HostLayout
	Width, Height      uint32
	HotspotX, HotspotY uint32
	Delay              uint32
}

// wlArg packs a wl_argument: integers in the low 32 bits, pointers whole.
func wlArg(v uint32) uintptr { return uintptr(v) }

// marshal sends a request; iface is the interface of a new object, if the
// request creates one.
func marshal(proxy unsafe.Pointer, op uint32, iface unsafe.Pointer, flags uint32, args ...uintptr) unsafe.Pointer {
	if len(args) == 0 {
		args = []uintptr{0}
	}
	return wlProxyMarshalArrayFlags(proxy, op, iface, wlProxyGetVersion(proxy), flags, args)
}

// marshalString sends a request whose argument at index i is s, copied to
// C memory for the call.
func marshalString(proxy unsafe.Pointer, op uint32, flags uint32, i int, s string, args ...uintptr) {
	cs := wlCString(s)
	defer wlFree(cs)
	args[i] = uintptr(cs)
	marshal(proxy, op, nil, flags, args...)
}

func wlDisplayGetRegistry(d unsafe.Pointer) unsafe.Pointer {
	return marshal(d, wlDisplayGetRegistryOp, wlIface.wl_registry, 0, 0)
}

// wlRegistryBind binds a global; the request's arguments are the global's
// name, the interface name and version, and the new object.
func wlRegistryBind(reg unsafe.Pointer, name uint32, iface unsafe.Pointer, ifaceName string, version uint32) unsafe.Pointer {
	cs := wlCString(ifaceName)
	defer wlFree(cs)
	args := []uintptr{wlArg(name), uintptr(cs), wlArg(version), 0}
	return wlProxyMarshalArrayFlags(reg, wlRegistryBindOp, iface, version, 0, args)
}

func wlRegistryDestroy(reg unsafe.Pointer) { wlProxyDestroy(reg) }

func wlCompositorCreateSurface(c unsafe.Pointer) unsafe.Pointer {
	return marshal(c, wlCompositorCreateSurfaceOp, wlIface.wl_surface, 0, 0)
}

func wlCompositorCreateRegion(c unsafe.Pointer) unsafe.Pointer {
	return marshal(c, wlCompositorCreateRegionOp, wlIface.wl_region, 0, 0)
}

func wlCompositorDestroy(c unsafe.Pointer) { wlProxyDestroy(c) }
func wlShmDestroy(s unsafe.Pointer)        { wlProxyDestroy(s) }
func wlOutputDestroy(o unsafe.Pointer)     { wlProxyDestroy(o) }
func wlCallbackDestroy(c unsafe.Pointer)   { wlProxyDestroy(c) }

func wlSurfaceDestroy(s unsafe.Pointer) {
	marshal(s, wlSurfaceDestroyOp, nil, wlMarshalFlagDestroy)
}

func wlSurfaceAttach(s, buf unsafe.Pointer, x, y int32) {
	marshal(s, wlSurfaceAttachOp, nil, 0, uintptr(buf), wlArg(uint32(x)), wlArg(uint32(y)))
}

func wlSurfaceDamage(s unsafe.Pointer, x, y, w, h int32) {
	marshal(s, wlSurfaceDamageOp, nil, 0, wlArg(uint32(x)), wlArg(uint32(y)), wlArg(uint32(w)), wlArg(uint32(h)))
}

func wlSurfaceFrame(s unsafe.Pointer) unsafe.Pointer {
	return marshal(s, wlSurfaceFrameOp, wlIface.wl_callback, 0, 0)
}

func wlSurfaceSetOpaqueRegion(s, reg unsafe.Pointer) {
	marshal(s, wlSurfaceSetOpaqueRegionOp, nil, 0, uintptr(reg))
}

func wlSurfaceCommit(s unsafe.Pointer) { marshal(s, wlSurfaceCommitOp, nil, 0) }

func wlSurfaceSetBufferScale(s unsafe.Pointer, scale int32) {
	marshal(s, wlSurfaceSetBufferScaleOp, nil, 0, wlArg(uint32(scale)))
}

func wlRegionDestroy(r unsafe.Pointer) { marshal(r, wlRegionDestroyOp, nil, wlMarshalFlagDestroy) }

func wlRegionAdd(r unsafe.Pointer, x, y, w, h int32) {
	marshal(r, wlRegionAddOp, nil, 0, wlArg(uint32(x)), wlArg(uint32(y)), wlArg(uint32(w)), wlArg(uint32(h)))
}

func wlSeatGetPointer(s unsafe.Pointer) unsafe.Pointer {
	return marshal(s, wlSeatGetPointerOp, wlIface.wl_pointer, 0, 0)
}

func wlSeatGetKeyboard(s unsafe.Pointer) unsafe.Pointer {
	return marshal(s, wlSeatGetKeyboardOp, wlIface.wl_keyboard, 0, 0)
}

func wlSeatGetTouch(s unsafe.Pointer) unsafe.Pointer {
	return marshal(s, wlSeatGetTouchOp, wlIface.wl_touch, 0, 0)
}

func wlSeatRelease(s unsafe.Pointer) { marshal(s, wlSeatReleaseOp, nil, wlMarshalFlagDestroy) }

func wlPointerSetCursor(p unsafe.Pointer, serial uint32, surf unsafe.Pointer, hx, hy int32) {
	marshal(p, wlPointerSetCursorOp, nil, 0, wlArg(serial), uintptr(surf), wlArg(uint32(hx)), wlArg(uint32(hy)))
}

func wlPointerRelease(p unsafe.Pointer)  { marshal(p, wlPointerReleaseOp, nil, wlMarshalFlagDestroy) }
func wlKeyboardRelease(k unsafe.Pointer) { marshal(k, wlKeyboardReleaseOp, nil, wlMarshalFlagDestroy) }
func wlTouchRelease(t unsafe.Pointer)    { marshal(t, wlTouchReleaseOp, nil, wlMarshalFlagDestroy) }

func wlDataDeviceManagerCreateDataSource(m unsafe.Pointer) unsafe.Pointer {
	return marshal(m, wlDataDeviceManagerCreateOp, wlIface.wl_data_source, 0, 0)
}

func wlDataDeviceManagerGetDataDevice(m, seat unsafe.Pointer) unsafe.Pointer {
	return marshal(m, wlDataDeviceManagerGetOp, wlIface.wl_data_device, 0, 0, uintptr(seat))
}

func wlDataDeviceSetSelection(d, source unsafe.Pointer, serial uint32) {
	marshal(d, wlDataDeviceSetSelectionOp, nil, 0, uintptr(source), wlArg(serial))
}

func wlDataDeviceRelease(d unsafe.Pointer) {
	marshal(d, wlDataDeviceReleaseOp, nil, wlMarshalFlagDestroy)
}

func wlDataSourceOffer(s unsafe.Pointer, mime string) {
	marshalString(s, wlDataSourceOfferOp, 0, 0, mime, 0)
}

func wlDataSourceDestroy(s unsafe.Pointer) {
	marshal(s, wlDataSourceDestroyOp, nil, wlMarshalFlagDestroy)
}

func wlDataOfferReceive(o unsafe.Pointer, mime string, fd int32) {
	marshalString(o, wlDataOfferReceiveOp, 0, 0, mime, 0, wlArg(uint32(fd)))
}

func wlDataOfferDestroy(o unsafe.Pointer) {
	marshal(o, wlDataOfferDestroyOp, nil, wlMarshalFlagDestroy)
}

func xdgWmBaseDestroy(wm unsafe.Pointer) {
	marshal(wm, xdgWmBaseDestroyOp, nil, wlMarshalFlagDestroy)
}

func xdgWmBaseGetXdgSurface(wm, surf unsafe.Pointer) unsafe.Pointer {
	return marshal(wm, xdgWmBaseGetXdgSurfaceOp, wlIface.xdg_surface, 0, 0, uintptr(surf))
}

func xdgWmBasePong(wm unsafe.Pointer, serial uint32) {
	marshal(wm, xdgWmBasePongOp, nil, 0, wlArg(serial))
}

func xdgSurfaceDestroy(s unsafe.Pointer) {
	marshal(s, xdgSurfaceDestroyOp, nil, wlMarshalFlagDestroy)
}

func xdgSurfaceGetToplevel(s unsafe.Pointer) unsafe.Pointer {
	return marshal(s, xdgSurfaceGetToplevelOp, wlIface.xdg_toplevel, 0, 0)
}

func xdgSurfaceAckConfigure(s unsafe.Pointer, serial uint32) {
	marshal(s, xdgSurfaceAckConfigureOp, nil, 0, wlArg(serial))
}

func xdgToplevelDestroy(t unsafe.Pointer) {
	marshal(t, xdgToplevelDestroyOp, nil, wlMarshalFlagDestroy)
}

func xdgToplevelSetTitle(t unsafe.Pointer, title string) {
	marshalString(t, xdgToplevelSetTitleOp, 0, 0, title, 0)
}

func xdgToplevelSetAppId(t unsafe.Pointer, id string) {
	marshalString(t, xdgToplevelSetAppIdOp, 0, 0, id, 0)
}

func xdgToplevelMove(t, seat unsafe.Pointer, serial uint32) {
	marshal(t, xdgToplevelMoveOp, nil, 0, uintptr(seat), wlArg(serial))
}

func xdgToplevelResize(t, seat unsafe.Pointer, serial, edges uint32) {
	marshal(t, xdgToplevelResizeOp, nil, 0, uintptr(seat), wlArg(serial), wlArg(edges))
}

func xdgToplevelSetMaxSize(t unsafe.Pointer, w, h int32) {
	marshal(t, xdgToplevelSetMaxSizeOp, nil, 0, wlArg(uint32(w)), wlArg(uint32(h)))
}

func xdgToplevelSetMinSize(t unsafe.Pointer, w, h int32) {
	marshal(t, xdgToplevelSetMinSizeOp, nil, 0, wlArg(uint32(w)), wlArg(uint32(h)))
}

func xdgToplevelSetMaximized(t unsafe.Pointer) { marshal(t, xdgToplevelSetMaximizedOp, nil, 0) }

func xdgToplevelUnsetMaximized(t unsafe.Pointer) {
	marshal(t, xdgToplevelUnsetMaximizedOp, nil, 0)
}

func xdgToplevelSetFullscreen(t, output unsafe.Pointer) {
	marshal(t, xdgToplevelSetFullscreenOp, nil, 0, uintptr(output))
}

func xdgToplevelUnsetFullscreen(t unsafe.Pointer) {
	marshal(t, xdgToplevelUnsetFullscreenOp, nil, 0)
}

func xdgToplevelSetMinimized(t unsafe.Pointer) { marshal(t, xdgToplevelSetMinimizedOp, nil, 0) }

func zxdgDecorationManagerV1Destroy(m unsafe.Pointer) {
	marshal(m, zxdgDecorationManagerDestroyOp, nil, wlMarshalFlagDestroy)
}

func zxdgDecorationManagerV1GetToplevelDecoration(m, toplevel unsafe.Pointer) unsafe.Pointer {
	return marshal(m, zxdgDecorationManagerGetOp, wlIface.zxdg_toplevel_decoration_v1, 0, 0, uintptr(toplevel))
}

func zxdgToplevelDecorationV1Destroy(d unsafe.Pointer) {
	marshal(d, zxdgToplevelDecorationDestroy, nil, wlMarshalFlagDestroy)
}

func zwpTextInputManagerV3Destroy(m unsafe.Pointer) {
	marshal(m, zwpTextInputManagerV3DestroyOp, nil, wlMarshalFlagDestroy)
}

func zwpTextInputManagerV3GetTextInput(m, seat unsafe.Pointer) unsafe.Pointer {
	return marshal(m, zwpTextInputManagerV3GetTextInputOp, wlIface.zwp_text_input_v3, 0, 0, uintptr(seat))
}

func zwpTextInputV3Destroy(t unsafe.Pointer) {
	marshal(t, zwpTextInputV3DestroyOp, nil, wlMarshalFlagDestroy)
}

func zwpTextInputV3Enable(t unsafe.Pointer)  { marshal(t, zwpTextInputV3EnableOp, nil, 0) }
func zwpTextInputV3Disable(t unsafe.Pointer) { marshal(t, zwpTextInputV3DisableOp, nil, 0) }
func zwpTextInputV3Commit(t unsafe.Pointer)  { marshal(t, zwpTextInputV3CommitOp, nil, 0) }

func zwpTextInputV3SetSurroundingText(t unsafe.Pointer, text string, cursor, anchor int32) {
	marshalString(t, zwpTextInputV3SetSurroundingTextOp, 0, 0, text, 0, wlArg(uint32(cursor)), wlArg(uint32(anchor)))
}

func zwpTextInputV3SetTextChangeCause(t unsafe.Pointer, cause uint32) {
	marshal(t, zwpTextInputV3SetTextChangeCauseOp, nil, 0, wlArg(cause))
}

func zwpTextInputV3SetContentType(t unsafe.Pointer, hint, purpose uint32) {
	marshal(t, zwpTextInputV3SetContentTypeOp, nil, 0, wlArg(hint), wlArg(purpose))
}

func zwpTextInputV3SetCursorRectangle(t unsafe.Pointer, x, y, w, h int32) {
	marshal(t, zwpTextInputV3SetCursorRectangleOp, nil, 0, wlArg(uint32(x)), wlArg(uint32(y)), wlArg(uint32(w)), wlArg(uint32(h)))
}
