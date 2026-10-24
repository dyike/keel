// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package xkb

import (
	"os"
	"testing"

	"github.com/dyike/keel/gio/io/key"
)

// TestDispatchKey compiles the default keymap and dispatches keys through
// either binding (cgo, or purego). Keel patch. Set GIO_XKB_REQUIRED=1 to
// fail rather than skip without libxkbcommon or keymap data.
func TestDispatchKey(t *testing.T) {
	skip := t.Skip
	if os.Getenv("GIO_XKB_REQUIRED") == "1" {
		skip = t.Fatal
	}
	x, err := New()
	if err != nil {
		skip(err)
		return
	}
	defer x.Destroy()
	km := keymapNewFromNames(x.Ctx)
	if km == nil {
		skip("no default keymap")
		return
	}
	x.SetKeymap(km, stateNew(km))
	const keyA = 30 + 8 // evdev KEY_A, offset by 8 as X11 and Wayland do
	check := func(wantSym uint32, wantText string, wantMods key.Modifiers) {
		t.Helper()
		sym, events := x.DispatchKey(keyA, key.Press)
		if sym != wantSym {
			t.Errorf("keysym %#x, want %#x", sym, wantSym)
		}
		var name key.Name
		var text string
		var mods key.Modifiers
		for _, e := range events {
			switch e := e.(type) {
			case key.Event:
				name, mods = e.Name, e.Modifiers
			case key.EditEvent:
				text = e.Text
			}
		}
		if name != "A" || text != wantText || mods != wantMods {
			t.Errorf("events %+v: name %q text %q mods %v, want A %q %v", events, name, text, mods, wantText, wantMods)
		}
	}
	check('a', "a", 0)
	// Shift is modifier index 0 in every xkeyboard-config keymap.
	x.UpdateMask(1, 0, 0, 0, 0, 0)
	if x.Modifiers() != key.ModShift {
		t.Errorf("modifiers %v, want Shift", x.Modifiers())
	}
	check('A', "A", key.ModShift)
	x.UpdateMask(0, 0, 0, 0, 0, 0)
	if !x.IsRepeatKey(keyA) {
		t.Error("A does not repeat")
	}
}
