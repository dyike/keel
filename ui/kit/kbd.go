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
	shortcut string
	plain    bool
}

// Kbd displays a ParseShortcut chord, or the literal label if parsing fails.
func Kbd(shortcut string) *KbdView { return &KbdView{shortcut: shortcut} }

// Plain hides the keycap outline, preserving its spacing.
func (v *KbdView) Plain() *KbdView { v.plain = true; return v }
func (v *KbdView) Render(cx *el.Context) el.Element {
	border := theme.Border
	if v.plain {
		border = color.NRGBA{}
	}
	return el.Text(core.ShortcutLabel(v.shortcut, runtime.GOOS)).Name(v.shortcut).
		TextColor(theme.Muted).Px(6).Py(3).Border(1, border).Rounded(theme.RadiusSm).MaxW(el.Full).MaxLines(1)
}
