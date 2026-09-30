package core

import (
	"fmt"
	"strings"

	"gioui.org/io/key"
)

// ParseShortcut reads a key chord such as "mod+s", "ctrl+shift+k" or "esc"
// into a Gio key name and modifiers. "mod" is Cmd on macOS and Ctrl elsewhere.
func ParseShortcut(s string) (key.Name, key.Modifiers, error) {
	var name key.Name
	var mods key.Modifiers
	for _, p := range strings.Split(strings.ToLower(s), "+") {
		switch p = strings.TrimSpace(p); p {
		case "mod":
			mods |= key.ModShortcut
		case "cmd", "command", "super":
			mods |= key.ModCommand
		case "ctrl", "control":
			mods |= key.ModCtrl
		case "shift":
			mods |= key.ModShift
		case "alt", "option":
			mods |= key.ModAlt
		default:
			if p == "" || name != "" {
				return "", 0, fmt.Errorf("invalid shortcut %q", s)
			}
			name = keyName(p)
		}
	}
	if name == "" {
		return "", 0, fmt.Errorf("shortcut %q has no key", s)
	}
	return name, mods, nil
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
