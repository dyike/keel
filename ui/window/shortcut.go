package window

import (
	"fmt"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/core"
)

type shortcut struct {
	name key.Name
	mods key.Modifiers
	fn   func()
}

func (w *Window) handleShortcuts(gtx core.C) {
	for _, s := range append(w.shortcuts, applicationMenuShortcuts()...) {
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
	name, mods, err := core.ParseShortcut(s)
	if err != nil {
		return shortcut{}, fmt.Errorf("window: %w", err)
	}
	return shortcut{name: name, mods: mods}, nil
}
