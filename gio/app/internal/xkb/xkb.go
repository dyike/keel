// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android) || freebsd || openbsd

// Package xkb implements a Go interface for the X Keyboard Extension library.
package xkb

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
)

// The libxkbcommon calls live in xkb_cgo.go and xkb_purego_linux.go.
// Keel patch: this file holds the logic both share, over the library's
// opaque pointers.

type Context struct {
	Ctx       unsafe.Pointer // struct xkb_context *
	keyMap    unsafe.Pointer // struct xkb_keymap *
	state     unsafe.Pointer // struct xkb_state *
	compTable unsafe.Pointer // struct xkb_compose_table *
	compState unsafe.Pointer // struct xkb_compose_state *
	utf8Buf   []byte
}

// libxkbcommon values; xkb_cgo.go checks them against its headers.
const (
	xkbStateModsEffective = 1 << 3
	xkbKeymapFormatTextV1 = 1

	xkbComposeNothing   = 0
	xkbComposeComposing = 1
	xkbComposeComposed  = 2
	xkbComposeCancelled = 3
)

const (
	_XKB_MOD_NAME_CTRL  = "Control\x00"
	_XKB_MOD_NAME_SHIFT = "Shift\x00"
	_XKB_MOD_NAME_ALT   = "Mod1\x00"
	_XKB_MOD_NAME_LOGO  = "Mod4\x00"
)

func (x *Context) Destroy() {
	if x.compState != nil {
		composeStateUnref(x.compState)
		x.compState = nil
	}
	if x.compTable != nil {
		composeTableUnref(x.compTable)
		x.compTable = nil
	}
	x.DestroyKeymapState()
	if x.Ctx != nil {
		contextUnref(x.Ctx)
		x.Ctx = nil
	}
}

func New() (*Context, error) {
	ctx := &Context{
		Ctx: contextNew(),
	}
	if ctx.Ctx == nil {
		return nil, errors.New("newXKB: xkb_context_new failed")
	}
	locale := os.Getenv("LC_ALL")
	if locale == "" {
		locale = os.Getenv("LC_CTYPE")
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}
	if locale == "" {
		locale = "C"
	}
	ctx.compTable = composeTableNewFromLocale(ctx.Ctx, locale)
	if ctx.compTable == nil {
		ctx.Destroy()
		return nil, errors.New("newXKB: xkb_compose_table_new_from_locale failed")
	}
	ctx.compState = composeStateNew(ctx.compTable)
	if ctx.compState == nil {
		ctx.Destroy()
		return nil, errors.New("newXKB: xkb_compose_state_new failed")
	}
	return ctx, nil
}

func (x *Context) DestroyKeymapState() {
	if x.state != nil {
		stateUnref(x.state)
		x.state = nil
	}
	if x.keyMap != nil {
		keymapUnref(x.keyMap)
		x.keyMap = nil
	}
}

// SetKeymap sets the keymap and state. The context takes ownership of the
// keymap and state and frees them in Destroy.
func (x *Context) SetKeymap(xkbKeyMap, xkbState unsafe.Pointer) {
	x.DestroyKeymapState()
	x.keyMap = xkbKeyMap
	x.state = xkbState
}

func (x *Context) LoadKeymap(format int, fd int, size int) error {
	x.DestroyKeymapState()
	mapData, err := syscall.Mmap(int(fd), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("newXKB: mmap of keymap failed: %v", err)
	}
	defer syscall.Munmap(mapData)
	keyMap := keymapNewFromBuffer(x.Ctx, mapData[:size-1])
	if keyMap == nil {
		return errors.New("newXKB: xkb_keymap_new_from_buffer failed")
	}
	state := stateNew(keyMap)
	if state == nil {
		keymapUnref(keyMap)
		return errors.New("newXKB: xkb_state_new failed")
	}
	x.keyMap = keyMap
	x.state = state
	return nil
}

func (x *Context) Modifiers() key.Modifiers {
	var mods key.Modifiers
	if x.state == nil {
		return mods
	}

	if stateModNameIsActive(x.state, _XKB_MOD_NAME_CTRL, xkbStateModsEffective) == 1 {
		mods |= key.ModCtrl
	}
	if stateModNameIsActive(x.state, _XKB_MOD_NAME_SHIFT, xkbStateModsEffective) == 1 {
		mods |= key.ModShift
	}
	if stateModNameIsActive(x.state, _XKB_MOD_NAME_ALT, xkbStateModsEffective) == 1 {
		mods |= key.ModAlt
	}
	if stateModNameIsActive(x.state, _XKB_MOD_NAME_LOGO, xkbStateModsEffective) == 1 {
		mods |= key.ModSuper
	}
	return mods
}

