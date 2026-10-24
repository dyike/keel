// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android && cgo && !purego) || freebsd) && !nowayland

package app

// libwayland through cgo, behind the functions os_wayland.go uses: Gio's
// listener tables in os_wayland.c and the scanner-generated protocol code
// deliver events through the //export shims below to the wlOn* handlers.
// os_wayland_purego.go does the same without cgo. Keel patch: moved from
// os_wayland.go.

/*
#cgo linux pkg-config: wayland-client wayland-cursor
#cgo freebsd openbsd LDFLAGS: -lwayland-client -lwayland-cursor
#cgo freebsd CFLAGS: -I/usr/local/include
#cgo freebsd LDFLAGS: -L/usr/local/lib

#include <stdlib.h>
#include <wayland-client.h>
#include <wayland-cursor.h>
#include "wayland_text_input.h"
#include "wayland_xdg_shell.h"
#include "wayland_xdg_decoration.h"

extern const struct wl_registry_listener gio_registry_listener;
extern const struct wl_surface_listener gio_surface_listener;
extern const struct xdg_surface_listener gio_xdg_surface_listener;
extern const struct xdg_toplevel_listener gio_xdg_toplevel_listener;
extern const struct zxdg_toplevel_decoration_v1_listener gio_zxdg_toplevel_decoration_v1_listener;
extern const struct xdg_wm_base_listener gio_xdg_wm_base_listener;
extern const struct wl_callback_listener gio_callback_listener;
extern const struct wl_output_listener gio_output_listener;
extern const struct wl_seat_listener gio_seat_listener;
extern const struct wl_pointer_listener gio_pointer_listener;
extern const struct wl_touch_listener gio_touch_listener;
extern const struct wl_keyboard_listener gio_keyboard_listener;
extern const struct zwp_text_input_v3_listener gio_zwp_text_input_v3_listener;
extern const struct wl_data_device_listener gio_data_device_listener;
extern const struct wl_data_offer_listener gio_data_offer_listener;
extern const struct wl_data_source_listener gio_data_source_listener;
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func init() {
	wlIface.wl_registry = unsafe.Pointer(&C.wl_registry_interface)
	wlIface.wl_compositor = unsafe.Pointer(&C.wl_compositor_interface)
	wlIface.wl_surface = unsafe.Pointer(&C.wl_surface_interface)
	wlIface.wl_region = unsafe.Pointer(&C.wl_region_interface)
	wlIface.wl_callback = unsafe.Pointer(&C.wl_callback_interface)
	wlIface.wl_seat = unsafe.Pointer(&C.wl_seat_interface)
	wlIface.wl_pointer = unsafe.Pointer(&C.wl_pointer_interface)
	wlIface.wl_keyboard = unsafe.Pointer(&C.wl_keyboard_interface)
	wlIface.wl_touch = unsafe.Pointer(&C.wl_touch_interface)
	wlIface.wl_output = unsafe.Pointer(&C.wl_output_interface)
	wlIface.wl_shm = unsafe.Pointer(&C.wl_shm_interface)
	wlIface.wl_data_device_manager = unsafe.Pointer(&C.wl_data_device_manager_interface)
	wlIface.wl_data_device = unsafe.Pointer(&C.wl_data_device_interface)
	wlIface.wl_data_source = unsafe.Pointer(&C.wl_data_source_interface)
	wlIface.xdg_wm_base = unsafe.Pointer(&C.xdg_wm_base_interface)
	wlIface.xdg_surface = unsafe.Pointer(&C.xdg_surface_interface)
	wlIface.xdg_toplevel = unsafe.Pointer(&C.xdg_toplevel_interface)
	wlIface.zxdg_decoration_manager_v1 = unsafe.Pointer(&C.zxdg_decoration_manager_v1_interface)
	wlIface.zxdg_toplevel_decoration_v1 = unsafe.Pointer(&C.zxdg_toplevel_decoration_v1_interface)
	wlIface.zwp_text_input_manager_v3 = unsafe.Pointer(&C.zwp_text_input_manager_v3_interface)
	wlIface.zwp_text_input_v3 = unsafe.Pointer(&C.zwp_text_input_v3_interface)

	// The opcodes and values wayland_requests.go spells out, and the cursor
	// structures, must match the headers.
	for _, c := range [...]struct {
		name      string
		want, got int64
	}{
		{"wlDisplayGetRegistryOp", C.WL_DISPLAY_GET_REGISTRY, wlDisplayGetRegistryOp},
		{"wlRegistryBindOp", C.WL_REGISTRY_BIND, wlRegistryBindOp},
		{"wlCompositorCreateSurfaceOp", C.WL_COMPOSITOR_CREATE_SURFACE, wlCompositorCreateSurfaceOp},
		{"wlCompositorCreateRegionOp", C.WL_COMPOSITOR_CREATE_REGION, wlCompositorCreateRegionOp},
		{"wlSurfaceDestroyOp", C.WL_SURFACE_DESTROY, wlSurfaceDestroyOp},
		{"wlSurfaceAttachOp", C.WL_SURFACE_ATTACH, wlSurfaceAttachOp},
		{"wlSurfaceDamageOp", C.WL_SURFACE_DAMAGE, wlSurfaceDamageOp},
		{"wlSurfaceFrameOp", C.WL_SURFACE_FRAME, wlSurfaceFrameOp},
		{"wlSurfaceSetOpaqueRegionOp", C.WL_SURFACE_SET_OPAQUE_REGION, wlSurfaceSetOpaqueRegionOp},
		{"wlSurfaceCommitOp", C.WL_SURFACE_COMMIT, wlSurfaceCommitOp},
		{"wlSurfaceSetBufferScaleOp", C.WL_SURFACE_SET_BUFFER_SCALE, wlSurfaceSetBufferScaleOp},
		{"wlRegionDestroyOp", C.WL_REGION_DESTROY, wlRegionDestroyOp},
		{"wlRegionAddOp", C.WL_REGION_ADD, wlRegionAddOp},
		{"wlSeatGetPointerOp", C.WL_SEAT_GET_POINTER, wlSeatGetPointerOp},
		{"wlSeatGetKeyboardOp", C.WL_SEAT_GET_KEYBOARD, wlSeatGetKeyboardOp},
		{"wlSeatGetTouchOp", C.WL_SEAT_GET_TOUCH, wlSeatGetTouchOp},
		{"wlSeatReleaseOp", C.WL_SEAT_RELEASE, wlSeatReleaseOp},
		{"wlPointerSetCursorOp", C.WL_POINTER_SET_CURSOR, wlPointerSetCursorOp},
		{"wlPointerReleaseOp", C.WL_POINTER_RELEASE, wlPointerReleaseOp},
		{"wlKeyboardReleaseOp", C.WL_KEYBOARD_RELEASE, wlKeyboardReleaseOp},
		{"wlTouchReleaseOp", C.WL_TOUCH_RELEASE, wlTouchReleaseOp},
		{"wlDataDeviceManagerCreateOp", C.WL_DATA_DEVICE_MANAGER_CREATE_DATA_SOURCE, wlDataDeviceManagerCreateOp},
		{"wlDataDeviceManagerGetOp", C.WL_DATA_DEVICE_MANAGER_GET_DATA_DEVICE, wlDataDeviceManagerGetOp},
		{"wlDataDeviceSetSelectionOp", C.WL_DATA_DEVICE_SET_SELECTION, wlDataDeviceSetSelectionOp},
		{"wlDataDeviceReleaseOp", C.WL_DATA_DEVICE_RELEASE, wlDataDeviceReleaseOp},
		{"wlDataSourceOfferOp", C.WL_DATA_SOURCE_OFFER, wlDataSourceOfferOp},
		{"wlDataSourceDestroyOp", C.WL_DATA_SOURCE_DESTROY, wlDataSourceDestroyOp},
		{"wlDataOfferReceiveOp", C.WL_DATA_OFFER_RECEIVE, wlDataOfferReceiveOp},
		{"wlDataOfferDestroyOp", C.WL_DATA_OFFER_DESTROY, wlDataOfferDestroyOp},
		{"xdgWmBaseDestroyOp", C.XDG_WM_BASE_DESTROY, xdgWmBaseDestroyOp},
		{"xdgWmBaseGetXdgSurfaceOp", C.XDG_WM_BASE_GET_XDG_SURFACE, xdgWmBaseGetXdgSurfaceOp},
		{"xdgWmBasePongOp", C.XDG_WM_BASE_PONG, xdgWmBasePongOp},
		{"xdgSurfaceDestroyOp", C.XDG_SURFACE_DESTROY, xdgSurfaceDestroyOp},
		{"xdgSurfaceGetToplevelOp", C.XDG_SURFACE_GET_TOPLEVEL, xdgSurfaceGetToplevelOp},
		{"xdgSurfaceAckConfigureOp", C.XDG_SURFACE_ACK_CONFIGURE, xdgSurfaceAckConfigureOp},
		{"xdgToplevelDestroyOp", C.XDG_TOPLEVEL_DESTROY, xdgToplevelDestroyOp},
		{"xdgToplevelSetTitleOp", C.XDG_TOPLEVEL_SET_TITLE, xdgToplevelSetTitleOp},
		{"xdgToplevelSetAppIdOp", C.XDG_TOPLEVEL_SET_APP_ID, xdgToplevelSetAppIdOp},
		{"xdgToplevelMoveOp", C.XDG_TOPLEVEL_MOVE, xdgToplevelMoveOp},
		{"xdgToplevelResizeOp", C.XDG_TOPLEVEL_RESIZE, xdgToplevelResizeOp},
		{"xdgToplevelSetMaxSizeOp", C.XDG_TOPLEVEL_SET_MAX_SIZE, xdgToplevelSetMaxSizeOp},
		{"xdgToplevelSetMinSizeOp", C.XDG_TOPLEVEL_SET_MIN_SIZE, xdgToplevelSetMinSizeOp},
		{"xdgToplevelSetMaximizedOp", C.XDG_TOPLEVEL_SET_MAXIMIZED, xdgToplevelSetMaximizedOp},
		{"xdgToplevelUnsetMaximizedOp", C.XDG_TOPLEVEL_UNSET_MAXIMIZED, xdgToplevelUnsetMaximizedOp},
		{"xdgToplevelSetFullscreenOp", C.XDG_TOPLEVEL_SET_FULLSCREEN, xdgToplevelSetFullscreenOp},
		{"xdgToplevelUnsetFullscreenOp", C.XDG_TOPLEVEL_UNSET_FULLSCREEN, xdgToplevelUnsetFullscreenOp},
		{"xdgToplevelSetMinimizedOp", C.XDG_TOPLEVEL_SET_MINIMIZED, xdgToplevelSetMinimizedOp},
		{"zxdgDecorationManagerDestroyOp", C.ZXDG_DECORATION_MANAGER_V1_DESTROY, zxdgDecorationManagerDestroyOp},
		{"zxdgDecorationManagerGetOp", C.ZXDG_DECORATION_MANAGER_V1_GET_TOPLEVEL_DECORATION, zxdgDecorationManagerGetOp},
		{"zxdgToplevelDecorationDestroy", C.ZXDG_TOPLEVEL_DECORATION_V1_DESTROY, zxdgToplevelDecorationDestroy},
		{"zwpTextInputV3DestroyOp", C.ZWP_TEXT_INPUT_V3_DESTROY, zwpTextInputV3DestroyOp},
		{"zwpTextInputV3EnableOp", C.ZWP_TEXT_INPUT_V3_ENABLE, zwpTextInputV3EnableOp},
		{"zwpTextInputV3DisableOp", C.ZWP_TEXT_INPUT_V3_DISABLE, zwpTextInputV3DisableOp},
		{"zwpTextInputV3SetSurroundingTextOp", C.ZWP_TEXT_INPUT_V3_SET_SURROUNDING_TEXT, zwpTextInputV3SetSurroundingTextOp},
		{"zwpTextInputV3SetTextChangeCauseOp", C.ZWP_TEXT_INPUT_V3_SET_TEXT_CHANGE_CAUSE, zwpTextInputV3SetTextChangeCauseOp},
		{"zwpTextInputV3SetContentTypeOp", C.ZWP_TEXT_INPUT_V3_SET_CONTENT_TYPE, zwpTextInputV3SetContentTypeOp},
		{"zwpTextInputV3SetCursorRectangleOp", C.ZWP_TEXT_INPUT_V3_SET_CURSOR_RECTANGLE, zwpTextInputV3SetCursorRectangleOp},
		{"zwpTextInputV3CommitOp", C.ZWP_TEXT_INPUT_V3_COMMIT, zwpTextInputV3CommitOp},
		{"zwpTextInputManagerV3DestroyOp", C.ZWP_TEXT_INPUT_MANAGER_V3_DESTROY, zwpTextInputManagerV3DestroyOp},
		{"zwpTextInputManagerV3GetTextInputOp", C.ZWP_TEXT_INPUT_MANAGER_V3_GET_TEXT_INPUT, zwpTextInputManagerV3GetTextInputOp},
		{"wlMarshalFlagDestroy", C.WL_MARSHAL_FLAG_DESTROY, wlMarshalFlagDestroy},
		{"_WL_SEAT_CAPABILITY_POINTER", C.WL_SEAT_CAPABILITY_POINTER, _WL_SEAT_CAPABILITY_POINTER},
		{"_WL_SEAT_CAPABILITY_KEYBOARD", C.WL_SEAT_CAPABILITY_KEYBOARD, _WL_SEAT_CAPABILITY_KEYBOARD},
		{"_WL_SEAT_CAPABILITY_TOUCH", C.WL_SEAT_CAPABILITY_TOUCH, _WL_SEAT_CAPABILITY_TOUCH},
		{"_WL_POINTER_AXIS_VERTICAL_SCROLL", C.WL_POINTER_AXIS_VERTICAL_SCROLL, _WL_POINTER_AXIS_VERTICAL_SCROLL},
		{"_WL_POINTER_AXIS_HORIZONTAL_SCROLL", C.WL_POINTER_AXIS_HORIZONTAL_SCROLL, _WL_POINTER_AXIS_HORIZONTAL_SCROLL},
		{"_WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1", C.WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1, _WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1},
		{"_WL_KEYBOARD_KEY_STATE_RELEASED", C.WL_KEYBOARD_KEY_STATE_RELEASED, _WL_KEYBOARD_KEY_STATE_RELEASED},
		{"_WL_KEYBOARD_KEY_STATE_PRESSED", C.WL_KEYBOARD_KEY_STATE_PRESSED, _WL_KEYBOARD_KEY_STATE_PRESSED},
		{"_WL_OUTPUT_MODE_CURRENT", C.WL_OUTPUT_MODE_CURRENT, _WL_OUTPUT_MODE_CURRENT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_TOP", C.XDG_TOPLEVEL_RESIZE_EDGE_TOP, _XDG_TOPLEVEL_RESIZE_EDGE_TOP},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM", C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_LEFT", C.XDG_TOPLEVEL_RESIZE_EDGE_LEFT, _XDG_TOPLEVEL_RESIZE_EDGE_LEFT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT", C.XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT, _XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT", C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_RIGHT", C.XDG_TOPLEVEL_RESIZE_EDGE_RIGHT, _XDG_TOPLEVEL_RESIZE_EDGE_RIGHT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT", C.XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT, _XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT},
		{"_XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT", C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT, _XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT},
		{"_ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE", C.ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE, _ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE},
		{"_ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE", C.ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE, _ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE},
		{"_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_INPUT_METHOD", C.ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_INPUT_METHOD, _ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_INPUT_METHOD},
		{"_ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER", C.ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER, _ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_HINT_COMPLETION", C.ZWP_TEXT_INPUT_V3_CONTENT_HINT_COMPLETION, _ZWP_TEXT_INPUT_V3_CONTENT_HINT_COMPLETION},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_HINT_SPELLCHECK", C.ZWP_TEXT_INPUT_V3_CONTENT_HINT_SPELLCHECK, _ZWP_TEXT_INPUT_V3_CONTENT_HINT_SPELLCHECK},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT", C.ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT, _ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA", C.ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA, _ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NUMBER", C.ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NUMBER, _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NUMBER},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PHONE", C.ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PHONE, _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PHONE},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_URL", C.ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_URL, _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_URL},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_EMAIL", C.ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_EMAIL, _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_EMAIL},
		{"_ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD", C.ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD, _ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD},
	} {
		if c.got != c.want {
			panic(fmt.Sprintf("wayland: %s is %d, the headers say %d", c.name, c.got, c.want))
		}
	}
	for _, c := range [...]struct {
		name      string
		got, want uintptr
	}{
		{"wlCursor", unsafe.Sizeof(wlCursor{}), unsafe.Sizeof(C.struct_wl_cursor{})},
		{"wlCursor.ImageCount", unsafe.Offsetof(wlCursor{}.ImageCount), unsafe.Offsetof(C.struct_wl_cursor{}.image_count)},
		{"wlCursor.Images", unsafe.Offsetof(wlCursor{}.Images), unsafe.Offsetof(C.struct_wl_cursor{}.images)},
		{"wlCursor.Name", unsafe.Offsetof(wlCursor{}.Name), unsafe.Offsetof(C.struct_wl_cursor{}.name)},
		{"wlCursorImage", unsafe.Sizeof(wlCursorImage{}), unsafe.Sizeof(C.struct_wl_cursor_image{})},
		{"wlCursorImage.Width", unsafe.Offsetof(wlCursorImage{}.Width), unsafe.Offsetof(C.struct_wl_cursor_image{}.width)},
		{"wlCursorImage.Height", unsafe.Offsetof(wlCursorImage{}.Height), unsafe.Offsetof(C.struct_wl_cursor_image{}.height)},
		{"wlCursorImage.HotspotX", unsafe.Offsetof(wlCursorImage{}.HotspotX), unsafe.Offsetof(C.struct_wl_cursor_image{}.hotspot_x)},
		{"wlCursorImage.HotspotY", unsafe.Offsetof(wlCursorImage{}.HotspotY), unsafe.Offsetof(C.struct_wl_cursor_image{}.hotspot_y)},
		{"wlCursorImage.Delay", unsafe.Offsetof(wlCursorImage{}.Delay), unsafe.Offsetof(C.struct_wl_cursor_image{}.delay)},
	} {
		if c.got != c.want {
			panic(fmt.Sprintf("wayland: %s is %d, the headers say %d", c.name, c.got, c.want))
		}
	}
}

func wlDisplayConnect() (unsafe.Pointer, error) {
	d, err := C.wl_display_connect(nil)
	if d == nil {
		return nil, err
	}
	return unsafe.Pointer(d), nil
}

func wlDisp(d unsafe.Pointer) *C.struct_wl_display { return (*C.struct_wl_display)(d) }
func wlProxy(p unsafe.Pointer) *C.struct_wl_proxy  { return (*C.struct_wl_proxy)(p) }

func wlDisplayDisconnect(d unsafe.Pointer)        { C.wl_display_disconnect(wlDisp(d)) }
func wlDisplayGetFd(d unsafe.Pointer) int32       { return int32(C.wl_display_get_fd(wlDisp(d))) }
func wlDisplayRoundtrip(d unsafe.Pointer)         { C.wl_display_roundtrip(wlDisp(d)) }
func wlDisplayDispatchPending(d unsafe.Pointer)   { C.wl_display_dispatch_pending(wlDisp(d)) }
func wlDisplayPrepareRead(d unsafe.Pointer) int32 { return int32(C.wl_display_prepare_read(wlDisp(d))) }
func wlDisplayCancelRead(d unsafe.Pointer)        { C.wl_display_cancel_read(wlDisp(d)) }

func wlDisplayFlush(d unsafe.Pointer) (int32, error) {
	r, err := C.wl_display_flush(wlDisp(d))
	return int32(r), err
}

func wlDisplayReadEvents(d unsafe.Pointer) (int32, error) {
	r, err := C.wl_display_read_events(wlDisp(d))
	return int32(r), err
}

func wlProxyMarshalArrayFlags(p unsafe.Pointer, op uint32, iface unsafe.Pointer, version, flags uint32, args []uintptr) unsafe.Pointer {
	return unsafe.Pointer(C.wl_proxy_marshal_array_flags(wlProxy(p), C.uint32_t(op), (*C.struct_wl_interface)(iface), C.uint32_t(version), C.uint32_t(flags), (*C.union_wl_argument)(unsafe.Pointer(&args[0]))))
}

func wlProxyGetVersion(p unsafe.Pointer) uint32 { return uint32(C.wl_proxy_get_version(wlProxy(p))) }
func wlProxyDestroy(p unsafe.Pointer)           { C.wl_proxy_destroy(wlProxy(p)) }

func wlAddListener(p unsafe.Pointer, kind int, data unsafe.Pointer) {
	var l unsafe.Pointer
	switch kind {
	case lk_registry:
		l = unsafe.Pointer(&C.gio_registry_listener)
	case lk_surface:
		l = unsafe.Pointer(&C.gio_surface_listener)
	case lk_xdg_surface:
		l = unsafe.Pointer(&C.gio_xdg_surface_listener)
	case lk_xdg_toplevel:
		l = unsafe.Pointer(&C.gio_xdg_toplevel_listener)
	case lk_zxdg_toplevel_decoration_v1:
		l = unsafe.Pointer(&C.gio_zxdg_toplevel_decoration_v1_listener)
	case lk_xdg_wm_base:
		l = unsafe.Pointer(&C.gio_xdg_wm_base_listener)
	case lk_callback:
		l = unsafe.Pointer(&C.gio_callback_listener)
	case lk_output:
		l = unsafe.Pointer(&C.gio_output_listener)
	case lk_seat:
		l = unsafe.Pointer(&C.gio_seat_listener)
	case lk_pointer:
		l = unsafe.Pointer(&C.gio_pointer_listener)
	case lk_touch:
		l = unsafe.Pointer(&C.gio_touch_listener)
	case lk_keyboard:
		l = unsafe.Pointer(&C.gio_keyboard_listener)
	case lk_zwp_text_input_v3:
		l = unsafe.Pointer(&C.gio_zwp_text_input_v3_listener)
	case lk_data_device:
		l = unsafe.Pointer(&C.gio_data_device_listener)
	case lk_data_offer:
		l = unsafe.Pointer(&C.gio_data_offer_listener)
	case lk_data_source:
		l = unsafe.Pointer(&C.gio_data_source_listener)
	default:
		panic("wayland: unknown listener kind")
	}
	C.wl_proxy_add_listener(wlProxy(p), (**[0]byte)(l), data)
}

func wlCString(s string) unsafe.Pointer { return unsafe.Pointer(C.CString(s)) }
func wlFree(p unsafe.Pointer)           { C.free(p) }

func wlCursorThemeLoad(name string, size int, shm unsafe.Pointer) unsafe.Pointer {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return unsafe.Pointer(C.wl_cursor_theme_load(cname, C.int(size), (*C.struct_wl_shm)(shm)))
}

func wlCursorThemeGetCursor(theme unsafe.Pointer, name string) *wlCursor {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return (*wlCursor)(unsafe.Pointer(C.wl_cursor_theme_get_cursor((*C.struct_wl_cursor_theme)(theme), cname)))
}

func wlCursorImageGetBuffer(img *wlCursorImage) unsafe.Pointer {
	return unsafe.Pointer(C.wl_cursor_image_get_buffer((*C.struct_wl_cursor_image)(unsafe.Pointer(img))))
}

func wlCursorThemeDestroy(theme unsafe.Pointer) {
	C.wl_cursor_theme_destroy((*C.struct_wl_cursor_theme)(theme))
}

//export gio_onSeatCapabilities
func gio_onSeatCapabilities(data unsafe.Pointer, seat *C.struct_wl_seat, caps C.uint32_t) {
	wlOnSeatCapabilities(data, unsafe.Pointer(seat), uint32(caps))
}

//export gio_onSeatName
func gio_onSeatName(data unsafe.Pointer, seat *C.struct_wl_seat, name *C.char) {
	wlOnSeatName(data, unsafe.Pointer(seat), C.GoString(name))
}

//export gio_onXdgSurfaceConfigure
func gio_onXdgSurfaceConfigure(data unsafe.Pointer, wmSurf *C.struct_xdg_surface, serial C.uint32_t) {
	wlOnXdgSurfaceConfigure(data, unsafe.Pointer(wmSurf), uint32(serial))
}

//export gio_onToplevelClose
func gio_onToplevelClose(data unsafe.Pointer, topLvl *C.struct_xdg_toplevel) {
	wlOnToplevelClose(data, unsafe.Pointer(topLvl))
}

//export gio_onToplevelConfigure
func gio_onToplevelConfigure(data unsafe.Pointer, topLvl *C.struct_xdg_toplevel, width, height C.int32_t, states *C.struct_wl_array) {
	wlOnToplevelConfigure(data, unsafe.Pointer(topLvl), int32(width), int32(height), unsafe.Pointer(states))
}

//export gio_onToplevelDecorationConfigure
func gio_onToplevelDecorationConfigure(data unsafe.Pointer, deco *C.struct_zxdg_toplevel_decoration_v1, mode C.uint32_t) {
	wlOnToplevelDecorationConfigure(data, unsafe.Pointer(deco), uint32(mode))
}

//export gio_onOutputMode
func gio_onOutputMode(data unsafe.Pointer, output *C.struct_wl_output, flags C.uint32_t, width, height, refresh C.int32_t) {
	wlOnOutputMode(data, unsafe.Pointer(output), uint32(flags), int32(width), int32(height), int32(refresh))
}

//export gio_onOutputGeometry
func gio_onOutputGeometry(data unsafe.Pointer, output *C.struct_wl_output, x, y, physWidth, physHeight, subpixel C.int32_t, make, model *C.char, transform C.int32_t) {
	wlOnOutputGeometry(data, unsafe.Pointer(output), int32(x), int32(y), int32(physWidth), int32(physHeight), int32(subpixel), C.GoString(make), C.GoString(model), int32(transform))
}

//export gio_onOutputScale
func gio_onOutputScale(data unsafe.Pointer, output *C.struct_wl_output, scale C.int32_t) {
	wlOnOutputScale(data, unsafe.Pointer(output), int32(scale))
}

//export gio_onOutputDone
func gio_onOutputDone(data unsafe.Pointer, output *C.struct_wl_output) {
	wlOnOutputDone(data, unsafe.Pointer(output))
}

//export gio_onSurfaceEnter
func gio_onSurfaceEnter(data unsafe.Pointer, surf *C.struct_wl_surface, output *C.struct_wl_output) {
	wlOnSurfaceEnter(data, unsafe.Pointer(surf), unsafe.Pointer(output))
}

//export gio_onSurfaceLeave
func gio_onSurfaceLeave(data unsafe.Pointer, surf *C.struct_wl_surface, output *C.struct_wl_output) {
	wlOnSurfaceLeave(data, unsafe.Pointer(surf), unsafe.Pointer(output))
}

//export gio_onRegistryGlobal
func gio_onRegistryGlobal(data unsafe.Pointer, reg *C.struct_wl_registry, name C.uint32_t, cintf *C.char, version C.uint32_t) {
	wlOnRegistryGlobal(data, unsafe.Pointer(reg), uint32(name), C.GoString(cintf), uint32(version))
}

//export gio_onDataOfferOffer
func gio_onDataOfferOffer(data unsafe.Pointer, offer *C.struct_wl_data_offer, mime *C.char) {
	wlOnDataOfferOffer(data, unsafe.Pointer(offer), C.GoString(mime))
}

//export gio_onDataOfferSourceActions
func gio_onDataOfferSourceActions(data unsafe.Pointer, offer *C.struct_wl_data_offer, acts C.uint32_t) {
	wlOnDataOfferSourceActions(data, unsafe.Pointer(offer), uint32(acts))
}

//export gio_onDataOfferAction
func gio_onDataOfferAction(data unsafe.Pointer, offer *C.struct_wl_data_offer, act C.uint32_t) {
	wlOnDataOfferAction(data, unsafe.Pointer(offer), uint32(act))
}

//export gio_onDataDeviceOffer
func gio_onDataDeviceOffer(data unsafe.Pointer, dataDev *C.struct_wl_data_device, id *C.struct_wl_data_offer) {
	wlOnDataDeviceOffer(data, unsafe.Pointer(dataDev), unsafe.Pointer(id))
}

//export gio_onDataDeviceEnter
func gio_onDataDeviceEnter(data unsafe.Pointer, dataDev *C.struct_wl_data_device, serial C.uint32_t, surf *C.struct_wl_surface, x, y C.wl_fixed_t, id *C.struct_wl_data_offer) {
	wlOnDataDeviceEnter(data, unsafe.Pointer(dataDev), uint32(serial), unsafe.Pointer(surf), int32(x), int32(y), unsafe.Pointer(id))
}

//export gio_onDataDeviceLeave
func gio_onDataDeviceLeave(data unsafe.Pointer, dataDev *C.struct_wl_data_device) {
	wlOnDataDeviceLeave(data, unsafe.Pointer(dataDev))
}

//export gio_onDataDeviceMotion
func gio_onDataDeviceMotion(data unsafe.Pointer, dataDev *C.struct_wl_data_device, t C.uint32_t, x, y C.wl_fixed_t) {
	wlOnDataDeviceMotion(data, unsafe.Pointer(dataDev), uint32(t), int32(x), int32(y))
}

//export gio_onDataDeviceDrop
func gio_onDataDeviceDrop(data unsafe.Pointer, dataDev *C.struct_wl_data_device) {
	wlOnDataDeviceDrop(data, unsafe.Pointer(dataDev))
}

//export gio_onDataDeviceSelection
func gio_onDataDeviceSelection(data unsafe.Pointer, dataDev *C.struct_wl_data_device, id *C.struct_wl_data_offer) {
	wlOnDataDeviceSelection(data, unsafe.Pointer(dataDev), unsafe.Pointer(id))
}

//export gio_onRegistryGlobalRemove
func gio_onRegistryGlobalRemove(data unsafe.Pointer, reg *C.struct_wl_registry, name C.uint32_t) {
	wlOnRegistryGlobalRemove(data, unsafe.Pointer(reg), uint32(name))
}

//export gio_onTouchDown
func gio_onTouchDown(data unsafe.Pointer, touch *C.struct_wl_touch, serial, t C.uint32_t, surf *C.struct_wl_surface, id C.int32_t, x, y C.wl_fixed_t) {
	wlOnTouchDown(data, unsafe.Pointer(touch), uint32(serial), uint32(t), unsafe.Pointer(surf), int32(id), int32(x), int32(y))
}

//export gio_onTouchUp
func gio_onTouchUp(data unsafe.Pointer, touch *C.struct_wl_touch, serial, t C.uint32_t, id C.int32_t) {
	wlOnTouchUp(data, unsafe.Pointer(touch), uint32(serial), uint32(t), int32(id))
}

//export gio_onTouchMotion
func gio_onTouchMotion(data unsafe.Pointer, touch *C.struct_wl_touch, t C.uint32_t, id C.int32_t, x, y C.wl_fixed_t) {
	wlOnTouchMotion(data, unsafe.Pointer(touch), uint32(t), int32(id), int32(x), int32(y))
}

//export gio_onTouchFrame
func gio_onTouchFrame(data unsafe.Pointer, touch *C.struct_wl_touch) {
	wlOnTouchFrame(data, unsafe.Pointer(touch))
}

//export gio_onTouchCancel
func gio_onTouchCancel(data unsafe.Pointer, touch *C.struct_wl_touch) {
	wlOnTouchCancel(data, unsafe.Pointer(touch))
}

//export gio_onPointerEnter
func gio_onPointerEnter(data unsafe.Pointer, pointer *C.struct_wl_pointer, serial C.uint32_t, surf *C.struct_wl_surface, x, y C.wl_fixed_t) {
	wlOnPointerEnter(data, unsafe.Pointer(pointer), uint32(serial), unsafe.Pointer(surf), int32(x), int32(y))
}

//export gio_onPointerLeave
func gio_onPointerLeave(data unsafe.Pointer, p *C.struct_wl_pointer, serial C.uint32_t, surf *C.struct_wl_surface) {
	wlOnPointerLeave(data, unsafe.Pointer(p), uint32(serial), unsafe.Pointer(surf))
}

//export gio_onPointerMotion
func gio_onPointerMotion(data unsafe.Pointer, p *C.struct_wl_pointer, t C.uint32_t, x, y C.wl_fixed_t) {
	wlOnPointerMotion(data, unsafe.Pointer(p), uint32(t), int32(x), int32(y))
}

//export gio_onPointerButton
func gio_onPointerButton(data unsafe.Pointer, p *C.struct_wl_pointer, serial, t, wbtn, state C.uint32_t) {
	wlOnPointerButton(data, unsafe.Pointer(p), uint32(serial), uint32(t), uint32(wbtn), uint32(state))
}

//export gio_onPointerAxis
func gio_onPointerAxis(data unsafe.Pointer, p *C.struct_wl_pointer, t, axis C.uint32_t, value C.wl_fixed_t) {
	wlOnPointerAxis(data, unsafe.Pointer(p), uint32(t), uint32(axis), int32(value))
}

//export gio_onPointerFrame
func gio_onPointerFrame(data unsafe.Pointer, p *C.struct_wl_pointer) {
	wlOnPointerFrame(data, unsafe.Pointer(p))
}

//export gio_onPointerAxisSource
func gio_onPointerAxisSource(data unsafe.Pointer, pointer *C.struct_wl_pointer, source C.uint32_t) {
	wlOnPointerAxisSource(data, unsafe.Pointer(pointer), uint32(source))
}

//export gio_onPointerAxisStop
func gio_onPointerAxisStop(data unsafe.Pointer, p *C.struct_wl_pointer, t, axis C.uint32_t) {
	wlOnPointerAxisStop(data, unsafe.Pointer(p), uint32(t), uint32(axis))
}

//export gio_onPointerAxisDiscrete
func gio_onPointerAxisDiscrete(data unsafe.Pointer, p *C.struct_wl_pointer, axis C.uint32_t, discrete C.int32_t) {
	wlOnPointerAxisDiscrete(data, unsafe.Pointer(p), uint32(axis), int32(discrete))
}

//export gio_onKeyboardKeymap
func gio_onKeyboardKeymap(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, format C.uint32_t, fd C.int32_t, size C.uint32_t) {
	wlOnKeyboardKeymap(data, unsafe.Pointer(keyboard), uint32(format), int32(fd), uint32(size))
}

//export gio_onKeyboardEnter
func gio_onKeyboardEnter(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial C.uint32_t, surf *C.struct_wl_surface, keys *C.struct_wl_array) {
	wlOnKeyboardEnter(data, unsafe.Pointer(keyboard), uint32(serial), unsafe.Pointer(surf), unsafe.Pointer(keys))
}

//export gio_onKeyboardLeave
func gio_onKeyboardLeave(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial C.uint32_t, surf *C.struct_wl_surface) {
	wlOnKeyboardLeave(data, unsafe.Pointer(keyboard), uint32(serial), unsafe.Pointer(surf))
}

//export gio_onKeyboardKey
func gio_onKeyboardKey(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial, timestamp, keyCode, state C.uint32_t) {
	wlOnKeyboardKey(data, unsafe.Pointer(keyboard), uint32(serial), uint32(timestamp), uint32(keyCode), uint32(state))
}

//export gio_onFrameDone
func gio_onFrameDone(data unsafe.Pointer, callback *C.struct_wl_callback, t C.uint32_t) {
	wlOnFrameDone(data, unsafe.Pointer(callback), uint32(t))
}

//export gio_onKeyboardModifiers
func gio_onKeyboardModifiers(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial, depressed, latched, locked, group C.uint32_t) {
	wlOnKeyboardModifiers(data, unsafe.Pointer(keyboard), uint32(serial), uint32(depressed), uint32(latched), uint32(locked), uint32(group))
}

//export gio_onKeyboardRepeatInfo
func gio_onKeyboardRepeatInfo(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, rate, delay C.int32_t) {
	wlOnKeyboardRepeatInfo(data, unsafe.Pointer(keyboard), int32(rate), int32(delay))
}

//export gio_onTextInputEnter
func gio_onTextInputEnter(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, surf *C.struct_wl_surface) {
	wlOnTextInputEnter(data, unsafe.Pointer(im), unsafe.Pointer(surf))
}

//export gio_onTextInputLeave
func gio_onTextInputLeave(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, surf *C.struct_wl_surface) {
	wlOnTextInputLeave(data, unsafe.Pointer(im), unsafe.Pointer(surf))
}

//export gio_onTextInputPreeditString
func gio_onTextInputPreeditString(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, ctxt *C.char, begin, end C.int32_t) {
	wlOnTextInputPreeditString(data, unsafe.Pointer(im), C.GoString(ctxt), ctxt != nil, int32(begin), int32(end))
}

//export gio_onTextInputCommitString
func gio_onTextInputCommitString(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, ctxt *C.char) {
	wlOnTextInputCommitString(data, unsafe.Pointer(im), C.GoString(ctxt), ctxt != nil)
}

//export gio_onTextInputDeleteSurroundingText
func gio_onTextInputDeleteSurroundingText(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, before, after C.uint32_t) {
	wlOnTextInputDeleteSurroundingText(data, unsafe.Pointer(im), uint32(before), uint32(after))
}

//export gio_onTextInputDone
func gio_onTextInputDone(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, serial C.uint32_t) {
	wlOnTextInputDone(data, unsafe.Pointer(im), uint32(serial))
}

//export gio_onDataSourceTarget
func gio_onDataSourceTarget(data unsafe.Pointer, source *C.struct_wl_data_source, mime *C.char) {
	wlOnDataSourceTarget(data, unsafe.Pointer(source), C.GoString(mime))
}

//export gio_onDataSourceSend
func gio_onDataSourceSend(data unsafe.Pointer, source *C.struct_wl_data_source, mime *C.char, fd C.int32_t) {
	wlOnDataSourceSend(data, unsafe.Pointer(source), C.GoString(mime), int32(fd))
}

//export gio_onDataSourceCancelled
func gio_onDataSourceCancelled(data unsafe.Pointer, source *C.struct_wl_data_source) {
	wlOnDataSourceCancelled(data, unsafe.Pointer(source))
}

//export gio_onDataSourceDNDDropPerformed
func gio_onDataSourceDNDDropPerformed(data unsafe.Pointer, source *C.struct_wl_data_source) {
	wlOnDataSourceDNDDropPerformed(data, unsafe.Pointer(source))
}

//export gio_onDataSourceDNDFinished
func gio_onDataSourceDNDFinished(data unsafe.Pointer, source *C.struct_wl_data_source) {
	wlOnDataSourceDNDFinished(data, unsafe.Pointer(source))
}

//export gio_onDataSourceAction
func gio_onDataSourceAction(data unsafe.Pointer, source *C.struct_wl_data_source, act C.uint32_t) {
	wlOnDataSourceAction(data, unsafe.Pointer(source), uint32(act))
}
