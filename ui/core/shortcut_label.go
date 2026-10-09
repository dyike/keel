package core

import (
	"strings"

	"github.com/dyike/keel/third_party/gio/io/key"
)

// ShortcutLabel formats a ParseShortcut chord for goos (darwin, windows or linux).
// Invalid chords are returned unchanged. It does not register a shortcut.
func ShortcutLabel(s, goos string) string {
	// Resolve mod for the display platform before parsing (also testable on macOS).
	parts := strings.Split(s, "+")
	for i, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), "mod") {
			if goos == "darwin" {
				parts[i] = "cmd"
			} else {
				parts[i] = "ctrl"
			}
		}
	}
	name, mods, err := ParseShortcut(strings.Join(parts, "+"))
	if err != nil {
		return s
	}
	var out []string
	mac := goos == "darwin"
	for _, m := range []struct {
		mod           key.Modifiers
		plain, symbol string
	}{
		{key.ModCtrl, "Ctrl", "⌃"}, {key.ModAlt, "Alt", "⌥"}, {key.ModShift, "Shift", "⇧"}, {key.ModCommand, "Super", "⌘"},
	} {
		if mods&m.mod != 0 {
			label := m.plain
			if mac {
				label = m.symbol
			} else if goos == "windows" && m.mod == key.ModCommand {
				label = "Win"
			}
			out = append(out, label)
		}
	}
	label := string(name)
	switch name {
	case key.NameEscape:
		label = "Esc"
	case key.NameReturn:
		label = "Enter"
	case key.NameSpace:
		label = "Space"
	case key.NameTab:
		label = "Tab"
	case key.NameDeleteBackward:
		label = "Backspace"
	case key.NameDeleteForward:
		label = "Delete"
	}
	if mac {
		switch name {
		case key.NameReturn:
			label = "↵"
		case key.NameDeleteBackward:
			label = "⌫"
		case key.NameEscape:
			// Escape's symbol is absent from several CJK fallback fonts.
			label = "Esc"
		case key.NameTab:
			label = "⇥"
		}
	}
	out = append(out, label)
	sep := "+"
	if mac {
		sep = ""
	}
	return strings.Join(out, sep)
}
