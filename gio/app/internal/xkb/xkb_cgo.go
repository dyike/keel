// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android && cgo && !purego) || freebsd || openbsd

package xkb

/*
#cgo linux pkg-config: xkbcommon
#cgo freebsd openbsd CFLAGS: -I/usr/local/include
#cgo freebsd openbsd LDFLAGS: -L/usr/local/lib -lxkbcommon

#include <stdlib.h>
#include <xkbcommon/xkbcommon.h>
#include <xkbcommon/xkbcommon-compose.h>
*/
import "C"

import "unsafe"

// The libxkbcommon calls of xkb.go through cgo. Keel patch: moved from
// xkb_unix.go behind the functions xkb_purego_linux.go also implements.

// The values xkb.go spells out must match the headers.
func init() {
	for _, c := range [...][2]uint64{
		{C.XKB_STATE_MODS_EFFECTIVE, xkbStateModsEffective},
		{C.XKB_KEYMAP_FORMAT_TEXT_V1, xkbKeymapFormatTextV1},
		{C.XKB_COMPOSE_NOTHING, xkbComposeNothing},
		{C.XKB_COMPOSE_COMPOSING, xkbComposeComposing},
		{C.XKB_COMPOSE_COMPOSED, xkbComposeComposed},
		{C.XKB_COMPOSE_CANCELLED, xkbComposeCancelled},
		{C.XKB_KEY_KP_0, xkbKeyKP0},
		{C.XKB_KEY_KP_9, xkbKeyKP9},
		{C.XKB_KEY_Escape, xkbKeyEscape},
		{C.XKB_KEY_Left, xkbKeyLeft},
		{C.XKB_KEY_Right, xkbKeyRight},
		{C.XKB_KEY_Return, xkbKeyReturn},
		{C.XKB_KEY_Up, xkbKeyUp},
		{C.XKB_KEY_Down, xkbKeyDown},
		{C.XKB_KEY_Home, xkbKeyHome},
		{C.XKB_KEY_End, xkbKeyEnd},
		{C.XKB_KEY_BackSpace, xkbKeyBackSpace},
		{C.XKB_KEY_Delete, xkbKeyDelete},
		{C.XKB_KEY_Page_Up, xkbKeyPageUp},
		{C.XKB_KEY_Page_Down, xkbKeyPageDown},
		{C.XKB_KEY_F1, xkbKeyF1},
		{C.XKB_KEY_F2, xkbKeyF2},
		{C.XKB_KEY_F3, xkbKeyF3},
		{C.XKB_KEY_F4, xkbKeyF4},
		{C.XKB_KEY_F5, xkbKeyF5},
		{C.XKB_KEY_F6, xkbKeyF6},
		{C.XKB_KEY_F7, xkbKeyF7},
		{C.XKB_KEY_F8, xkbKeyF8},
		{C.XKB_KEY_F9, xkbKeyF9},
		{C.XKB_KEY_F10, xkbKeyF10},
		{C.XKB_KEY_F11, xkbKeyF11},
		{C.XKB_KEY_F12, xkbKeyF12},
		{C.XKB_KEY_Tab, xkbKeyTab},
		{C.XKB_KEY_ISO_Left_Tab, xkbKeyISOLeftTab},
		{C.XKB_KEY_Control_L, xkbKeyControlL},
		{C.XKB_KEY_Control_R, xkbKeyControlR},
		{C.XKB_KEY_Shift_L, xkbKeyShiftL},
		{C.XKB_KEY_Shift_R, xkbKeyShiftR},
		{C.XKB_KEY_Alt_L, xkbKeyAltL},
		{C.XKB_KEY_Alt_R, xkbKeyAltR},
		{C.XKB_KEY_Super_L, xkbKeySuperL},
		{C.XKB_KEY_Super_R, xkbKeySuperR},
		{C.XKB_KEY_KP_Space, xkbKeyKPSpace},
		{C.XKB_KEY_KP_Tab, xkbKeyKPTab},
		{C.XKB_KEY_KP_Enter, xkbKeyKPEnter},
		{C.XKB_KEY_KP_F1, xkbKeyKPF1},
		{C.XKB_KEY_KP_F2, xkbKeyKPF2},
		{C.XKB_KEY_KP_F3, xkbKeyKPF3},
		{C.XKB_KEY_KP_F4, xkbKeyKPF4},
		{C.XKB_KEY_KP_Home, xkbKeyKPHome},
		{C.XKB_KEY_KP_Left, xkbKeyKPLeft},
		{C.XKB_KEY_KP_Up, xkbKeyKPUp},
		{C.XKB_KEY_KP_Right, xkbKeyKPRight},
		{C.XKB_KEY_KP_Down, xkbKeyKPDown},
		{C.XKB_KEY_KP_Prior, xkbKeyKPPrior},
		{C.XKB_KEY_KP_Next, xkbKeyKPNext},
		{C.XKB_KEY_KP_End, xkbKeyKPEnd},
		{C.XKB_KEY_KP_Begin, xkbKeyKPBegin},
		{C.XKB_KEY_KP_Insert, xkbKeyKPInsert},
		{C.XKB_KEY_KP_Delete, xkbKeyKPDelete},
		{C.XKB_KEY_KP_Multiply, xkbKeyKPMultiply},
		{C.XKB_KEY_KP_Add, xkbKeyKPAdd},
		{C.XKB_KEY_KP_Separator, xkbKeyKPSeparator},
		{C.XKB_KEY_KP_Subtract, xkbKeyKPSubtract},
		{C.XKB_KEY_KP_Decimal, xkbKeyKPDecimal},
		{C.XKB_KEY_KP_Divide, xkbKeyKPDivide},
		{C.XKB_KEY_KP_Equal, xkbKeyKPEqual},
	} {
		if c[0] != c[1] {
			panic("xkb: value differs from xkbcommon's headers")
		}
	}
}

