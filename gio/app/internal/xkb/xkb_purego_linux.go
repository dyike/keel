// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && (!cgo || purego)

package xkb

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The libxkbcommon calls of xkb.go without cgo, loaded at run time. Every
// argument is an integer or a pointer. Keel patch.

var lib struct {
	once sync.Once
	err  error

	contextNew, contextUnref                              uintptr
	composeTableNew, composeTableUnref                    uintptr
	composeStateNew, composeStateUnref                    uintptr
	composeFeed, composeStatus, composeUTF8, composeReset uintptr
	keymapNewFromBuffer, keymapUnref, keymapKeyRepeats    uintptr
	stateNew, stateUnref, stateModNameIsActive            uintptr
	stateKeyGetOneSym, stateKeyGetUTF8, stateUpdateMask   uintptr
	keymapNewFromNames                                    uintptr
}

// Load loads libxkbcommon; the other functions require it.
func Load() error {
	lib.once.Do(func() {
		var h uintptr
		for _, name := range []string{"libxkbcommon.so.0", "libxkbcommon.so"} {
			if h, lib.err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL); lib.err == nil {
				break
			}
		}
		if lib.err != nil {
			return
		}
		for _, s := range []struct {
			name string
			addr *uintptr
		}{
			{"xkb_context_new", &lib.contextNew},
			{"xkb_context_unref", &lib.contextUnref},
			{"xkb_compose_table_new_from_locale", &lib.composeTableNew},
			{"xkb_compose_table_unref", &lib.composeTableUnref},
			{"xkb_compose_state_new", &lib.composeStateNew},
			{"xkb_compose_state_unref", &lib.composeStateUnref},
			{"xkb_compose_state_feed", &lib.composeFeed},
			{"xkb_compose_state_get_status", &lib.composeStatus},
			{"xkb_compose_state_get_utf8", &lib.composeUTF8},
			{"xkb_compose_state_reset", &lib.composeReset},
			{"xkb_keymap_new_from_buffer", &lib.keymapNewFromBuffer},
			{"xkb_keymap_unref", &lib.keymapUnref},
			{"xkb_keymap_key_repeats", &lib.keymapKeyRepeats},
			{"xkb_state_new", &lib.stateNew},
			{"xkb_state_unref", &lib.stateUnref},
			{"xkb_state_mod_name_is_active", &lib.stateModNameIsActive},
			{"xkb_state_key_get_one_sym", &lib.stateKeyGetOneSym},
			{"xkb_state_key_get_utf8", &lib.stateKeyGetUTF8},
			{"xkb_state_update_mask", &lib.stateUpdateMask},
			{"xkb_keymap_new_from_names", &lib.keymapNewFromNames},
		} {
			if *s.addr, lib.err = purego.Dlsym(h, s.name); lib.err != nil {
				return
			}
		}
	})
	return lib.err
}

// call calls a libxkbcommon function; pointers to Go memory stay valid and
// in place for the call.
//
//go:uintptrescapes
func call(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

func ptr(p unsafe.Pointer) uintptr { return uintptr(p) }

// cptr returns a C pointer result as unsafe.Pointer; it points outside the
// Go heap.
func cptr(r uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&r)) }

func contextNew() unsafe.Pointer {
	if Load() != nil {
		return nil
	}
	return cptr(call(lib.contextNew, 0)) // XKB_CONTEXT_NO_FLAGS
}

func contextUnref(ctx unsafe.Pointer) { call(lib.contextUnref, ptr(ctx)) }

func composeTableNewFromLocale(ctx unsafe.Pointer, locale string) unsafe.Pointer {
	loc := append([]byte(locale), 0)
	return cptr(call(lib.composeTableNew, ptr(ctx), uintptr(unsafe.Pointer(&loc[0])), 0))
}

func composeTableUnref(t unsafe.Pointer) { call(lib.composeTableUnref, ptr(t)) }

func composeStateNew(t unsafe.Pointer) unsafe.Pointer {
	return cptr(call(lib.composeStateNew, ptr(t), 0))
}

func composeStateUnref(s unsafe.Pointer)            { call(lib.composeStateUnref, ptr(s)) }
func composeStateFeed(s unsafe.Pointer, sym uint32) { call(lib.composeFeed, ptr(s), uintptr(sym)) }

// composeStateGetStatus returns an enum, which is 32 bits wide.
func composeStateGetStatus(s unsafe.Pointer) int {
	return int(int32(call(lib.composeStatus, ptr(s))))
}

func composeStateGetUTF8(s unsafe.Pointer, buf []byte) int {
	return int(int32(call(lib.composeUTF8, ptr(s), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))))
}

func composeStateReset(s unsafe.Pointer) { call(lib.composeReset, ptr(s)) }

func keymapNewFromBuffer(ctx unsafe.Pointer, data []byte) unsafe.Pointer {
	return cptr(call(lib.keymapNewFromBuffer, ptr(ctx), uintptr(unsafe.Pointer(unsafe.SliceData(data))), uintptr(len(data)), xkbKeymapFormatTextV1, 0))
}

func keymapUnref(km unsafe.Pointer) { call(lib.keymapUnref, ptr(km)) }

func keymapKeyRepeats(km unsafe.Pointer, kc uint32) int {
	return int(int32(call(lib.keymapKeyRepeats, ptr(km), uintptr(kc))))
}

func stateNew(km unsafe.Pointer) unsafe.Pointer { return cptr(call(lib.stateNew, ptr(km))) }
func stateUnref(s unsafe.Pointer)               { call(lib.stateUnref, ptr(s)) }

// stateModNameIsActive takes a NUL-terminated name.
func stateModNameIsActive(s unsafe.Pointer, name string, typ int) int {
	return int(int32(call(lib.stateModNameIsActive, ptr(s), uintptr(unsafe.Pointer(unsafe.StringData(name))), uintptr(typ))))
}

func stateKeyGetOneSym(s unsafe.Pointer, kc uint32) uint32 {
	return uint32(call(lib.stateKeyGetOneSym, ptr(s), uintptr(kc)))
}

func stateKeyGetUTF8(s unsafe.Pointer, kc uint32, buf []byte) int {
	return int(int32(call(lib.stateKeyGetUTF8, ptr(s), uintptr(kc), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))))
}

func stateUpdateMask(s unsafe.Pointer, depressed, latched, locked, depressedGroup, latchedGroup, lockedGroup uint32) {
	call(lib.stateUpdateMask, ptr(s), uintptr(depressed), uintptr(latched), uintptr(locked), uintptr(depressedGroup), uintptr(latchedGroup), uintptr(lockedGroup))
}

// keymapNewFromNames compiles the default keymap (RMLVO from the
// environment, else xkeyboard-config's defaults).
func keymapNewFromNames(ctx unsafe.Pointer) unsafe.Pointer {
	return cptr(call(lib.keymapNewFromNames, ptr(ctx), 0, 0))
}
