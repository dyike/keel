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
	target           string // element whose KeyContext path resolves action
	plain            bool
	size             float32
	style            func(*el.TextEl)
}

// KbdFor displays the first chord bound to a keymap action (core.Bind), and
// follows rebinding; it shows nothing while the action is unbound.
func KbdFor(action string) *KbdView { return &KbdView{action: action} }

// At resolves a KbdFor action as it applies to an element, through that
// element's KeyContext path and core.BindIn predicates, as its own key
// handling would: the hint on a toolbar button shows the editor's binding.
// It shows nothing while the element is missing, hidden or disabled.
func (v *KbdView) At(elementID string) *KbdView { v.target = elementID; return v }

// Kbd displays a ParseShortcut chord, or the literal label if parsing fails.
func Kbd(shortcut string) *KbdView { return &KbdView{shortcut: shortcut} }

// Plain hides the keycap outline, preserving its spacing.
func (v *KbdView) Plain() *KbdView { v.plain = true; return v }

// Size sets the font size in sp and scales padding with it. Zero restores
// inherited text size and default padding. Values outside 0–128 are ignored.
func (v *KbdView) Size(sp float32) *KbdView {
	if sp >= 0 && sp <= 128 {
		v.size = sp
	}
	return v
}

// Style refines the newly built keycap after defaults; nil restores defaults.
// Do not retain the element. The shortcut remains its accessible name.
func (v *KbdView) Style(fn func(*el.TextEl)) *KbdView { v.style = fn; return v }
func (v *KbdView) Render(cx *el.Context) el.Element {
	if v.action != "" && v.target != "" {
		// Resolved after the frame's tree is built, when contexts are known.
		return el.KeyHint(v.action, v.target, func(chord string) el.Element { return v.keycap(chord) })
	}
	shortcut := v.shortcut
	if v.action != "" {
		b := core.Bindings(v.action)
		if len(b) == 0 {
			return el.Div()
		}
		shortcut = b[0]
	}
	return v.keycap(shortcut)
}

func (v *KbdView) keycap(shortcut string) el.Element {
	border := theme.Border
	if v.plain {
		border = color.NRGBA{}
	}
	cap := el.Text(core.ShortcutLabel(shortcut, runtime.GOOS)).Name(shortcut).
		TextColor(theme.Muted).Px(theme.SpaceSm).Py(3).Border(1, border).Rounded(theme.RadiusSm).MaxW(el.Full).MaxLines(1)
	if v.size > 0 {
		ratio := v.size / theme.TextBody
		cap.TextSize(v.size).Px(theme.SpaceSm * ratio).Py(3 * ratio)
	}
	if v.style != nil {
		v.style(cap)
	}
	return cap.Name(shortcut)
}
