package window

import (
	"github.com/dyike/keel/ui/core"
	"runtime"
	"strings"
)

func requestMenuEdit(action MenuAction) bool {
	w, ok := core.CurrentWindow().(*Window)
	if !ok || w.closed {
		return false
	}
	w.menuEdits = append(w.menuEdits, core.EditAction(action))
	return true
}

// TakeEditAction implements the focused-editor bridge used by core.NextEditAction.
// Call from UI code; editors normally use core.NextEditAction instead.
func (w *Window) TakeEditAction() (core.EditAction, bool) {
	if len(w.menuEdits) == 0 {
		return "", false
	}
	action := w.menuEdits[0]
	w.menuEdits = w.menuEdits[1:]
	return action, true
}
func menuWireShortcut(item menuWireItem) string {
	parts := []string{}
	command := "Win"
	if runtime.GOOS == "darwin" {
		command = "Cmd"
	}
	for _, m := range []struct {
		bit  uint32
		name string
	}{{2, "Ctrl"}, {1, command}, {4, "Alt"}, {8, "Shift"}} {
		if item.Modifiers&m.bit != 0 {
			parts = append(parts, m.name)
		}
	}
	name := strings.ToUpper(item.Key)
	switch item.Key {
	case "⏎", "⌤":
		name = "Enter"
	case "⎋":
		name = "Esc"
	case "tab":
		name = "Tab"
	case "space":
		name = "Space"
	case "⌫":
		name = "Backspace"
	case "⌦":
		name = "Delete"
	}
	return strings.Join(append(parts, name), "+")
}
