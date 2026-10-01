package widget

import (
	"image"
	"runtime"
	"strings"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// KbdView displays a shortcut. It does not register a keyboard handler.
type KbdView struct {
	shortcut string
	size     ComponentSize
	plain    bool
}

// Kbd uses core.ParseShortcut syntax (e.g. mod+shift+p). Invalid chords are
// displayed literally, allowing arbitrary key labels as well.
func Kbd(shortcut string) *KbdView                  { return &KbdView{shortcut: shortcut} }
func (k *KbdView) Size(size ComponentSize) *KbdView { k.size = size; return k }
func (k *KbdView) Plain() *KbdView                  { k.plain = true; return k }
func (k *KbdView) SetShortcut(shortcut string)      { k.shortcut = shortcut }
func (k *KbdView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	return core.Semantic(gtx, func(gtx C) D {
		content := func(gtx C) D {
			return giolayout.Inset{Left: 6, Right: 6, Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
				size, _, _ := k.size.metrics()
				st := material.Label(theme.Material, size-2, shortcutLabel(k.shortcut, runtime.GOOS))
				st.Color = theme.Muted
				return layoutLabel(gtx, st)
			})
		}
		if k.plain {
			return content(gtx)
		}
		return widget.Border{Color: theme.Border, CornerRadius: 4, Width: 1}.Layout(gtx, content)
	}, semantic.LabelOp(k.shortcut))
}

func shortcutLabel(s, platform string) string {
	// Resolve mod for the display platform before parsing (also testable on macOS).
	parts := strings.Split(s, "+")
	for i, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), "mod") {
			if platform == "darwin" {
				parts[i] = "cmd"
			} else {
				parts[i] = "ctrl"
			}
		}
	}
	name, mods, err := core.ParseShortcut(strings.Join(parts, "+"))
	if err != nil {
		return s
	}
	var out []string
	mac := platform == "darwin"
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
			} else if platform == "windows" && m.mod == key.ModCommand {
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