func (x *Context) DispatchKey(keyCode uint32, state key.State) (symbol uint32, events []event.Event) {
	if x.state == nil {
		return
	}
	kc := keyCode
	if len(x.utf8Buf) == 0 {
		x.utf8Buf = make([]byte, 1)
	}
	sym := stateKeyGetOneSym(x.state, kc)
	symbol = sym
	if name, ok := convertKeysym(sym); ok {
		cmd := key.Event{
			Name:      name,
			Modifiers: x.Modifiers(),
			State:     state,
		}
		// Ensure that a physical backtab key is translated to
		// Shift-Tab.
		if sym == xkbKeyISOLeftTab {
			cmd.Modifiers |= key.ModShift
		}
		events = append(events, cmd)
	}
	composeStateFeed(x.compState, sym)
	var str []byte
	switch composeStateGetStatus(x.compState) {
	case xkbComposeCancelled, xkbComposeComposing:
		return
	case xkbComposeComposed:
		size := composeStateGetUTF8(x.compState, x.utf8Buf)
		if size >= len(x.utf8Buf) {
			x.utf8Buf = make([]byte, size+1)
			size = composeStateGetUTF8(x.compState, x.utf8Buf)
		}
		composeStateReset(x.compState)
		str = x.utf8Buf[:size]
	case xkbComposeNothing:
		mod := x.Modifiers()
		if mod&(key.ModCtrl|key.ModAlt|key.ModSuper) == 0 {
			str = x.charsForKeycode(kc)
		}
	}
	// Report only printable runes.
	var n int
	for n < len(str) {
		r, s := utf8.DecodeRune(str)
		if unicode.IsPrint(r) {
			n += s
		} else {
			copy(str[n:], str[n+s:])
			str = str[:len(str)-s]
		}
	}
	if state == key.Press && len(str) > 0 {
		events = append(events, key.EditEvent{Text: string(str)})
	}
	return
}

func (x *Context) charsForKeycode(keyCode uint32) []byte {
	size := stateKeyGetUTF8(x.state, keyCode, x.utf8Buf)
	if size >= len(x.utf8Buf) {
		x.utf8Buf = make([]byte, size+1)
		size = stateKeyGetUTF8(x.state, keyCode, x.utf8Buf)
	}
	return x.utf8Buf[:size]
}

func (x *Context) IsRepeatKey(keyCode uint32) bool {
	if x.state == nil {
		return false
	}
	return keymapKeyRepeats(x.keyMap, keyCode) == 1
}

func (x *Context) UpdateMask(depressed, latched, locked, depressedGroup, latchedGroup, lockedGroup uint32) {
	if x.state == nil {
		return
	}
	stateUpdateMask(x.state, depressed, latched, locked, depressedGroup, latchedGroup, lockedGroup)
}

// Keysyms, from xkbcommon-keysyms.h.
const (
	xkbKeyKP0         = 0xffb0
	xkbKeyKP9         = 0xffb9
	xkbKeyEscape      = 0xff1b
	xkbKeyLeft        = 0xff51
	xkbKeyRight       = 0xff53
	xkbKeyReturn      = 0xff0d
	xkbKeyUp          = 0xff52
	xkbKeyDown        = 0xff54
	xkbKeyHome        = 0xff50
	xkbKeyEnd         = 0xff57
	xkbKeyBackSpace   = 0xff08
	xkbKeyDelete      = 0xffff
	xkbKeyPageUp      = 0xff55
	xkbKeyPageDown    = 0xff56
	xkbKeyF1          = 0xffbe
	xkbKeyF2          = 0xffbf
	xkbKeyF3          = 0xffc0
	xkbKeyF4          = 0xffc1
	xkbKeyF5          = 0xffc2
	xkbKeyF6          = 0xffc3
	xkbKeyF7          = 0xffc4
	xkbKeyF8          = 0xffc5
	xkbKeyF9          = 0xffc6
	xkbKeyF10         = 0xffc7
	xkbKeyF11         = 0xffc8
	xkbKeyF12         = 0xffc9
	xkbKeyTab         = 0xff09
	xkbKeyISOLeftTab  = 0xfe20
	xkbKeyControlL    = 0xffe3
	xkbKeyControlR    = 0xffe4
	xkbKeyShiftL      = 0xffe1
	xkbKeyShiftR      = 0xffe2
	xkbKeyAltL        = 0xffe9
	xkbKeyAltR        = 0xffea
	xkbKeySuperL      = 0xffeb
	xkbKeySuperR      = 0xffec
	xkbKeyKPSpace     = 0xff80
	xkbKeyKPTab       = 0xff89
	xkbKeyKPEnter     = 0xff8d
	xkbKeyKPF1        = 0xff91
	xkbKeyKPF2        = 0xff92
	xkbKeyKPF3        = 0xff93
	xkbKeyKPF4        = 0xff94
	xkbKeyKPHome      = 0xff95
	xkbKeyKPLeft      = 0xff96
	xkbKeyKPUp        = 0xff97
	xkbKeyKPRight     = 0xff98
	xkbKeyKPDown      = 0xff99
	xkbKeyKPPrior     = 0xff9a
	xkbKeyKPNext      = 0xff9b
	xkbKeyKPEnd       = 0xff9c
	xkbKeyKPBegin     = 0xff9d
	xkbKeyKPInsert    = 0xff9e
	xkbKeyKPDelete    = 0xff9f
	xkbKeyKPMultiply  = 0xffaa
	xkbKeyKPAdd       = 0xffab
	xkbKeyKPSeparator = 0xffac
	xkbKeyKPSubtract  = 0xffad
	xkbKeyKPDecimal   = 0xffae
	xkbKeyKPDivide    = 0xffaf
	xkbKeyKPEqual     = 0xffbd
)