func contextNew() unsafe.Pointer {
	return unsafe.Pointer(C.xkb_context_new(C.XKB_CONTEXT_NO_FLAGS))
}

func contextUnref(ctx unsafe.Pointer) { C.xkb_context_unref((*C.struct_xkb_context)(ctx)) }

func composeTableNewFromLocale(ctx unsafe.Pointer, locale string) unsafe.Pointer {
	cloc := C.CString(locale)
	defer C.free(unsafe.Pointer(cloc))
	return unsafe.Pointer(C.xkb_compose_table_new_from_locale((*C.struct_xkb_context)(ctx), cloc, C.XKB_COMPOSE_COMPILE_NO_FLAGS))
}

func composeTableUnref(t unsafe.Pointer) {
	C.xkb_compose_table_unref((*C.struct_xkb_compose_table)(t))
}

func composeStateNew(t unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_compose_state_new((*C.struct_xkb_compose_table)(t), C.XKB_COMPOSE_STATE_NO_FLAGS))
}

func composeStateUnref(s unsafe.Pointer) {
	C.xkb_compose_state_unref((*C.struct_xkb_compose_state)(s))
}

func composeStateFeed(s unsafe.Pointer, sym uint32) {
	C.xkb_compose_state_feed((*C.struct_xkb_compose_state)(s), C.xkb_keysym_t(sym))
}

func composeStateGetStatus(s unsafe.Pointer) int {
	return int(C.xkb_compose_state_get_status((*C.struct_xkb_compose_state)(s)))
}

func composeStateGetUTF8(s unsafe.Pointer, buf []byte) int {
	return int(C.xkb_compose_state_get_utf8((*C.struct_xkb_compose_state)(s), (*C.char)(unsafe.Pointer(unsafe.SliceData(buf))), C.size_t(len(buf))))
}

func composeStateReset(s unsafe.Pointer) {
	C.xkb_compose_state_reset((*C.struct_xkb_compose_state)(s))
}

func keymapNewFromBuffer(ctx unsafe.Pointer, data []byte) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_keymap_new_from_buffer((*C.struct_xkb_context)(ctx), (*C.char)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), C.XKB_KEYMAP_FORMAT_TEXT_V1, C.XKB_KEYMAP_COMPILE_NO_FLAGS))
}

func keymapUnref(km unsafe.Pointer) { C.xkb_keymap_unref((*C.struct_xkb_keymap)(km)) }

func keymapKeyRepeats(km unsafe.Pointer, kc uint32) int {
	return int(C.xkb_keymap_key_repeats((*C.struct_xkb_keymap)(km), C.xkb_keycode_t(kc)))
}

func stateNew(km unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_state_new((*C.struct_xkb_keymap)(km)))
}

func stateUnref(s unsafe.Pointer) { C.xkb_state_unref((*C.struct_xkb_state)(s)) }

// stateModNameIsActive takes a NUL-terminated name.
func stateModNameIsActive(s unsafe.Pointer, name string, typ int) int {
	return int(C.xkb_state_mod_name_is_active((*C.struct_xkb_state)(s), (*C.char)(unsafe.Pointer(unsafe.StringData(name))), C.enum_xkb_state_component(typ)))
}

func stateKeyGetOneSym(s unsafe.Pointer, kc uint32) uint32 {
	return uint32(C.xkb_state_key_get_one_sym((*C.struct_xkb_state)(s), C.xkb_keycode_t(kc)))
}

func stateKeyGetUTF8(s unsafe.Pointer, kc uint32, buf []byte) int {
	return int(C.xkb_state_key_get_utf8((*C.struct_xkb_state)(s), C.xkb_keycode_t(kc), (*C.char)(unsafe.Pointer(unsafe.SliceData(buf))), C.size_t(len(buf))))
}

func stateUpdateMask(s unsafe.Pointer, depressed, latched, locked, depressedGroup, latchedGroup, lockedGroup uint32) {
	C.xkb_state_update_mask((*C.struct_xkb_state)(s), C.xkb_mod_mask_t(depressed), C.xkb_mod_mask_t(latched), C.xkb_mod_mask_t(locked),
		C.xkb_layout_index_t(depressedGroup), C.xkb_layout_index_t(latchedGroup), C.xkb_layout_index_t(lockedGroup))
}

// keymapNewFromNames compiles the default keymap (RMLVO from the
// environment, else xkeyboard-config's defaults).
func keymapNewFromNames(ctx unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_keymap_new_from_names((*C.struct_xkb_context)(ctx), nil, C.XKB_KEYMAP_COMPILE_NO_FLAGS))
}
