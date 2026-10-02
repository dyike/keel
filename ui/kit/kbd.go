package kit

import (
	"image/color"
	"runtime"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// KbdView displays a shortcut without registering a keyboard handler.
type KbdView struct {
	shortcut, action string
	plain            bool
}

// KbdFor displays the first chord bound to a keymap action (core.Bind), and
// follows rebinding; it shows nothing while the action is unbound.
func KbdFor(action string) *KbdView { return &KbdView{action: action} }

// Kbd displays a ParseShortcut chord, or the literal label if parsing fails.
func Kbd(shortcut string) *KbdView { return &KbdView{shortcut: shortcut} }

// Plain hides the keycap outline, preserving its spacing.
func (v *KbdView) Plain() *KbdView { v.plain = true; return v }
func (v *KbdView) Render(cx *el.Context) el.Element {
	shortcut := v.shortcut
	if v.action != "" {
		b := core.Bindings(v.action)
		if len(b) == 0 {
			return el.Div()
		}
		shortcut = b[0]
	}
	border := theme.Border
	if v.plain {
		border = color.NRGBA{}
	}
	return el.Text(core.ShortcutLabel(shortcut, runtime.GOOS)).Name(shortcut).
		TextColor(theme.Muted).Px(6).Py(3).Border(1, border).Rounded(theme.RadiusSm).MaxW(el.Full).MaxLines(1)
}