func convertKeysym(s uint32) (key.Name, bool) {
	if 'a' <= s && s <= 'z' {
		return key.Name(rune(s - 'a' + 'A')), true
	}
	if xkbKeyKP0 <= s && s <= xkbKeyKP9 {
		return key.Name(rune(s - xkbKeyKP0 + '0')), true
	}
	if ' ' < s && s <= '~' {
		return key.Name(rune(s)), true
	}
	var n key.Name
	switch s {
	case xkbKeyEscape:
		n = key.NameEscape
	case xkbKeyLeft:
		n = key.NameLeftArrow
	case xkbKeyRight:
		n = key.NameRightArrow
	case xkbKeyReturn:
		n = key.NameReturn
	case xkbKeyUp:
		n = key.NameUpArrow
	case xkbKeyDown:
		n = key.NameDownArrow
	case xkbKeyHome:
		n = key.NameHome
	case xkbKeyEnd:
		n = key.NameEnd
	case xkbKeyBackSpace:
		n = key.NameDeleteBackward
	case xkbKeyDelete:
		n = key.NameDeleteForward
	case xkbKeyPageUp:
		n = key.NamePageUp
	case xkbKeyPageDown:
		n = key.NamePageDown
	case xkbKeyF1:
		n = key.NameF1
	case xkbKeyF2:
		n = key.NameF2
	case xkbKeyF3:
		n = key.NameF3
	case xkbKeyF4:
		n = key.NameF4
	case xkbKeyF5:
		n = key.NameF5
	case xkbKeyF6:
		n = key.NameF6
	case xkbKeyF7:
		n = key.NameF7
	case xkbKeyF8:
		n = key.NameF8
	case xkbKeyF9:
		n = key.NameF9
	case xkbKeyF10:
		n = key.NameF10
	case xkbKeyF11:
		n = key.NameF11
	case xkbKeyF12:
		n = key.NameF12
	case xkbKeyTab, xkbKeyISOLeftTab:
		n = key.NameTab
	case 0x20:
		n = key.NameSpace
	case xkbKeyControlL, xkbKeyControlR:
		n = key.NameCtrl
	case xkbKeyShiftL, xkbKeyShiftR:
		n = key.NameShift
	case xkbKeyAltL, xkbKeyAltR:
		n = key.NameAlt
	case xkbKeySuperL, xkbKeySuperR:
		n = key.NameSuper

	case xkbKeyKPSpace:
		n = key.NameSpace
	case xkbKeyKPTab:
		n = key.NameTab
	case xkbKeyKPEnter:
		n = key.NameEnter
	case xkbKeyKPF1:
		n = key.NameF1
	case xkbKeyKPF2:
		n = key.NameF2
	case xkbKeyKPF3:
		n = key.NameF3
	case xkbKeyKPF4:
		n = key.NameF4
	case xkbKeyKPHome:
		n = key.NameHome
	case xkbKeyKPLeft:
		n = key.NameLeftArrow
	case xkbKeyKPUp:
		n = key.NameUpArrow
	case xkbKeyKPRight:
		n = key.NameRightArrow
	case xkbKeyKPDown:
		n = key.NameDownArrow
	case xkbKeyKPPrior:
		// not supported
		return "", false
	case xkbKeyKPNext:
		// not supported
		return "", false
	case xkbKeyKPEnd:
		n = key.NameEnd
	case xkbKeyKPBegin:
		n = key.NameHome
	case xkbKeyKPInsert:
		// not supported
		return "", false
	case xkbKeyKPDelete:
		n = key.NameDeleteForward
	case xkbKeyKPMultiply:
		n = "*"
	case xkbKeyKPAdd:
		n = "+"
	case xkbKeyKPSeparator:
		// not supported
		return "", false
	case xkbKeyKPSubtract:
		n = "-"
	case xkbKeyKPDecimal:
		// TODO(dh): does a German keyboard layout also translate the numpad key to XKB_KEY_KP_DECIMAL? Because in
		// German, the decimal is a comma, not a period.
		n = "."
	case xkbKeyKPDivide:
		n = "/"
	case xkbKeyKPEqual:
		n = "="

	default:
		return "", false
	}
	return n, true
}
