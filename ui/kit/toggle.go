package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ToggleView is a button that stays pressed, e.g. bold in a toolbar.
type ToggleView struct {
	text            string
	icon            *IconView
	value, disabled bool
	onChange        func(bool)
}

func Toggle(text string, on bool) *ToggleView            { return &ToggleView{text: text, value: on} }
func (v *ToggleView) Icon(name IconName) *ToggleView     { v.icon = Icon(name); return v }
func (v *ToggleView) OnChange(fn func(bool)) *ToggleView { v.onChange = fn; return v }
func (v *ToggleView) Value() bool                        { return v.value }
func (v *ToggleView) SetValue(on bool)                   { v.value = on }
func (v *ToggleView) SetDisabled(on bool)                { v.disabled = on }

func (v *ToggleView) Render(cx *el.Context) el.Element {
	return toggleButton(cx, autoID("toggle", v), v.text, v.icon, v.value, v.disabled, func() {
		v.value = !v.value
		if v.onChange != nil {
			v.onChange(v.value)
		}
	})
}

func toggleButton(cx *el.Context, id, text string, icon *IconView, on, disabled bool, fn func()) *el.DivEl {
	bg, fg, border := theme.Surface, theme.Text, theme.Border
	if on {
		bg, fg, border = theme.Highlight, theme.PrimaryText, theme.Primary
	}
	if disabled {
		fg = theme.Muted
	}
	b := el.Div().ID(id).Role("toggle").Name(text).Selected(on).Disabled(disabled).
		Row().Items(el.Center).Gap(6).H(el.Dp(32)).Px(12).Rounded(6).Bg(bg).Border(1, border).TextColor(fg).TextSize(14).
		Focusable(true).OnClick(fn).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	if !disabled {
		b.CursorPointer().Hover(func(s *el.Style) {
			if !on {
				s.Bg(theme.SubtleHover)
			}
		})
	}
	if icon != nil {
		ic := *icon
		b.Child(ic.Size(16).Color(fg).Render(cx))
	}
	if text != "" {
		b.Child(el.Text(text))
	}
	return b
}
