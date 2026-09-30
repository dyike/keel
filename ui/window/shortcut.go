package window

import (
	"fmt"
	"strings"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/core"
)

type shortcut struct {
	name key.Name
	mods key.Modifiers
	fn   func()
}

func (w *Window) handleShortcuts(gtx core.C) {
	for _, s := range w.shortcuts {
		for {
			ev, ok := gtx.Event(key.Filter{Name: s.name, Required: s.mods})
			if !ok {
				break
			}
			if e, ok := ev.(key.Event); ok && e.State == key.Press {
				core.Call(gtx, s.fn)
			}
		}
	}
}

func mustParseShortcuts(m map[string]func()) []shortcut {
	var out []shortcut
	for accel, fn := range m {
		s, err := parseShortcut(accel)
		if err != nil {
			panic(err)
		}
		s.fn = fn
		out = append(out, s)
	}
	return out
}

func parseShortcut(s string) (shortcut, error) {
	var sc shortcut
	for _, p := range strings.Split(strings.ToLower(s), "+") {
		switch p = strings.TrimSpace(p); p {
		case "mod":
			sc.mods |= key.ModShortcut
		case "cmd", "command", "super":
			sc.mods |= key.ModCommand
		case "ctrl", "control":
			sc.mods |= key.ModCtrl
		case "shift":
			sc.mods |= key.ModShift
		case "alt", "option":
			sc.mods |= key.ModAlt
		default:
			if p == "" || sc.name != "" {
				return sc, fmt.Errorf("app: invalid shortcut %q", s)
			}
			sc.name = keyName(p)
		}
	}
	if sc.name == "" {
		return sc, fmt.Errorf("app: shortcut %q has no key", s)
	}
	return sc, nil
}

func keyName(p string) key.Name {
	switch p {
	case "esc", "escape":
		return key.NameEscape
	case "enter", "return":
		return key.NameReturn
	case "space":
		return key.NameSpace
	case "tab":
		return key.NameTab
	case "backspace":
		return key.NameDeleteBackward
	case "delete":
		return key.NameDeleteForward
	case "up":
		return key.NameUpArrow
	case "down":
		return key.NameDownArrow
	case "left":
		return key.NameLeftArrow
	case "right":
		return key.NameRightArrow
	}
	return key.Name(strings.ToUpper(p)) // letters and F-keys are upper case in Gio
}
