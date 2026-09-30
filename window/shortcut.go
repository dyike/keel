package window

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/shortcut"
)

// RegisterShortcut binds an application shortcut to this window, before focus
// dispatch so it works while editing an input. It does not reserve a system-wide
// hotkey. Call before Run for startup windows, or in OnReady for runtime windows.
// The callback runs on the UI thread. The returned function removes the binding.
func (w *Window) RegisterShortcut(accelerator string, handler func()) (func(), error) {
	if handler == nil {
		return nil, capability.ErrInvalidArgument
	}
	chord, err := shortcut.Parse(accelerator)
	if err != nil {
		return nil, err
	}
	key, err := guiKey(chord.Key)
	if err != nil {
		return nil, err
	}
	host := w.Host()
	if host == nil {
		return nil, capability.ErrNotReady
	}
	if host.Ctx().Err() != nil {
		return nil, capability.ErrClosed
	}
	var modifiers gui.Modifier
	if chord.Modifiers&shortcut.Control != 0 {
		modifiers |= gui.ModCtrl
	}
	if chord.Modifiers&shortcut.Alt != 0 {
		modifiers |= gui.ModAlt
	}
	if chord.Modifiers&shortcut.Shift != 0 {
		modifiers |= gui.ModShift
	}
	if chord.Modifiers&shortcut.Super != 0 {
		modifiers |= gui.ModSuper
	}
	id := fmt.Sprintf("keel-shortcut:%d:%d", modifiers, key)
	err = host.RegisterCommand(gui.Command{
		ID: id, Shortcut: gui.Shortcut{Key: key, Modifiers: modifiers}, Global: true,
		Execute: func(e *gui.Event, _ *gui.Window) {
			if e == nil || !e.KeyRepeat {
				handler()
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", capability.ErrConflict, err)
	}
	var once sync.Once
	return func() { once.Do(func() { host.UnregisterCommand(id) }) }, nil
}

func guiKey(key string) (gui.KeyCode, error) {
	if len(key) == 1 {
		if key[0] >= 'a' && key[0] <= 'z' {
			return gui.KeyA + gui.KeyCode(key[0]-'a'), nil
		}
		if key[0] >= '0' && key[0] <= '9' {
			return gui.Key0 + gui.KeyCode(key[0]-'0'), nil
		}
	}
	if strings.HasPrefix(key, "f") {
		if n, err := strconv.Atoi(key[1:]); err == nil && n >= 1 && n <= 25 {
			return gui.KeyF1 + gui.KeyCode(n-1), nil
		}
	}
	keys := map[string]gui.KeyCode{
		"comma": gui.KeyComma, ",": gui.KeyComma, "period": gui.KeyPeriod, ".": gui.KeyPeriod,
		"space": gui.KeySpace, "enter": gui.KeyEnter, "tab": gui.KeyTab, "backspace": gui.KeyBackspace,
		"escape": gui.KeyEscape, "delete": gui.KeyDelete, "insert": gui.KeyInsert,
		"up": gui.KeyUp, "down": gui.KeyDown, "left": gui.KeyLeft, "right": gui.KeyRight,
		"home": gui.KeyHome, "end": gui.KeyEnd, "pageup": gui.KeyPageUp, "pagedown": gui.KeyPageDown,
		"minus": gui.KeyMinus, "equal": gui.KeyEqual, "leftbracket": gui.KeyLeftBracket, "rightbracket": gui.KeyRightBracket,
		"semicolon": gui.KeySemicolon, "quote": gui.KeyApostrophe, "backslash": gui.KeyBackslash, "slash": gui.KeySlash, "backquote": gui.KeyGraveAccent,
	}
	if code, ok := keys[key]; ok {
		return code, nil
	}
	return gui.KeyInvalid, fmt.Errorf("%w: unknown shortcut key %q", capability.ErrInvalidArgument, key)
}
