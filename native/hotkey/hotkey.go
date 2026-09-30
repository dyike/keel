// Package hotkey registers system-wide shortcuts that fire even when another
// application is in front.
package hotkey

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

// Register reserves an accelerator such as "cmd+shift+k". fn runs on its own
// goroutine, so wrap UI changes in ui.Update. Presses that arrive while fn is
// running are merged. Call the returned func to unregister.
func Register(accel string, fn func()) (unregister func() error, err error) {
	key, mods, err := parse(accel)
	if err != nil {
		return nil, err
	}
	if fn == nil {
		return nil, native.ErrInvalidArgument
	}
	return sys.Hotkey(key, mods, fn)
}

func parse(s string) (key string, mods uint, err error) {
	for _, p := range strings.Split(strings.ToLower(s), "+") {
		switch p = strings.TrimSpace(p); p {
		case "ctrl", "control":
			mods |= sys.ModCtrl
		case "alt", "option":
			mods |= sys.ModAlt
		case "shift":
			mods |= sys.ModShift
		case "cmd", "command", "super":
			mods |= sys.ModCmd
		default:
			if p == "" || key != "" {
				return "", 0, fmt.Errorf("%w: hotkey %q", native.ErrInvalidArgument, s)
			}
			key = p
		}
	}
	if key == "" || mods == 0 {
		return "", 0, fmt.Errorf("%w: hotkey %q needs a modifier and a key", native.ErrInvalidArgument, s)
	}
	return key, mods, nil
}
